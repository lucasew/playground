//! Encode arm64 instructions. Branches and `adr` are PC-relative, so the image does not care which container it lands in.

use std::collections::HashMap;

use crate::amd64::Object;

struct Fix {
    at: usize,
    sym: String,
    kind: Kind,
}

enum Kind {
    Branch,
    Cond(u8),
    Adr(u8),
}

pub struct Buf {
    text: Vec<u8>,
    labels: HashMap<String, usize>,
    fixes: Vec<Fix>,
}

impl Buf {
    pub fn new() -> Self {
        Self { text: Vec::new(), labels: HashMap::new(), fixes: Vec::new() }
    }

    pub fn label(&mut self, name: &str) {
        self.labels.insert(name.to_string(), self.text.len());
    }

    pub fn word(&mut self, word: u32) {
        self.text.extend(word.to_le_bytes());
    }

    pub fn ret(&mut self) {
        self.word(0xD65F03C0);
    }

    pub fn svc(&mut self) {
        self.word(0xD4000001);
    }

    pub fn movz(&mut self, rd: u8, imm: u16) {
        self.word(0xD2800000 | (u32::from(imm) << 5) | u32::from(rd));
    }

    pub fn mov(&mut self, rd: u8, rm: u8) {
        self.word(0xAA0003E0 | (u32::from(rm) << 16) | u32::from(rd));
    }

    pub fn add_imm(&mut self, rd: u8, rn: u8, imm: u16) {
        self.word(0x91000000 | (u32::from(imm) << 10) | (u32::from(rn) << 5) | u32::from(rd));
    }

    pub fn sub_imm(&mut self, rd: u8, rn: u8, imm: u16) {
        self.word(0xD1000000 | (u32::from(imm) << 10) | (u32::from(rn) << 5) | u32::from(rd));
    }

    pub fn add(&mut self, rd: u8, rn: u8, rm: u8) {
        self.word(0x8B000000 | (u32::from(rm) << 16) | (u32::from(rn) << 5) | u32::from(rd));
    }

    pub fn add_lsl(&mut self, rd: u8, rn: u8, rm: u8, shift: u8) {
        self.word(0x8B000000 | (u32::from(shift) << 10) | (u32::from(rm) << 16) | (u32::from(rn) << 5) | u32::from(rd));
    }

    pub fn sub(&mut self, rd: u8, rn: u8, rm: u8) {
        self.word(0xCB000000 | (u32::from(rm) << 16) | (u32::from(rn) << 5) | u32::from(rd));
    }

    pub fn stp_pre(&mut self, rt: u8, rt2: u8, rn: u8, byte_off: i16) {
        let imm7 = (byte_off / 8) as u32 & 0x7F;
        self.word(0xA9800000 | (imm7 << 15) | (u32::from(rt2) << 10) | (u32::from(rn) << 5) | u32::from(rt));
    }

    pub fn ldp_post(&mut self, rt: u8, rt2: u8, rn: u8, byte_off: i16) {
        let imm7 = (byte_off / 8) as u32 & 0x7F;
        self.word(0xA8C00000 | (imm7 << 15) | (u32::from(rt2) << 10) | (u32::from(rn) << 5) | u32::from(rt));
    }

    pub fn br(&mut self, rn: u8) {
        self.word(0xD61F0000 | (u32::from(rn) << 5));
    }

    pub fn lsl(&mut self, rd: u8, rn: u8, shift: u8) {
        let immr = (64 - u32::from(shift)) & 63;
        let imms = 63 - u32::from(shift);
        self.word(0xD3400000 | (1 << 22) | (immr << 16) | (imms << 10) | (u32::from(rn) << 5) | u32::from(rd));
    }

    pub fn cmp_imm(&mut self, rn: u8, imm: u16) {
        self.word(0xF1000000 | (u32::from(imm) << 10) | (u32::from(rn) << 5) | 31);
    }

    pub fn cmp(&mut self, rn: u8, rm: u8) {
        self.word(0xEB00001F | (u32::from(rm) << 16) | (u32::from(rn) << 5));
    }

