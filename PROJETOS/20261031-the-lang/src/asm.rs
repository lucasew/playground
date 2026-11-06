//! One assembly file. Callers compose it. [`File::make`] prints it. [`File::take`] reads that text back.

#[derive(Clone, Debug, PartialEq, Eq)]
pub struct File {
    pub intel: bool,
    pub items: Vec<Item>,
}

#[derive(Clone, Debug, PartialEq, Eq)]
pub enum Item {
    Section(Section),
    Align(u64),
    Global(String),
    Label(String),
    Inst(Inst),
    Quad(Vec<String>),
    Ascii(Vec<u8>),
    Zero(u64),
}

#[derive(Clone, Debug, PartialEq, Eq)]
pub enum Section {
    Text,
    Data,
    Bss,
    Rodata,
    NoteGnuStack,
}

#[derive(Clone, Debug, PartialEq, Eq)]
pub enum Operand {
    Reg(String),
    Imm(i64),
    Sym(String),
    Mem(Mem),
}

#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Mem {
    pub size: Option<Size>,
    pub base: MemBase,
    pub index: Option<String>,
    pub scale: u8,
    pub disp: i64,
}

#[derive(Clone, Debug, PartialEq, Eq)]
pub enum MemBase {
    Reg(String),
    Rip(String),
}

#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Size {
    Byte,
    Qword,
}

/// One machine instruction. A jump names its condition and its label.
#[derive(Clone, Debug, PartialEq, Eq)]
pub enum Inst {
    Mov(Operand, Operand),
    Lea(Operand, Operand),
    Alu(Alu, Operand, Operand),
    Imul(Operand, Operand, Operand),
    Unary(Unary, Operand),
    Cmp(Operand, Operand),
    Push(Operand),
    Pop(Operand),
    Call(Target),
    Jump { when: Cond, to: String },
    Ret,
    Leave,
    Syscall,
    Xchg(Operand, Operand),
}

#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Alu {
    Add,
    Sub,
    Xor,
    Shl,
    Shr,
}

#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Unary {
    Inc,
    Dec,
    Div,
}

/// Why a jump is taken. `Always` is `jmp`.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Cond {
    Always,
    E,
    Ne,
    B,
    Ae,
}

#[derive(Clone, Debug, PartialEq, Eq)]
pub enum Target {
    Sym(String),
    Reg(String),
}

impl File {
    pub fn new() -> Self {
        Self { intel: false, items: Vec::new() }
    }

    pub fn intel(&mut self) -> &mut Self {
        self.intel = true;
        self
    }

    pub fn text(&mut self) -> &mut Self {
        self.section(Section::Text)
    }

    pub fn data(&mut self) -> &mut Self {
        self.section(Section::Data)
    }

    pub fn bss(&mut self) -> &mut Self {
        self.section(Section::Bss)
    }

    pub fn rodata(&mut self) -> &mut Self {
        self.section(Section::Rodata)
    }

    pub fn note_gnu_stack(&mut self) -> &mut Self {
        self.section(Section::NoteGnuStack)
    }

    pub fn section(&mut self, section: Section) -> &mut Self {
        self.items.push(Item::Section(section));
        self
    }

    pub fn align(&mut self, n: u64) -> &mut Self {
        self.items.push(Item::Align(n));
        self
    }

    pub fn global(&mut self, name: &str) -> &mut Self {
        self.items.push(Item::Global(name.to_string()));
        self
    }

    pub fn label(&mut self, name: &str) -> &mut Self {
        self.items.push(Item::Label(name.to_string()));
        self
    }

    pub fn mov(&mut self, dst: Operand, src: Operand) -> &mut Self {
        self.inst(Inst::Mov(dst, src))
    }

    pub fn lea(&mut self, dst: Operand, src: Operand) -> &mut Self {
        self.inst(Inst::Lea(dst, src))
    }

    pub fn alu(&mut self, op: Alu, dst: Operand, src: Operand) -> &mut Self {
        self.inst(Inst::Alu(op, dst, src))
    }

