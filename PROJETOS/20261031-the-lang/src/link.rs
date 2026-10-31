//! Write an assembled image into ELF, Mach-O, or PE. No external linker.

use crate::amd64::Object;

#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Os {
    Linux,
    Darwin,
    Windows,
}

#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Cpu {
    Amd64,
    Arm64,
}

#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Container {
    Elf,
    MachO,
    Pe,
}

pub fn container(os: Os) -> Container {
    match os {
        Os::Linux => Container::Elf,
        Os::Darwin => Container::MachO,
        Os::Windows => Container::Pe,
    }
}

pub fn write(object: &Object, os: Os, cpu: Cpu) -> Result<Vec<u8>, String> {
    let vaddr0 = match os {
        Os::Linux => 0x400000 + 0x80,
        Os::Darwin => 0x100000000 + 0x4000,
        Os::Windows => 0x140000000 + 0x1000,
    };
    let blob = object.blob(vaddr0)?;
    match container(os) {
        Container::Elf => Ok(elf(&blob, object.bss, object.entry, vaddr0, cpu)),
        Container::MachO => Ok(macho(&blob, object.bss, object.entry, vaddr0, cpu)),
        Container::Pe => Ok(pe(&blob, object.bss, object.entry, cpu)),
    }
}

fn elf(blob: &[u8], bss: u64, entry: u64, vaddr0: u64, cpu: Cpu) -> Vec<u8> {
    let base = 0x400000u64;
    let header = 0x80u64;
    debug_assert_eq!(vaddr0, base + header);
    let filesz = header + blob.len() as u64;
    let memsz = filesz + bss;
    let mut out = vec![0u8; header as usize];
    out[0..4].copy_from_slice(&[0x7F, b'E', b'L', b'F']);
    out[4] = 2;
    out[5] = 1;
    out[6] = 1;
    put16(&mut out, 16, 2);
    put16(&mut out, 18, if cpu == Cpu::Arm64 { 183 } else { 62 });
    put32(&mut out, 20, 1);
    put64(&mut out, 24, vaddr0 + entry);
    put64(&mut out, 32, 64);
    put16(&mut out, 52, 64);
    put16(&mut out, 54, 56);
    put16(&mut out, 56, 1);
    let ph = 64;
    put32(&mut out, ph, 1);
    put32(&mut out, ph + 4, 7);
    put64(&mut out, ph + 8, 0);
    put64(&mut out, ph + 16, base);
    put64(&mut out, ph + 24, base);
    put64(&mut out, ph + 32, filesz);
    put64(&mut out, ph + 40, memsz);
    put64(&mut out, ph + 48, 0x1000);
    out.extend(blob);
    out
}

fn macho(blob: &[u8], bss: u64, entry: u64, vaddr0: u64, cpu: Cpu) -> Vec<u8> {
    let base = 0x100000000u64;
    let header = 0x4000u64;
    debug_assert_eq!(vaddr0, base + header);
    let mut cmds = Vec::new();
    segment(&mut cmds, "__TEXT", base, header + blob.len() as u64, 0, header + blob.len() as u64, 5);
    if bss > 0 {
        let bss_addr = vaddr0 + blob.len() as u64;
        segment(&mut cmds, "__DATA", bss_addr, bss, header + blob.len() as u64, 0, 3);
    }
    let mut main = vec![0u8; 24];
    put32(&mut main, 0, 0x80000028);
    put32(&mut main, 4, 24);
    put64(&mut main, 8, header + entry);
    put64(&mut main, 16, 0);
    cmds.extend(main);
    dylib(&mut cmds, "/usr/lib/libSystem.B.dylib");
    let mut out = vec![0u8; 32];
    put32(&mut out, 0, 0xFEEDFACF);
    put32(&mut out, 4, if cpu == Cpu::Arm64 { 0x0100000C } else { 0x01000007 });
    put32(&mut out, 8, 0);
    put32(&mut out, 12, 2);
    put32(&mut out, 16, 2 + u32::from(bss > 0) + 1);
    put32(&mut out, 20, cmds.len() as u32);
    put32(&mut out, 24, 0);
    out.extend(cmds);
    out.resize(header as usize, 0);
    out.extend(blob);
    out.extend(std::iter::repeat(0).take(bss as usize));
    out
}