    pub fn ldr(&mut self, rt: u8, rn: u8, byte_off: u16) {
        self.word(0xF9400000 | ((u32::from(byte_off) / 8) << 10) | (u32::from(rn) << 5) | u32::from(rt));
    }

    pub fn str(&mut self, rt: u8, rn: u8, byte_off: u16) {
        self.word(0xF9000000 | ((u32::from(byte_off) / 8) << 10) | (u32::from(rn) << 5) | u32::from(rt));
    }

    pub fn strb_zr(&mut self, rn: u8) {
        self.word(0x39000000 | (u32::from(rn) << 5) | 31);
    }

    pub fn strb(&mut self, rt: u8, rn: u8) {
        self.word(0x39000000 | (u32::from(rn) << 5) | u32::from(rt));
    }

    pub fn udiv(&mut self, rd: u8, rn: u8, rm: u8) {
        self.word(0x9AC00800 | (u32::from(rm) << 16) | (u32::from(rn) << 5) | u32::from(rd));
    }

    pub fn msub(&mut self, rd: u8, rn: u8, rm: u8, ra: u8) {
        self.word(0x9B008000 | (u32::from(rm) << 16) | (u32::from(ra) << 10) | (u32::from(rn) << 5) | u32::from(rd));
    }

    pub fn ldrb(&mut self, rt: u8, rn: u8) {
        self.word(0x39400000 | (u32::from(rn) << 5) | u32::from(rt));
    }

    pub fn bl(&mut self, name: &str) {
        self.branch(name, Kind::Branch, true);
    }

    pub fn b(&mut self, name: &str) {
        self.branch(name, Kind::Branch, false);
    }

    pub fn bcond(&mut self, cond: u8, name: &str) {
        let at = self.text.len();
        self.word(0x54000000 | u32::from(cond));
        self.fixes.push(Fix { at, sym: name.to_string(), kind: Kind::Cond(cond) });
    }

    pub fn adr(&mut self, rd: u8, name: &str) {
        let at = self.text.len();
        self.word(0x10000000 | u32::from(rd));
        self.fixes.push(Fix { at, sym: name.to_string(), kind: Kind::Adr(rd) });
    }

    fn branch(&mut self, name: &str, kind: Kind, link: bool) {
        let at = self.text.len();
        self.word(if link { 0x94000000 } else { 0x14000000 });
        self.fixes.push(Fix { at, sym: name.to_string(), kind });
    }

    pub fn finish(self, entry: &str, bss_labels: &[(&str, u64)], bss: u64) -> Result<Object, String> {
        let text_len = self.text.len() as u64;
        let mut labels = self.labels;
        let mut cursor = 0u64;
        for (name, size) in bss_labels {
            labels.insert((*name).to_string(), (text_len + cursor) as usize);
            cursor += *size;
        }
        let mut text = self.text;
        for fix in &self.fixes {
            let target = *labels.get(&fix.sym).ok_or_else(|| format!("missing {}", fix.sym))? as i64;
            let here = fix.at as i64;
            let delta = target - here;
            let word = match &fix.kind {
                Kind::Branch => {
                    let imm = (delta / 4) as u32 & 0x03FF_FFFF;
                    let base = u32::from_le_bytes(text[fix.at..fix.at + 4].try_into().unwrap()) & 0xFC00_0000;
                    base | imm
                }
                Kind::Cond(cond) => {
                    let imm = ((delta / 4) as u32) & 0x7FFFF;
                    0x54000000 | (imm << 5) | u32::from(*cond)
                }
                Kind::Adr(rd) => {
                    let imm = delta as u32;
                    let immlo = imm & 0x3;
                    let immhi = (imm >> 2) & 0x7FFFF;
                    0x10000000 | (immlo << 29) | (immhi << 5) | u32::from(*rd)
                }
            };
            text[fix.at..fix.at + 4].copy_from_slice(&word.to_le_bytes());
        }
        let entry = *labels.get(entry).ok_or("missing entry")? as u64;
        Ok(Object::linked(text, bss, entry))
    }
}