    pub fn imul(&mut self, dst: Operand, a: Operand, b: Operand) -> &mut Self {
        self.inst(Inst::Imul(dst, a, b))
    }

    pub fn unary(&mut self, op: Unary, value: Operand) -> &mut Self {
        self.inst(Inst::Unary(op, value))
    }

    pub fn cmp(&mut self, left: Operand, right: Operand) -> &mut Self {
        self.inst(Inst::Cmp(left, right))
    }

    pub fn push(&mut self, value: Operand) -> &mut Self {
        self.inst(Inst::Push(value))
    }

    pub fn pop(&mut self, value: Operand) -> &mut Self {
        self.inst(Inst::Pop(value))
    }

    pub fn call(&mut self, name: &str) -> &mut Self {
        self.inst(Inst::Call(Target::Sym(name.to_string())))
    }

    pub fn call_reg(&mut self, name: &str) -> &mut Self {
        self.inst(Inst::Call(Target::Reg(name.to_string())))
    }

    pub fn jump(&mut self, when: Cond, to: &str) -> &mut Self {
        self.inst(Inst::Jump { when, to: to.to_string() })
    }

    pub fn ret(&mut self) -> &mut Self {
        self.inst(Inst::Ret)
    }

    pub fn leave(&mut self) -> &mut Self {
        self.inst(Inst::Leave)
    }

    pub fn syscall(&mut self) -> &mut Self {
        self.inst(Inst::Syscall)
    }

    pub fn xchg(&mut self, dst: Operand, src: Operand) -> &mut Self {
        self.inst(Inst::Xchg(dst, src))
    }

    fn inst(&mut self, inst: Inst) -> &mut Self {
        self.items.push(Item::Inst(inst));
        self
    }

    pub fn quad(&mut self, items: &[&str]) -> &mut Self {
        self.items.push(Item::Quad(items.iter().map(|s| (*s).to_string()).collect()));
        self
    }

    pub fn ascii(&mut self, bytes: &[u8]) -> &mut Self {
        self.items.push(Item::Ascii(bytes.to_vec()));
        self
    }

    pub fn zero(&mut self, n: u64) -> &mut Self {
        self.items.push(Item::Zero(n));
        self
    }

    pub fn make(&self) -> String {
        let mut out = String::new();
        if self.intel {
            out.push_str(".intel_syntax noprefix\n");
        }
        for item in &self.items {
            match item {
                Item::Section(Section::Text) => out.push_str(".text\n"),
                Item::Section(Section::Data) => out.push_str(".data\n"),
                Item::Section(Section::Bss) => out.push_str(".bss\n"),
                Item::Section(Section::Rodata) => out.push_str(".section .rodata\n"),
                Item::Section(Section::NoteGnuStack) => {
                    out.push_str(".section .note.GNU-stack,\"\",@progbits\n");
                }
                Item::Align(n) => out.push_str(&format!(".align {n}\n")),
                Item::Global(name) => out.push_str(&format!(".global {name}\n")),
                Item::Label(name) => out.push_str(&format!("{name}:\n")),
                Item::Inst(inst) => out.push_str(&make_inst(inst)),
                Item::Quad(items) => out.push_str(&format!("    .quad {}\n", items.join(", "))),
                Item::Ascii(bytes) => {
                    out.push_str("    .ascii \"");
                    for b in bytes {
                        out.push_str(&format!("\\x{b:02x}"));
                    }
                    out.push_str("\"\n");
                }
                Item::Zero(n) => out.push_str(&format!("    .zero {n}\n")),
            }
        }
        out
    }

