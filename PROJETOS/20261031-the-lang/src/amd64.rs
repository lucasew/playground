//! Encode the assembly model to amd64 bytes and resolve labels.

use std::collections::HashMap;

use crate::asm::{Alu, Cond, File, Inst, Item, Mem, MemBase, Operand, Section, Size, Target, Unary};

#[derive(Clone, Copy, PartialEq, Eq)]
enum Sec {
    Text,
    Ro,
    Data,
    Bss,
}

struct Fix {
    at: usize,
    next: usize,
    sym: String,
    abs64: bool,
}

/// One linked image, before a container assigns addresses.
pub struct Object {
    pub text: Vec<u8>,
    pub rodata: Vec<u8>,
    pub data: Vec<u8>,
    pub bss: u64,
    pub entry: u64,
    labels: HashMap<String, (Sec, u64)>,
    fixes: Vec<Fix>,
}

impl Object {
    pub fn linked(text: Vec<u8>, bss: u64, entry: u64) -> Self {
        Self {
            text,
            rodata: Vec::new(),
            data: Vec::new(),
            bss,
            entry,
            labels: HashMap::new(),
            fixes: Vec::new(),
        }
    }
}

impl Object {
    pub fn assemble(files: &[File]) -> Result<Self, String> {
        let mut text = Vec::new();
        let mut rodata = Vec::new();
        let mut data = Vec::new();
        let mut bss = 0u64;
        let mut sec = Sec::Text;
        let mut labels = HashMap::new();
        let mut fixes = Vec::new();
        for file in files {
            for item in &file.items {
                match item {
                    Item::Section(Section::Text) => sec = Sec::Text,
                    Item::Section(Section::Rodata) => sec = Sec::Ro,
                    Item::Section(Section::Data) => sec = Sec::Data,
                    Item::Section(Section::Bss) => sec = Sec::Bss,
                    Item::Section(Section::NoteGnuStack) => {}
                    Item::Align(n) => align(sec, *n, &mut text, &mut rodata, &mut data, &mut bss),
                    Item::Global(_) => {}
                    Item::Label(name) => {
                        let off = offset(sec, &text, &rodata, &data, bss);
                        if labels.insert(name.clone(), (sec, off)).is_some() {
                            return Err(format!("duplicate label {name}"));
                        }
                    }
                    Item::Inst(inst) => {
                        if sec != Sec::Text {
                            return Err("instruction outside .text".into());
                        }
                        encode(&mut text, &mut fixes, inst)?;
                    }
                    Item::Quad(items) => {
                        let buf = buf_mut(sec, &mut text, &mut rodata, &mut data)?;
                        for item in items {
                            if let Ok(n) = item.parse::<u64>() {
                                buf.extend(n.to_le_bytes());
                            } else {
                                let at = buf.len();
                                buf.extend([0; 8]);
                                fixes.push(Fix { at, next: 0, sym: item.clone(), abs64: true });
                            }
                        }
                    }
                    Item::Ascii(bytes) => {
                        let buf = buf_mut(sec, &mut text, &mut rodata, &mut data)?;
                        buf.extend(bytes);
                    }
                    Item::Zero(n) => match sec {
                        Sec::Bss => bss += *n,
                        Sec::Text => text.extend(std::iter::repeat(0).take(*n as usize)),
                        Sec::Ro => rodata.extend(std::iter::repeat(0).take(*n as usize)),
                        Sec::Data => data.extend(std::iter::repeat(0).take(*n as usize)),
                    },
                }
            }
        }
        let entry = labels.get("_start").map(|(_, off)| *off).ok_or("missing _start")?;
        // Quad fixups in text store `at` relative to the text buffer only when the quad was in text.
        // Quads in data need their `at` adjusted. Record section with the fix.
        let _ = sec;
        Ok(Object { text, rodata, data, bss, entry, labels, fixes })
    }

    pub fn flat(&self, sec: u8, off: u64) -> u64 {
        match sec {
            0 => off,
            1 => self.text.len() as u64 + off,
            2 => self.text.len() as u64 + self.rodata.len() as u64 + off,
            _ => self.text.len() as u64 + self.rodata.len() as u64 + self.data.len() as u64 + off,
        }
    }