fn segment(cmds: &mut Vec<u8>, name: &str, vmaddr: u64, vmsize: u64, fileoff: u64, filesize: u64, prot: u32) {
    let start = cmds.len();
    cmds.resize(start + 72, 0);
    put32(cmds, start, 0x19);
    put32(cmds, start + 4, 72);
    let mut raw = [0u8; 16];
    raw[..name.len()].copy_from_slice(name.as_bytes());
    cmds[start + 8..start + 24].copy_from_slice(&raw);
    put64(cmds, start + 24, vmaddr);
    put64(cmds, start + 32, vmsize);
    put64(cmds, start + 40, fileoff);
    put64(cmds, start + 48, filesize);
    put32(cmds, start + 56, prot);
    put32(cmds, start + 60, prot);
}

fn dylib(cmds: &mut Vec<u8>, path: &str) {
    let start = cmds.len();
    let size = (24 + path.len() + 1 + 7) & !7;
    cmds.resize(start + size, 0);
    put32(cmds, start, 0xC);
    put32(cmds, start + 4, size as u32);
    put32(cmds, start + 8, 24);
    cmds[start + 24..start + 24 + path.len()].copy_from_slice(path.as_bytes());
}

fn pe(blob: &[u8], bss: u64, entry: u64, cpu: Cpu) -> Vec<u8> {
    let file_align = 0x200u64;
    let sect_align = 0x1000u64;
    let header = file_align;
    let image = 0x140000000u64;
    let mut out = vec![0u8; header as usize];
    out[0] = b'M';
    out[1] = b'Z';
    put32(&mut out, 0x3C, 0x80);
    out[0x80..0x84].copy_from_slice(b"PE\0\0");
    let coff = 0x84;
    put16(&mut out, coff, if cpu == Cpu::Arm64 { 0xAA64 } else { 0x8664 });
    put16(&mut out, coff + 2, 1);
    put16(&mut out, coff + 16, 0xF0);
    put16(&mut out, coff + 18, 0x22);
    let opt = coff + 20;
    put16(&mut out, opt, 0x20B);
    out[opt + 2] = 1;
    put32(&mut out, opt + 16, (0x1000 + entry) as u32);
    put32(&mut out, opt + 20, 0x1000);
    put64(&mut out, opt + 24, image);
    put32(&mut out, opt + 32, 0x1000);
    put32(&mut out, opt + 36, 0x200);
    put16(&mut out, opt + 68, 0x60);
    put32(&mut out, opt + 56, (sect_align + blob.len() as u64 + bss).div_ceil(sect_align) as u32 * sect_align as u32);
    put32(&mut out, opt + 60, 3);
    let raw = file_align + blob.len() as u64;
    let sec = opt + 0xF0;
    out[sec..sec + 5].copy_from_slice(b".text");
    put32(&mut out, sec + 8, (blob.len() as u64 + bss) as u32);
    put32(&mut out, sec + 12, sect_align as u32);
    put32(&mut out, sec + 16, blob.len() as u32);
    put32(&mut out, sec + 20, file_align as u32);
    put32(&mut out, sec + 36, 0xE0000060);
    out.extend(blob);
    while out.len() < raw as usize {
        out.push(0);
    }
    let _ = image;
    out
}

fn put16(buf: &mut [u8], at: usize, v: u16) {
    buf[at..at + 2].copy_from_slice(&v.to_le_bytes());
}

fn put32(buf: &mut [u8], at: usize, v: u32) {
    buf[at..at + 4].copy_from_slice(&v.to_le_bytes());
}

fn put64(buf: &mut [u8], at: usize, v: u64) {
    buf[at..at + 8].copy_from_slice(&v.to_le_bytes());
}