    pub fn take(text: &str) -> Result<Self, String> {
        let mut file = File::new();
        for (n, raw) in text.lines().enumerate() {
            let line_no = n + 1;
            let line = raw.trim();
            if line.is_empty() || line.starts_with('#') {
                continue;
            }
            if line == ".intel_syntax noprefix" {
                file.intel = true;
                continue;
            }
            if line == ".text" {
                file.text();
                continue;
            }
            if line == ".data" {
                file.data();
                continue;
            }
            if line == ".bss" {
                file.bss();
                continue;
            }
            if line == ".section .rodata" {
                file.rodata();
                continue;
            }
            if line == ".section .note.GNU-stack,\"\",@progbits" {
                file.note_gnu_stack();
                continue;
            }
            if let Some(rest) = line.strip_prefix(".align ") {
                file.align(parse_u64(rest, line_no)?);
                continue;
            }
            if let Some(rest) = line.strip_prefix(".global ") {
                file.global(rest.trim());
                continue;
            }
            if let Some(rest) = line.strip_prefix(".quad ") {
                let items: Vec<&str> = split_args(rest).iter().map(|s| s.trim()).collect();
                file.quad(&items);
                continue;
            }
            if let Some(rest) = line.strip_prefix(".ascii ") {
                file.ascii(&parse_ascii(rest.trim(), line_no)?);
                continue;
            }
            if let Some(rest) = line.strip_prefix(".zero ") {
                file.zero(parse_u64(rest.trim(), line_no)?);
                continue;
            }
            if let Some(name) = line.strip_suffix(':') {
                file.label(name.trim());
                continue;
            }
            let (op, rest) = split_op(line);
            let args = if rest.is_empty() {
                Vec::new()
            } else {
                split_args(rest)
                    .iter()
                    .map(|part| parse_operand(part.trim(), line_no))
                    .collect::<Result<Vec<_>, _>>()?
            };
            file.inst(decode(op, args, line_no)?);
        }
        Ok(file)
    }
}

pub fn reg(name: &str) -> Operand {
    Operand::Reg(name.to_string())
}

pub fn imm(n: i64) -> Operand {
    Operand::Imm(n)
}

pub fn rip(name: &str) -> Operand {
    Operand::Mem(Mem { size: None, base: MemBase::Rip(name.to_string()), index: None, scale: 1, disp: 0 })
}

pub fn qword_rip(name: &str) -> Operand {
    Operand::Mem(Mem { size: Some(Size::Qword), base: MemBase::Rip(name.to_string()), index: None, scale: 1, disp: 0 })
}

pub fn byte_base(base: &str) -> Operand {
    Operand::Mem(Mem { size: Some(Size::Byte), base: MemBase::Reg(base.to_string()), index: None, scale: 1, disp: 0 })
}

pub fn byte_base_index(base: &str, index: &str) -> Operand {
    Operand::Mem(Mem {
        size: Some(Size::Byte),
        base: MemBase::Reg(base.to_string()),
        index: Some(index.to_string()),
        scale: 1,
        disp: 0,
    })
}

pub fn qword_base_disp(base: &str, disp: i64) -> Operand {
    Operand::Mem(Mem {
        size: Some(Size::Qword),
        base: MemBase::Reg(base.to_string()),
        index: None,
        scale: 1,
        disp,
    })
}

pub fn qword_base_index(base: &str, index: &str, scale: u8) -> Operand {
    Operand::Mem(Mem {
        size: Some(Size::Qword),
        base: MemBase::Reg(base.to_string()),
        index: Some(index.to_string()),
        scale,
        disp: 0,
    })
}

pub fn qword_base_index_disp(base: &str, index: &str, scale: u8, disp: i64) -> Operand {
    Operand::Mem(Mem {
        size: Some(Size::Qword),
        base: MemBase::Reg(base.to_string()),
        index: Some(index.to_string()),
        scale,
        disp,
    })
}

pub fn base_disp(base: &str, disp: i64) -> Operand {
    Operand::Mem(Mem { size: None, base: MemBase::Reg(base.to_string()), index: None, scale: 1, disp })
}