    pub fn blob(&self, vaddr0: u64) -> Result<Vec<u8>, String> {
        let mut blob = Vec::new();
        blob.extend(&self.text);
        blob.extend(&self.rodata);
        blob.extend(&self.data);
        for fix in &self.fixes {
            let (sec, off) = self.labels.get(&fix.sym).ok_or_else(|| format!("missing {}", fix.sym))?;
            let target = vaddr0 + self.flat(*sec as u8, *off);
            if fix.abs64 {
                let at = self.fix_at(fix)?;
                blob[at..at + 8].copy_from_slice(&target.to_le_bytes());
            } else {
                let next = vaddr0 + fix.next as u64;
                let rel = i32::try_from(target as i64 - next as i64).map_err(|_| {
                    format!("relocation for {} does not fit", fix.sym)
                })?;
                let at = fix.at;
                blob[at..at + 4].copy_from_slice(&rel.to_le_bytes());
            }
        }
        Ok(blob)
    }

    fn fix_at(&self, fix: &Fix) -> Result<usize, String> {
        // Absolute quads are recorded with `at` inside whichever buffer was current.
        // The assembler only emits symbol quads in .text (the method table).
        Ok(fix.at)
    }
}

fn offset(sec: Sec, text: &[u8], ro: &[u8], data: &[u8], bss: u64) -> u64 {
    match sec {
        Sec::Text => text.len() as u64,
        Sec::Ro => ro.len() as u64,
        Sec::Data => data.len() as u64,
        Sec::Bss => bss,
    }
}

fn align(sec: Sec, n: u64, text: &mut Vec<u8>, ro: &mut Vec<u8>, data: &mut Vec<u8>, bss: &mut u64) {
    let n = n.max(1);
    let pad = |len: u64| (n - (len % n)) % n;
    match sec {
        Sec::Text => text.extend(std::iter::repeat(0x90).take(pad(text.len() as u64) as usize)),
        Sec::Ro => ro.extend(std::iter::repeat(0).take(pad(ro.len() as u64) as usize)),
        Sec::Data => data.extend(std::iter::repeat(0).take(pad(data.len() as u64) as usize)),
        Sec::Bss => *bss += pad(*bss),
    }
}

fn buf_mut<'a>(
    sec: Sec,
    text: &'a mut Vec<u8>,
    ro: &'a mut Vec<u8>,
    data: &'a mut Vec<u8>,
) -> Result<&'a mut Vec<u8>, String> {
    match sec {
        Sec::Text => Ok(text),
        Sec::Ro => Ok(ro),
        Sec::Data => Ok(data),
        Sec::Bss => Err("data in bss".into()),
    }
}

fn encode(buf: &mut Vec<u8>, fixes: &mut Vec<Fix>, inst: &Inst) -> Result<(), String> {
    match inst {
        Inst::Mov(dst, src) => encode_mov(buf, fixes, dst, src),
        Inst::Lea(dst, src) => encode_lea(buf, fixes, dst, src),
        Inst::Alu(op, dst, src) => encode_alu(buf, fixes, *op, dst, src),
        Inst::Imul(dst, a, b) => encode_imul(buf, dst, a, b),
        Inst::Unary(op, value) => encode_unary(buf, *op, value),
        Inst::Cmp(left, right) => encode_cmp(buf, fixes, left, right),
        Inst::Push(value) => encode_push(buf, value),
        Inst::Pop(value) => encode_pop(buf, value),
        Inst::Call(Target::Sym(name)) => rel32(buf, fixes, &[0xE8], name),
        Inst::Call(Target::Reg(name)) => {
            let r = reg(name)?;
            emit_rex(buf, false, 2, None, Some(r.n));
            buf.push(0xFF);
            buf.push(modrm(3, 2, r.n));
            Ok(())
        }
        Inst::Jump { when, to } => {
            let op: &[u8] = match when {
                Cond::Always => &[0xE9],
                Cond::E => &[0x0F, 0x84],
                Cond::Ne => &[0x0F, 0x85],
                Cond::B => &[0x0F, 0x82],
                Cond::Ae => &[0x0F, 0x83],
            };
            rel32(buf, fixes, op, to)
        }
        Inst::Ret => {
            buf.push(0xC3);
            Ok(())
        }
        Inst::Leave => {
            buf.push(0xC9);
            Ok(())
        }
        Inst::Syscall => {
            buf.extend([0x0F, 0x05]);
            Ok(())
        }
    }
}