pub struct Method {
    pub name: String,
    pub steps: Vec<Step>,
}

pub enum Step {
    Load(u16),
    Add(u16),
    Store(u16),
}

pub enum Main {
    Actor(u16),
    Await(u16, u16),
    Write(Vec<u8>),
    AwaitI64(u16, u16),
}

pub fn assemble(methods: &[Method], main: &[Main]) -> Result<Object, String> {
    let mut b = Buf::new();
    b.label("_start");
    b.bl("rt_start");
    b.mov(0, 0);
    b.bl("hal_exit");

    b.label("hal_write");
    b.mov(2, 1);
    b.mov(1, 0);
    b.movz(0, 2);
    b.movz(8, 64);
    b.svc();
    b.ret();

    b.label("hal_exit");
    b.movz(8, 93);
    b.svc();

    b.label("rt_start");
    b.stp_pre(29, 30, 31, -16);
    b.bl("lang_main");
    b.movz(0, 0);
    b.ldp_post(29, 30, 31, 16);
    b.ret();

    b.label("rt_actor");
    b.adr(1, "next_actor");
    b.ldr(0, 1, 0);
    b.add_imm(2, 0, 1);
    b.str(2, 1, 0);
    b.adr(3, "actors");
    b.lsl(2, 0, 6);
    b.add(3, 3, 2);
    b.movz(5, 8);
    b.label("zero");
    b.str(31, 3, 0);
    b.add_imm(3, 3, 8);
    b.sub_imm(5, 5, 1);
    b.cmp_imm(5, 0);
    b.bcond(1, "zero");
    b.ret();

    b.label("rt_load");
    b.adr(2, "actors");
    b.lsl(3, 0, 6);
    b.add(2, 2, 3);
    b.lsl(3, 1, 3);
    b.add(2, 2, 3);
    b.ldr(0, 2, 0);
    b.ret();

    b.label("rt_store");
    b.adr(3, "actors");
    b.lsl(4, 0, 6);
    b.add(3, 3, 4);
    b.lsl(4, 1, 3);
    b.add(3, 3, 4);
    b.str(2, 3, 0);
    b.ret();

    b.label("rt_call");
    b.adr(2, "next_promise");
    b.ldr(3, 2, 0);
    b.add_imm(4, 3, 1);
    b.str(4, 2, 0);
    b.adr(5, "mbox_w");
    b.ldr(6, 5, 0);
    b.adr(7, "mbox");
    b.add_lsl(8, 6, 6, 1);
    b.lsl(8, 8, 3);
    b.add(7, 7, 8);
    b.str(0, 7, 0);
    b.str(1, 7, 8);
    b.str(3, 7, 16);
    b.add_imm(6, 6, 1);
    b.str(6, 5, 0);
    b.adr(8, "prom_ready");
    b.add(8, 8, 3);
    b.strb_zr(8);
    b.mov(0, 3);
    b.ret();

    b.label("rt_await");
    b.stp_pre(29, 30, 31, -16);
    b.stp_pre(19, 20, 31, -16);
    b.stp_pre(21, 22, 31, -16);
    b.mov(19, 0);
    b.label("await_loop");
    b.adr(1, "prom_ready");
    b.add(1, 1, 19);
    b.ldrb(0, 1);
    b.cmp_imm(0, 0);
    b.bcond(1, "await_ready");
    b.adr(1, "mbox_r");
    b.ldr(2, 1, 0);
    b.adr(3, "mbox_w");
    b.ldr(4, 3, 0);
    b.cmp(2, 4);
    b.bcond(2, "await_dead");
    b.adr(5, "mbox");
    b.add_lsl(6, 2, 2, 1);
    b.lsl(6, 6, 3);
    b.add(5, 5, 6);
    b.ldr(20, 5, 0);
    b.ldr(21, 5, 8);
    b.ldr(22, 5, 16);
    b.add_imm(2, 2, 1);
    b.str(2, 1, 0);
    b.adr(7, "lang_methods");
    b.lsl(6, 21, 3);
    b.add(7, 7, 6);
    b.mov(0, 20);
    b.br(7);
    b.label("await_resume");
    b.adr(1, "prom_val");
    b.lsl(6, 22, 3);
    b.add(1, 1, 6);
    b.str(0, 1, 0);
    b.adr(1, "prom_ready");
    b.add(1, 1, 22);
    b.movz(3, 1);
    b.strb(3, 1);
    b.b("await_loop");
    b.label("await_ready");
    b.adr(1, "prom_val");
    b.lsl(2, 19, 3);
    b.add(1, 1, 2);
    b.ldr(0, 1, 0);
    b.ldp_post(21, 22, 31, 16);
    b.ldp_post(19, 20, 31, 16);
    b.ldp_post(29, 30, 31, 16);
    b.ret();
    b.label("await_dead");
    b.movz(0, 2);
    b.bl("hal_exit");

    b.label("rt_write");
    b.b("hal_write");

    b.label("rt_write_i64");
    b.stp_pre(29, 30, 31, -16);
    b.sub_imm(31, 31, 48);
    b.add_imm(5, 31, 40);
    b.mov(1, 5);
    b.movz(2, 10);
    b.cmp_imm(0, 0);
    b.bcond(1, "digits");
    b.movz(3, 48);
    b.strb(3, 1);
    b.mov(0, 1);
    b.movz(1, 1);
    b.bl("rt_write");
    b.b("numdone");
    b.label("digits");
    b.udiv(3, 0, 2);
    b.msub(4, 3, 2, 0);
    b.add_imm(4, 4, 48);
    b.strb(4, 1);
    b.sub_imm(1, 1, 1);
    b.mov(0, 3);
    b.cmp_imm(0, 0);
    b.bcond(1, "digits");
    b.add_imm(1, 1, 1);
    b.sub(2, 5, 1);
    b.add_imm(2, 2, 1);
    b.mov(0, 1);
    b.mov(1, 2);
    b.bl("rt_write");
    b.label("numdone");
    b.add_imm(31, 31, 48);
    b.ldp_post(29, 30, 31, 16);
    b.ret();

    b.label("lang_methods");
    for method in methods {
        b.bl(&method.name);
        b.b("await_resume");
    }
    for method in methods {
        emit_method(&mut b, method);
    }
    emit_main(&mut b, main);
    b.finish("_start", &[("actors", 64 * 64), ("mbox", 128 * 24), ("prom_val", 256 * 8), ("prom_ready", 256), ("next_actor", 8), ("next_promise", 8), ("mbox_r", 8), ("mbox_w", 8)], 64 * 64 + 128 * 24 + 256 * 8 + 256 + 32)
}