fn make_inst(inst: &Inst) -> String {
    let (op, args): (&str, Vec<String>) = match inst {
        Inst::Mov(dst, src) => ("mov", vec![make_operand(dst), make_operand(src)]),
        Inst::Lea(dst, src) => ("lea", vec![make_operand(dst), make_operand(src)]),
        Inst::Alu(op, dst, src) => (alu_name(*op), vec![make_operand(dst), make_operand(src)]),
        Inst::Imul(dst, a, b) => ("imul", vec![make_operand(dst), make_operand(a), make_operand(b)]),
        Inst::Unary(op, value) => (unary_name(*op), vec![make_operand(value)]),
        Inst::Cmp(left, right) => ("cmp", vec![make_operand(left), make_operand(right)]),
        Inst::Push(value) => ("push", vec![make_operand(value)]),
        Inst::Pop(value) => ("pop", vec![make_operand(value)]),
        Inst::Call(Target::Sym(name)) => ("call", vec![name.clone()]),
        Inst::Call(Target::Reg(name)) => ("call", vec![name.clone()]),
        Inst::Jump { when, to } => (cond_name(*when), vec![to.clone()]),
        Inst::Ret => ("ret", Vec::new()),
        Inst::Leave => ("leave", Vec::new()),
        Inst::Syscall => ("syscall", Vec::new()),
        Inst::Xchg(dst, src) => ("xchg", vec![make_operand(dst), make_operand(src)]),
    };
    let mut line = format!("    {op}");
    if !args.is_empty() {
        line.push(' ');
        line.push_str(&args.join(", "));
    }
    line.push('\n');
    line
}

fn alu_name(op: Alu) -> &'static str {
    match op {
        Alu::Add => "add",
        Alu::Sub => "sub",
        Alu::Xor => "xor",
        Alu::Shl => "shl",
        Alu::Shr => "shr",
    }
}

fn unary_name(op: Unary) -> &'static str {
    match op {
        Unary::Inc => "inc",
        Unary::Dec => "dec",
        Unary::Div => "div",
    }
}

fn cond_name(when: Cond) -> &'static str {
    match when {
        Cond::Always => "jmp",
        Cond::E => "je",
        Cond::Ne => "jne",
        Cond::B => "jb",
        Cond::Ae => "jae",
    }
}

fn decode(op: &str, mut args: Vec<Operand>, line: usize) -> Result<Inst, String> {
    let need = |n: usize| -> Result<(), String> {
        if args.len() == n {
            Ok(())
        } else {
            Err(format!("line {line}: {op} needs {n} operands"))
        }
    };
    let two = |args: &mut Vec<Operand>| -> (Operand, Operand) { (args.remove(0), args.remove(0)) };
    Ok(match op {
        "mov" => {
            need(2)?;
            let (dst, src) = two(&mut args);
            Inst::Mov(dst, src)
        }
        "lea" => {
            need(2)?;
            let (dst, src) = two(&mut args);
            Inst::Lea(dst, src)
        }
        "add" | "sub" | "xor" | "shl" | "shr" => {
            need(2)?;
            let alu = match op {
                "add" => Alu::Add,
                "sub" => Alu::Sub,
                "xor" => Alu::Xor,
                "shl" => Alu::Shl,
                _ => Alu::Shr,
            };
            let (dst, src) = two(&mut args);
            Inst::Alu(alu, dst, src)
        }
        "imul" => {
            need(3)?;
            let dst = args.remove(0);
            let a = args.remove(0);
            let b = args.remove(0);
            Inst::Imul(dst, a, b)
        }
        "inc" | "dec" | "div" => {
            need(1)?;
            let unary = match op {
                "inc" => Unary::Inc,
                "dec" => Unary::Dec,
                _ => Unary::Div,
            };
            Inst::Unary(unary, args.remove(0))
        }
        "cmp" => {
            need(2)?;
            let (left, right) = two(&mut args);
            Inst::Cmp(left, right)
        }
        "push" => {
            need(1)?;
            Inst::Push(args.remove(0))
        }
        "pop" => {
            need(1)?;
            Inst::Pop(args.remove(0))
        }
        "call" => {
            need(1)?;
            Inst::Call(match args.remove(0) {
                Operand::Reg(name) => Target::Reg(name),
                Operand::Sym(name) => Target::Sym(name),
                other => return Err(format!("line {line}: call cannot target {other:?}")),
            })
        }
        "jmp" | "je" | "jne" | "jb" | "jae" => {
            need(1)?;
            let when = match op {
                "jmp" => Cond::Always,
                "je" => Cond::E,
                "jne" => Cond::Ne,
                "jb" => Cond::B,
                _ => Cond::Ae,
            };
            let Operand::Sym(to) = args.remove(0) else {
                return Err(format!("line {line}: a jump targets a label"));
            };
            Inst::Jump { when, to }
        }
        "ret" => {
            need(0)?;
            Inst::Ret
        }
        "leave" => {
            need(0)?;
            Inst::Leave
        }
        "syscall" => {
            need(0)?;
            Inst::Syscall
        }
        "xchg" => {
            need(2)?;
            let (dst, src) = two(&mut args);
            Inst::Xchg(dst, src)
        }
        _ => return Err(format!("line {line}: unknown instruction {op}")),
    })
}