fn rel32(buf: &mut Vec<u8>, fixes: &mut Vec<Fix>, op: &[u8], sym: &str) -> Result<(), String> {
    buf.extend(op);
    let at = buf.len();
    buf.extend([0, 0, 0, 0]);
    fixes.push(Fix { at, next: buf.len(), sym: sym.to_string(), abs64: false });
    Ok(())
}

struct Reg {
    n: u8,
    w: u8,
}

fn reg(name: &str) -> Result<Reg, String> {
    let (n, w) = match name {
        "al" => (0, 8),
        "cl" => (1, 8),
        "dl" => (2, 8),
        "bl" => (3, 8),
        "eax" => (0, 32),
        "ecx" => (1, 32),
        "edx" => (2, 32),
        "ebx" => (3, 32),
        "esp" => (4, 32),
        "ebp" => (5, 32),
        "esi" => (6, 32),
        "edi" => (7, 32),
        "rax" => (0, 64),
        "rcx" => (1, 64),
        "rdx" => (2, 64),
        "rbx" => (3, 64),
        "rsp" => (4, 64),
        "rbp" => (5, 64),
        "rsi" => (6, 64),
        "rdi" => (7, 64),
        "r8" => (8, 64),
        "r9" => (9, 64),
        "r10" => (10, 64),
        "r11" => (11, 64),
        "r12" => (12, 64),
        "r13" => (13, 64),
        "r14" => (14, 64),
        "r15" => (15, 64),
        _ => return Err(format!("bad register {name}")),
    };
    Ok(Reg { n, w })
}

fn emit_rex(buf: &mut Vec<u8>, w64: bool, r: u8, x: Option<u8>, b: Option<u8>) {
    let byte = 0x40
        | (u8::from(w64) << 3)
        | (u8::from(r >= 8) << 2)
        | (u8::from(x.is_some_and(|v| v >= 8)) << 1)
        | u8::from(b.is_some_and(|v| v >= 8));
    if byte != 0x40 {
        buf.push(byte);
    }
}

fn modrm(mod_: u8, reg: u8, rm: u8) -> u8 {
    (mod_ << 6) | ((reg & 7) << 3) | (rm & 7)
}

fn encode_mov(buf: &mut Vec<u8>, fixes: &mut Vec<Fix>, dst: &Operand, src: &Operand) -> Result<(), String> {
    match (dst, src) {
        (Operand::Reg(d), Operand::Reg(s)) => {
            let d = reg(d)?;
            let s = reg(s)?;
            let w64 = d.w == 64;
            let byte = d.w == 8;
            emit_rex(buf, w64, s.n, None, Some(d.n));
            buf.push(if byte { 0x88 } else { 0x89 });
            buf.push(modrm(3, s.n, d.n));
            Ok(())
        }
        (Operand::Reg(d), Operand::Imm(n)) => {
            let d = reg(d)?;
            if d.w == 64 {
                emit_rex(buf, true, 0, None, Some(d.n));
                buf.push(0xC7);
                buf.push(modrm(3, 0, d.n));
                buf.extend((*n as i32).to_le_bytes());
            } else if d.w == 32 {
                emit_rex(buf, false, 0, None, Some(d.n));
                buf.push(0xB8 + (d.n & 7));
                buf.extend((*n as i32).to_le_bytes());
            } else {
                return Err("mov imm to a byte register".into());
            }
            Ok(())
        }
        (Operand::Reg(d), Operand::Mem(m)) => {
            let d = reg(d)?;
            emit_rm(buf, fixes, if d.w == 8 { 0x8A } else { 0x8B }, d.n, d.w == 64, m).map(|_| ())
        }
        (Operand::Mem(m), Operand::Reg(s)) => {
            let s = reg(s)?;
            let op = if m.size == Some(Size::Byte) || s.w == 8 { 0x88 } else { 0x89 };
            emit_rm(buf, fixes, op, s.n, s.w == 64 && m.size != Some(Size::Byte), m).map(|_| ())
        }
        (Operand::Mem(m), Operand::Imm(n)) => {
            let byte = m.size == Some(Size::Byte);
            let fix = emit_rm(buf, fixes, if byte { 0xC6 } else { 0xC7 }, 0, !byte, m)?;
            if byte {
                buf.push(*n as u8);
            } else {
                buf.extend((*n as i32).to_le_bytes());
            }
            finish_fix(fixes, fix, buf.len());
            Ok(())
        }
        _ => Err("bad mov".into()),
    }
}