fn emit_method(b: &mut Buf, method: &Method) {
    b.label(&method.name);
    b.stp_pre(29, 30, 31, -16);
    b.stp_pre(20, 21, 31, -16);
    b.mov(20, 0);
    for step in &method.steps {
        match step {
            Step::Load(slot) => {
                b.mov(0, 20);
                b.movz(1, *slot);
                b.bl("rt_load");
            }
            Step::Add(n) => b.add_imm(0, 0, *n),
            Step::Store(slot) => {
                b.mov(2, 0);
                b.mov(0, 20);
                b.movz(1, *slot);
                b.bl("rt_store");
                b.mov(0, 2);
            }
        }
    }
    b.ldp_post(20, 21, 31, 16);
    b.ldp_post(29, 30, 31, 16);
    b.ret();
}

fn emit_main(b: &mut Buf, main: &[Main]) {
    b.label("lang_main");
    b.stp_pre(29, 30, 31, -16);
    let mut strings = Vec::new();
    for op in main {
        match op {
            Main::Actor(slot) => {
                b.bl("rt_actor");
                b.mov(19 + *slot as u8, 0);
            }
            Main::Await(slot, method) | Main::AwaitI64(slot, method) => {
                b.mov(0, 19 + *slot as u8);
                b.movz(1, *method);
                b.bl("rt_call");
                b.bl("rt_await");
                if matches!(op, Main::AwaitI64(_, _)) {
                    b.bl("rt_write_i64");
                }
            }
            Main::Write(bytes) => {
                let name = format!("S{}", strings.len());
                strings.push((name.clone(), bytes.clone()));
                b.adr(0, &name);
                b.movz(1, bytes.len() as u16);
                b.bl("rt_write");
            }
        }
    }
    b.ldp_post(29, 30, 31, 16);
    b.ret();
    for (name, bytes) in strings {
        b.label(&name);
        let mut padded = bytes;
        while padded.len() % 4 != 0 {
            padded.push(0);
        }
        for chunk in padded.chunks(4) {
            let mut word = [0u8; 4];
            word[..chunk.len()].copy_from_slice(chunk);
            b.word(u32::from_le_bytes(word));
        }
    }
}