fn make_operand(op: &Operand) -> String {
    match op {
        Operand::Reg(name) => name.clone(),
        Operand::Imm(n) => n.to_string(),
        Operand::Sym(name) => name.clone(),
        Operand::Mem(mem) => {
            let mut inner = match &mem.base {
                MemBase::Rip(name) => format!("rip+{name}"),
                MemBase::Reg(name) => name.clone(),
            };
            if let Some(index) = &mem.index {
                inner.push('+');
                inner.push_str(index);
                if mem.scale != 1 {
                    inner.push('*');
                    inner.push_str(&mem.scale.to_string());
                }
            }
            if mem.disp > 0 {
                inner.push('+');
                inner.push_str(&mem.disp.to_string());
            } else if mem.disp < 0 {
                inner.push_str(&mem.disp.to_string());
            }
            let body = format!("[{inner}]");
            match mem.size {
                Some(Size::Qword) => format!("qword ptr {body}"),
                Some(Size::Byte) => format!("byte ptr {body}"),
                None => body,
            }
        }
    }
}

fn split_op(line: &str) -> (&str, &str) {
    let line = line.trim();
    match line.find(char::is_whitespace) {
        Some(i) => (&line[..i], line[i..].trim()),
        None => (line, ""),
    }
}

fn split_args(text: &str) -> Vec<&str> {
    let mut args = Vec::new();
    let mut start = 0;
    let mut depth = 0;
    for (i, c) in text.char_indices() {
        match c {
            '[' => depth += 1,
            ']' => depth -= 1,
            ',' if depth == 0 => {
                args.push(&text[start..i]);
                start = i + 1;
            }
            _ => {}
        }
    }
    if start < text.len() {
        args.push(&text[start..]);
    }
    args
}

fn parse_operand(text: &str, line: usize) -> Result<Operand, String> {
    let text = text.trim();
    if let Some(rest) = text.strip_prefix("qword ptr ") {
        return parse_mem(rest, Some(Size::Qword), line);
    }
    if let Some(rest) = text.strip_prefix("byte ptr ") {
        return parse_mem(rest, Some(Size::Byte), line);
    }
    if text.starts_with('[') {
        return parse_mem(text, None, line);
    }
    if let Some(n) = text.parse::<i64>().ok() {
        return Ok(Operand::Imm(n));
    }
    if is_reg(text) {
        return Ok(Operand::Reg(text.to_string()));
    }
    if !text.is_empty() {
        return Ok(Operand::Sym(text.to_string()));
    }
    Err(format!("line {line}: empty operand"))
}

fn parse_mem(text: &str, size: Option<Size>, line: usize) -> Result<Operand, String> {
    let inner = text.trim().strip_prefix('[').and_then(|s| s.strip_suffix(']')).ok_or_else(|| {
        format!("line {line}: bad memory {text}")
    })?;
    let mut mem = parse_mem_inner(inner, line)?;
    mem.size = size;
    Ok(Operand::Mem(mem))
}