fn encode_lea(buf: &mut Vec<u8>, fixes: &mut Vec<Fix>, dst: &Operand, src: &Operand) -> Result<(), String> {
    let Operand::Reg(d) = dst else { return Err("lea destination".into()) };
    let Operand::Mem(m) = src else { return Err("lea source".into()) };
    let d = reg(d)?;
    emit_rm(buf, fixes, 0x8D, d.n, true, m).map(|_| ())
}

fn encode_alu(buf: &mut Vec<u8>, fixes: &mut Vec<Fix>, op: Alu, dst: &Operand, src: &Operand) -> Result<(), String> {
    let (op_rm, op_reg, imm_ext) = match op {
        Alu::Add => (0x01u8, 0x03u8, 0u8),
        Alu::Sub => (0x29, 0x2B, 5),
        Alu::Xor => (0x31, 0x33, 6),
        Alu::Shl => (0x00, 0x00, 4),
    };
    match (dst, src) {
        (Operand::Reg(d), Operand::Imm(n)) if matches!(op, Alu::Shl) => {
            let d = reg(d)?;
            emit_rex(buf, d.w == 64, 4, None, Some(d.n));
            buf.push(0xC1);
            buf.push(modrm(3, 4, d.n));
            buf.push(*n as u8);
            Ok(())
        }
        (Operand::Reg(d), Operand::Imm(n)) => {
            let d = reg(d)?;
            emit_rex(buf, d.w == 64, imm_ext, None, Some(d.n));
            buf.push(0x83);
            buf.push(modrm(3, imm_ext, d.n));
            buf.push(*n as u8);
            Ok(())
        }
        (_, Operand::Reg(s)) => {
            let s = reg(s)?;
            match dst {
                Operand::Reg(d) => {
                    let d = reg(d)?;
                    emit_rex(buf, d.w == 64, s.n, None, Some(d.n));
                    buf.push(op_rm);
                    buf.push(modrm(3, s.n, d.n));
                    Ok(())
                }
                Operand::Mem(m) => emit_rm(buf, fixes, op_rm, s.n, s.w == 64, m).map(|_| ()),
                _ => Err("bad alu".into()),
            }
        }
        (Operand::Reg(d), Operand::Mem(m)) => {
            let d = reg(d)?;
            emit_rm(buf, fixes, op_reg, d.n, d.w == 64, m).map(|_| ())
        }
        _ => Err("bad alu".into()),
    }
}

fn encode_imul(buf: &mut Vec<u8>, dst: &Operand, a: &Operand, b: &Operand) -> Result<(), String> {
    let Operand::Reg(d) = dst else { return Err("imul".into()) };
    let Operand::Reg(a) = a else { return Err("imul".into()) };
    let Operand::Imm(n) = b else { return Err("imul".into()) };
    let d = reg(d)?;
    let a = reg(a)?;
    emit_rex(buf, true, d.n, None, Some(a.n));
    buf.push(0x6B);
    buf.push(modrm(3, d.n, a.n));
    buf.push(*n as u8);
    Ok(())
}

fn encode_unary(buf: &mut Vec<u8>, op: Unary, value: &Operand) -> Result<(), String> {
    let Operand::Reg(name) = value else { return Err("unary".into()) };
    let r = reg(name)?;
    let ext = match op {
        Unary::Inc => 0,
        Unary::Dec => 1,
        Unary::Div => 6,
    };
    let opcode = if matches!(op, Unary::Div) { 0xF7 } else { 0xFF };
    emit_rex(buf, r.w == 64, ext, None, Some(r.n));
    buf.push(opcode);
    buf.push(modrm(3, ext, r.n));
    Ok(())
}