#[cfg(test)]
pub(crate) fn simulate(object: &Object) -> Result<(Vec<u8>, u64), String> {
    let mut mem = vec![0u8; 0x2_0000];
    mem[..object.text.len()].copy_from_slice(&object.text);
    let mut x = [0u64; 31];
    let mut sp = mem.len() as u64;
    let mut pc = object.entry as usize;
    let mut flags = 0u8;
    let mut stderr = Vec::new();
    let mut hits = HashMap::new();
    for _ in 0..200_000 {
        if pc + 4 > object.text.len() {
            return Err(format!("pc {pc:#x} left the text"));
        }
        let word = u32::from_le_bytes(mem[pc..pc + 4].try_into().unwrap());
        let seen = hits.entry(pc).or_insert(0);
        *seen += 1;
        if *seen == 1000 {
            return Err(format!(
                "loop at {pc:#x} word {word:#x} x0={x0:#x} x1={x1:#x} x2={x2:#x} x8={x8:#x} x19={x19:#x} x30={x30:#x} sp={sp:#x} flags={flags} err={stderr:?}",
                x0 = x[0],
                x1 = x[1],
                x2 = x[2],
                x8 = x[8],
                x19 = x[19],
                x30 = x[30],
            ));
        }
        let mut next = pc + 4;
        if word == 0xD65F03C0 {
            next = x[30] as usize;
        } else if word == 0xD4000001 {
            match x[8] {
                64 => {
                    let fd = x[0];
                    let ptr = x[1] as usize;
                    let len = x[2] as usize;
                    let bytes = mem.get(ptr..ptr + len).ok_or_else(|| format!("write {ptr:#x}+{len}"))?;
                    if fd == 2 {
                        stderr.extend(bytes);
                    }
                    x[0] = len as u64;
                }
                93 => return Ok((stderr, x[0])),
                other => return Err(format!("syscall {other}")),
            }
        } else if word & 0xFFFFFC1F == 0xD61F0000 {
            next = x[((word >> 5) & 31) as usize] as usize;
        } else if word & 0xFC000000 == 0x14000000 || word & 0xFC000000 == 0x94000000 {
            if word & 0xFC000000 == 0x94000000 {
                x[30] = next as u64;
            }
            next = (pc as i64 + sextract(word, 26) * 4) as usize;
        } else if word & 0xFF000010 == 0x54000000 {
            let cond = word & 15;
            let take = match cond {
                1 => flags != 2,
                2 => flags != 0,
                _ => return Err(format!("cond {cond} at {pc:#x}")),
            };
            if take {
                next = (pc as i64 + sextract(word >> 5, 19) * 4) as usize;
            }
        } else if word & 0x9F000000 == 0x10000000 {
            let rd = (word & 31) as usize;
            let imm = ((word >> 5) & 0x7FFFF) << 2 | (word >> 29) & 3;
            x[rd] = (pc as i64 + sextract(imm, 21)) as u64;
        } else if word & 0xFF800000 == 0xD2800000 {
            x[(word & 31) as usize] = ((word >> 5) & 0xFFFF) as u64;
        } else if word & 0xFFE0FFE0 == 0xAA0003E0 {
            let rm = ((word >> 16) & 31) as usize;
            let rd = (word & 31) as usize;
            x[rd] = if rm == 31 { 0 } else { x[rm] };
        } else if word & 0xFF000000 == 0x91000000 || word & 0xFF000000 == 0xD1000000 {
            let imm = ((word >> 10) & 0xFFF) as u64;
            let rn = ((word >> 5) & 31) as u8;
            let rd = (word & 31) as u8;
            let base = if rn == 31 { sp } else { x[rn as usize] };
            let value = if word & 0xFF000000 == 0x91000000 { base.wrapping_add(imm) } else { base.wrapping_sub(imm) };
            if rd == 31 {
                sp = value;
            } else {
                x[rd as usize] = value;
            }
        } else if word & 0xFF200000 == 0x8B000000 {
            let shift = (word >> 10) & 63;
            let rm = reg(x, (word >> 16) & 31);
            let rn = reg(x, (word >> 5) & 31);
            x[(word & 31) as usize] = rn.wrapping_add(rm.wrapping_shl(shift));
        } else if word & 0xFF200000 == 0xCB000000 {
            let rm = reg(x, (word >> 16) & 31);
            let rn = reg(x, (word >> 5) & 31);
            x[(word & 31) as usize] = rn.wrapping_sub(rm);
        } else if word & 0xFFC00000 == 0xD3400000 {
            let immr = (word >> 16) & 63;
            let shift = (64 - immr) & 63;
            let rn = reg(x, (word >> 5) & 31);
            x[(word & 31) as usize] = rn.wrapping_shl(shift);
        } else if word & 0xFF00001F == 0xF100001F {
            let imm = ((word >> 10) & 0xFFF) as u64;
            let rn = reg(x, (word >> 5) & 31);
            flags = if rn == imm { 2 } else if rn >= imm { 1 } else { 0 };
        } else if word & 0xFF20001F == 0xEB00001F {
            let rm = reg(x, (word >> 16) & 31);
            let rn = reg(x, (word >> 5) & 31);
            flags = if rn == rm { 2 } else if rn >= rm { 1 } else { 0 };
        } else if word & 0xFFC00000 == 0xF9400000 || word & 0xFFC00000 == 0xF9000000 {
            let off = (((word >> 10) & 0xFFF) * 8) as u64;
            let rn = ((word >> 5) & 31) as u8;
            let rt = (word & 31) as u8;
            let addr = (if rn == 31 { sp } else { x[rn as usize] }).wrapping_add(off) as usize;
            if word & 0xFFC00000 == 0xF9400000 {
                let val = load64(&mem, addr)?;
                if rt != 31 {
                    x[rt as usize] = val;
                }
            } else {
                store64(&mut mem, addr, if rt == 31 { 0 } else { x[rt as usize] })?;
            }
        } else if word & 0xFFC00000 == 0x39400000 || word & 0xFFC00000 == 0x39000000 {
            let off = ((word >> 10) & 0xFFF) as usize;
            let rn = ((word >> 5) & 31) as u8;
            let rt = (word & 31) as u8;
            let addr = (if rn == 31 { sp } else { x[rn as usize] }) as usize + off;
            if word & 0xFFC00000 == 0x39400000 {
                let val = *mem.get(addr).ok_or_else(|| format!("ldrb {addr:#x}"))? as u64;
                if rt != 31 {
                    x[rt as usize] = val;
                }
            } else {
                let val = if rt == 31 { 0 } else { x[rt as usize] as u8 };
                *mem.get_mut(addr).ok_or_else(|| format!("strb {addr:#x}"))? = val;
            }
        } else if word & 0xFFE0FC00 == 0x9AC00800 {
            let rm = reg(x, (word >> 16) & 31);
            let rn = reg(x, (word >> 5) & 31);
            if rm == 0 {
                return Err("division by zero".into());
            }
            x[(word & 31) as usize] = rn / rm;
        } else if word & 0xFFE08000 == 0x9B008000 {
            let rm = reg(x, (word >> 16) & 31);
            let ra = reg(x, (word >> 10) & 31);
            let rn = reg(x, (word >> 5) & 31);
            x[(word & 31) as usize] = ra.wrapping_sub(rn.wrapping_mul(rm));
        } else if word & 0xFFC00000 == 0xA9800000 || word & 0xFFC00000 == 0xA8C00000 {
            let imm = sextract(word >> 15, 7) * 8;
            let rt2 = ((word >> 10) & 31) as usize;
            let rn = ((word >> 5) & 31) as u8;
            let rt = (word & 31) as usize;
            let base = if rn == 31 { sp } else { x[rn as usize] };
            if word & 0xFFC00000 == 0xA9800000 {
                let addr = (base as i64 + imm) as u64;
                if rn == 31 {
                    sp = addr;
                }
                store64(&mut mem, addr as usize, x[rt])?;
                store64(&mut mem, addr as usize + 8, x[rt2])?;
            } else {
                let lo = load64(&mem, base as usize)?;
                let hi = load64(&mem, base as usize + 8)?;
                x[rt] = lo;
                x[rt2] = hi;
                let addr = (base as i64 + imm) as u64;
                if rn == 31 {
                    sp = addr;
                }
            }
        } else {
            return Err(format!("word {word:#x} at {pc:#x}"));
        }
        pc = next;
    }
    Err("arm64 step limit".into())
}