fn parse_mem_inner(text: &str, line: usize) -> Result<Mem, String> {
    if let Some(name) = text.strip_prefix("rip+") {
        return Ok(Mem { size: None, base: MemBase::Rip(name.to_string()), index: None, scale: 1, disp: 0 });
    }
    let (base, rest) = take_reg(text).ok_or_else(|| format!("line {line}: bad memory {text}"))?;
    if rest.is_empty() {
        return Ok(Mem { size: None, base: MemBase::Reg(base), index: None, scale: 1, disp: 0 });
    }
    if rest.starts_with('-') || (rest.starts_with('+') && rest[1..].chars().all(|c| c.is_ascii_digit())) {
        let disp = rest.parse::<i64>().map_err(|_| format!("line {line}: bad displacement {rest}"))?;
        return Ok(Mem { size: None, base: MemBase::Reg(base), index: None, scale: 1, disp });
    }
    if let Some(rest) = rest.strip_prefix('+') {
        if rest.chars().all(|c| c.is_ascii_digit()) {
            let disp = rest.parse::<i64>().map_err(|_| format!("line {line}: bad displacement"))?;
            return Ok(Mem { size: None, base: MemBase::Reg(base), index: None, scale: 1, disp });
        }
        let (index, rest) = take_reg(rest).ok_or_else(|| format!("line {line}: bad index in {text}"))?;
        let (scale, rest) = if let Some(rest) = rest.strip_prefix('*') {
            let (num, rest) = take_number(rest);
            (num.parse::<u8>().map_err(|_| format!("line {line}: bad scale"))?, rest)
        } else {
            (1, rest)
        };
        let disp = if rest.is_empty() {
            0
        } else {
            rest.parse::<i64>().map_err(|_| format!("line {line}: bad displacement {rest}"))?
        };
        return Ok(Mem { size: None, base: MemBase::Reg(base), index: Some(index), scale, disp });
    }
    Err(format!("line {line}: bad memory {text}"))
}

fn take_reg(text: &str) -> Option<(String, &str)> {
    const REGS: &[&str] = &[
        "rax", "rbx", "rcx", "rdx", "rsi", "rdi", "rbp", "rsp", "eax", "ebx", "ecx", "edx", "al", "r15",
        "r14", "r13", "r12", "r11", "r10", "r9", "r8", "dl",
    ];
    let mut best: Option<(&str, &str)> = None;
    for reg in REGS {
        if let Some(rest) = text.strip_prefix(reg) {
            if best.is_none_or(|(prev, _)| reg.len() > prev.len()) {
                best = Some((reg, rest));
            }
        }
    }
    best.map(|(reg, rest)| (reg.to_string(), rest))
}

fn take_number(text: &str) -> (&str, &str) {
    let n = text.chars().take_while(|c| c.is_ascii_digit()).count();
    text.split_at(n)
}

fn is_reg(text: &str) -> bool {
    take_reg(text).is_some_and(|(_, rest)| rest.is_empty())
}

fn parse_u64(text: &str, line: usize) -> Result<u64, String> {
    text.trim().parse::<u64>().map_err(|_| format!("line {line}: bad number {text}"))
}

fn parse_ascii(text: &str, line: usize) -> Result<Vec<u8>, String> {
    let body = text.strip_prefix('"').and_then(|s| s.strip_suffix('"')).ok_or_else(|| {
        format!("line {line}: bad ascii")
    })?;
    let mut out = Vec::new();
    let bytes = body.as_bytes();
    let mut i = 0;
    while i < bytes.len() {
        if bytes[i] == b'\\' && i + 3 < bytes.len() && bytes[i + 1] == b'x' {
            let hex = std::str::from_utf8(&bytes[i + 2..i + 4]).unwrap_or("");
            let b = u8::from_str_radix(hex, 16).map_err(|_| format!("line {line}: bad ascii"))?;
            out.push(b);
            i += 4;
        } else {
            out.push(bytes[i]);
            i += 1;
        }
    }
    Ok(out)
}