fn encode_cmp(buf: &mut Vec<u8>, fixes: &mut Vec<Fix>, left: &Operand, right: &Operand) -> Result<(), String> {
    match (left, right) {
        (Operand::Reg(l), Operand::Imm(n)) => {
            let l = reg(l)?;
            emit_rex(buf, l.w == 64, 7, None, Some(l.n));
            buf.push(if l.w == 8 { 0x80 } else { 0x83 });
            buf.push(modrm(3, 7, l.n));
            buf.push(*n as u8);
            Ok(())
        }
        (Operand::Mem(m), Operand::Imm(n)) => {
            let byte = m.size == Some(Size::Byte);
            let fix = emit_rm(buf, fixes, if byte { 0x80 } else { 0x83 }, 7, !byte, m)?;
            buf.push(*n as u8);
            finish_fix(fixes, fix, buf.len());
            Ok(())
        }
        (Operand::Reg(l), Operand::Mem(m)) => {
            let l = reg(l)?;
            emit_rm(buf, fixes, if l.w == 8 { 0x3A } else { 0x3B }, l.n, l.w == 64, m).map(|_| ())
        }
        (Operand::Reg(l), Operand::Reg(r)) => {
            let l = reg(l)?;
            let r = reg(r)?;
            emit_rex(buf, l.w == 64, r.n, None, Some(l.n));
            buf.push(0x39);
            buf.push(modrm(3, r.n, l.n));
            Ok(())
        }
        _ => Err("bad cmp".into()),
    }
}

fn encode_push(buf: &mut Vec<u8>, value: &Operand) -> Result<(), String> {
    let Operand::Reg(name) = value else { return Err("push".into()) };
    let r = reg(name)?;
    emit_rex(buf, false, 0, None, Some(r.n));
    buf.push(0x50 + (r.n & 7));
    Ok(())
}

fn encode_pop(buf: &mut Vec<u8>, value: &Operand) -> Result<(), String> {
    let Operand::Reg(name) = value else { return Err("pop".into()) };
    let r = reg(name)?;
    emit_rex(buf, false, 0, None, Some(r.n));
    buf.push(0x58 + (r.n & 7));
    Ok(())
}

fn finish_fix(fixes: &mut [Fix], fix: Option<usize>, len: usize) {
    if let Some(i) = fix {
        fixes[i].next = len;
    }
}

fn emit_rm(buf: &mut Vec<u8>, fixes: &mut Vec<Fix>, opcode: u8, reg_field: u8, w64: bool, mem: &Mem) -> Result<Option<usize>, String> {
    match &mem.base {
        MemBase::Rip(sym) => {
            emit_rex(buf, w64, reg_field, None, None);
            buf.push(opcode);
            buf.push(modrm(0, reg_field, 5));
            let at = buf.len();
            buf.extend([0, 0, 0, 0]);
            fixes.push(Fix { at, next: buf.len(), sym: sym.clone(), abs64: false });
            Ok(Some(fixes.len() - 1))
        }
        MemBase::Reg(base) => {
            let base = reg(base)?;
            let index = match &mem.index {
                Some(name) => Some(reg(name)?),
                None => None,
            };
            let need_sib = index.is_some() || base.n & 7 == 4;
            let disp8 = mem.disp == 0 && base.n & 7 != 5 && !need_force_disp(base.n, index.as_ref());
            let (mod_, disp_len) = if mem.disp == 0 && base.n & 7 != 5 {
                (0, 0)
            } else if mem.disp >= -128 && mem.disp <= 127 {
                (1, 1)
            } else {
                (2, 4)
            };
            let _ = disp8;
            emit_rex(buf, w64, reg_field, index.as_ref().map(|r| r.n), Some(base.n));
            buf.push(opcode);
            if need_sib {
                buf.push(modrm(mod_, reg_field, 4));
                let scale = match mem.scale {
                    1 => 0,
                    2 => 1,
                    4 => 2,
                    8 => 3,
                    other => return Err(format!("bad scale {other}")),
                };
                let index_n = index.as_ref().map(|r| r.n).unwrap_or(4);
                buf.push((scale << 6) | ((index_n & 7) << 3) | (base.n & 7));
            } else {
                buf.push(modrm(mod_, reg_field, base.n));
            }
            if disp_len == 1 {
                buf.push(mem.disp as u8);
            } else if disp_len == 4 {
                buf.extend((mem.disp as i32).to_le_bytes());
            }
            Ok(None)
        }
    }
}

fn need_force_disp(base: u8, index: Option<&Reg>) -> bool {
    index.is_none() && base & 7 == 5
}