#[cfg(test)]
fn reg(x: [u64; 31], n: u32) -> u64 {
    if n == 31 { 0 } else { x[n as usize] }
}

#[cfg(test)]
fn sextract(value: u32, bits: u32) -> i64 {
    let mask = (1u32 << bits) - 1;
    let v = value & mask;
    let sign = 1u32 << (bits - 1);
    if v & sign == 0 { v as i64 } else { (v | !mask) as i32 as i64 }
}

#[cfg(test)]
fn load64(mem: &[u8], addr: usize) -> Result<u64, String> {
    let bytes = mem.get(addr..addr + 8).ok_or_else(|| format!("load {addr:#x}"))?;
    Ok(u64::from_le_bytes(bytes.try_into().unwrap()))
}

#[cfg(test)]
fn store64(mem: &mut [u8], addr: usize, value: u64) -> Result<(), String> {
    let bytes = mem.get_mut(addr..addr + 8).ok_or_else(|| format!("store {addr:#x}"))?;
    bytes.copy_from_slice(&value.to_le_bytes());
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::{simulate, Buf, Main, Method, Step};

    #[test]
    fn encodings_match_known_words() {
        let mut b = Buf::new();
        b.lsl(0, 0, 1);
        b.stp_pre(29, 30, 31, -16);
        b.ldp_post(29, 30, 31, 16);
        b.add_lsl(8, 6, 6, 1);
        let words: Vec<u32> = b.text.chunks(4).map(|c| u32::from_le_bytes(c.try_into().unwrap())).collect();
        assert_eq!(words, vec![0xD37FF800, 0xA9BF7BFD, 0xA8C17BFD, 0x8B0604C8]);
    }

    #[test]
    fn counter_prints_one() {
        let methods = vec![
            Method { name: "Counter_incr".into(), steps: vec![Step::Load(0), Step::Add(1), Step::Store(0)] },
            Method { name: "Counter_value".into(), steps: vec![Step::Load(0)] },
        ];
        let main = vec![
            Main::Actor(0),
            Main::Await(0, 0),
            Main::Write(b"count: ".to_vec()),
            Main::AwaitI64(0, 1),
            Main::Write(b"\n".to_vec()),
        ];
        let object = super::assemble(&methods, &main).unwrap();
        let (stderr, code) = simulate(&object).unwrap();
        assert_eq!(code, 0);
        assert_eq!(stderr, b"count: 1\n");
    }
}
