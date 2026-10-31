//! Lower a program to calls on the runtime. The runtime calls `hal_write` and `hal_exit` only.

use std::path::Path;

use crate::amd64::Object;
use crate::asm::{self, Alu, Cond, File, Operand, Unary};
use crate::ast::*;
use crate::link::{self, Cpu, Os};
use crate::parse;

pub fn build_linux_amd64(src: &str, output: &Path) -> Result<(), String> {
    build(src, Os::Linux, Cpu::Amd64, output)
}

pub fn build(src: &str, os: Os, cpu: Cpu, output: &Path) -> Result<(), String> {
    let program = parse::parse(src)?;
    let object = match cpu {
        Cpu::Amd64 => {
            let prog_file = lower(&program)?;
            emit(&prog_file)?;
            emit(&hal_linux_amd64())?;
            emit(&runtime_amd64())?;
            Object::assemble(&[hal_linux_amd64(), runtime_amd64(), prog_file])?
        }
        Cpu::Arm64 => arm_object(&program)?,
    };
    let bytes = link::write(&object, os, cpu)?;
    if let Some(parent) = output.parent() {
        if !parent.as_os_str().is_empty() {
            std::fs::create_dir_all(parent).map_err(|e| e.to_string())?;
        }
    }
    std::fs::write(output, bytes).map_err(|e| e.to_string())?;
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt;
        let _ = std::fs::set_permissions(output, std::fs::Permissions::from_mode(0o755));
    }
    Ok(())
}

fn emit(file: &File) -> Result<String, String> {
    let text = file.make();
    let again = File::take(&text)?;
    let rendered = again.make();
    if rendered != text {
        return Err("assembly model did not round-trip".into());
    }
    Ok(text)
}

struct MethodRef {
    type_name: String,
    name: String,
    fields: Vec<String>,
    body: Vec<Stmt>,
}

fn lower(program: &Program) -> Result<File, String> {
    let mut methods = Vec::new();
    let mut actors: Vec<(String, Vec<String>)> = Vec::new();
    let mut main = None;
    for item in &program.items {
        match item {
            Item::Actor { name, fields, .. } => {
                actors.push((name.clone(), fields.iter().map(|f| f.name.clone()).collect()));
            }
            Item::Method { type_name, name, body, .. } => {
                let fields = actors
                    .iter()
                    .find(|(n, _)| n == type_name)
                    .map(|(_, f)| f.clone())
                    .unwrap_or_default();
                methods.push(MethodRef {
                    type_name: type_name.clone(),
                    name: name.clone(),
                    fields,
                    body: body.clone(),
                });
            }
            Item::Main(body) => main = Some(body.clone()),
            Item::Test(_) => {}
            Item::ImportNamed { .. } | Item::ImportModule { .. } | Item::Func { .. } => {}
            Item::Struct { .. } | Item::Enum { .. } => {
                return Err("build: this port lowers actors, integer fields, and fprintf".into());
            }
        }
    }
    let main = main.ok_or("program has no main")?;
    let mut file = File::new();
    file.intel().text();
    let mut names = Vec::new();
    for (i, method) in methods.iter().enumerate() {
        let symbol = format!("m{i}_{}_{}", method.type_name, method.name);
        names.push(symbol.clone());
        file.global(&symbol).label(&symbol);
        prologue(&mut file);
        file.push(reg("rbx")).mov(reg("rbx"), reg("rdi"));
        lower_method(&method.body, &method.fields, &mut file)?;
        file.pop(reg("rbx")).pop(reg("rbp")).ret();
    }
    file.global("lang_methods").label("lang_methods");
    if names.is_empty() {
        file.quad(&["0"]);
    } else {
        let refs: Vec<&str> = names.iter().map(String::as_str).collect();
        file.quad(&refs);
    }
    file.global("lang_main").label("lang_main");
    prologue(&mut file);
    file.alu(Alu::Sub, reg("rsp"), imm(64));
    let strings = lower_main(&main, &actors, &methods, &mut file)?;
    file.leave().ret();
    if !strings.is_empty() {
        file.rodata();
        for (i, text) in strings.iter().enumerate() {
            file.label(&format!("S{i}")).ascii(text.as_bytes());
        }
    }
    file.note_gnu_stack();
    Ok(file)
}

fn prologue(file: &mut File) {
    file.push(reg("rbp")).mov(reg("rbp"), reg("rsp"));
}

fn reg(name: &str) -> Operand {
    asm::reg(name)
}

fn imm(n: i64) -> Operand {
    asm::imm(n)
}

fn lower_method(body: &[Stmt], fields: &[String], file: &mut File) -> Result<(), String> {
    if body.is_empty() {
        file.alu(Alu::Xor, reg("rax"), reg("rax"));
        return Ok(());
    }
    for stmt in body {
        match stmt {
            Stmt::TestBlock(_) => {}
            Stmt::Assign { target, op, value } => {
                let Expr::Field { base, name } = target else {
                    return Err("build: a method assigns self.field".into());
                };
                let Expr::Name(recv) = base.as_ref() else {
                    return Err("build: a method assigns self.field".into());
                };
                if recv != "self" {
                    return Err("build: a method assigns self.field".into());
                }
                let slot = field_slot(fields, name)? as i64;
                match op {
                    AssignOp::PlusEq => {
                        let Expr::Int(n) = value else {
                            return Err("build: self.field += needs an integer".into());
                        };
                        file.mov(reg("rdi"), reg("rbx"))
                            .mov(reg("rsi"), imm(slot))
                            .call("rt_load")
                            .alu(Alu::Add, reg("rax"), imm(*n))
                            .mov(reg("rdx"), reg("rax"))
                            .mov(reg("rdi"), reg("rbx"))
                            .mov(reg("rsi"), imm(slot))
                            .call("rt_store");
                    }
                    AssignOp::Eq => return Err("build: this port lowers self.field +=".into()),
                }
            }
            Stmt::Expr(Expr::Field { base, name }) => {
                let Expr::Name(recv) = base.as_ref() else {
                    return Err("build: a method returns self.field".into());
                };
                if recv != "self" {
                    return Err("build: a method returns self.field".into());
                }
                let slot = field_slot(fields, name)? as i64;
                file.mov(reg("rdi"), reg("rbx"))
                    .mov(reg("rsi"), imm(slot))
                    .call("rt_load");
            }
            Stmt::Expr(_) => return Err("build: this port lowers self.field in a method".into()),
        }
    }
    Ok(())
}

fn lower_main(
    body: &[Stmt],
    actors: &[(String, Vec<String>)],
    methods: &[MethodRef],
    file: &mut File,
) -> Result<Vec<String>, String> {
    let mut locals: Vec<String> = Vec::new();
    let mut strings: Vec<String> = Vec::new();
    for stmt in body {
        match stmt {
            Stmt::TestBlock(_) => {}
            Stmt::Assign { target, op, value } => {
                if !matches!(op, AssignOp::Eq) {
                    return Err("build: main assigns with =".into());
                }
                let Expr::Name(name) = target else {
                    return Err("build: main assigns a local".into());
                };
                let Expr::Construct { name: actor, .. } = value else {
                    return Err("build: main assigns an actor".into());
                };
                if !actors.iter().any(|(n, _)| n == actor) {
                    return Err(format!("build: no actor {actor}"));
                }
                let off = (locals.len() as i64 + 1) * 8;
                locals.push(name.clone());
                file.call("rt_actor")
                    .mov(asm::qword_base_disp("rbp", -off), reg("rax"));
            }
            Stmt::Expr(expr) => lower_main_expr(expr, &locals, methods, &mut strings, file)?,
        }
    }
    Ok(strings)
}

fn lower_main_expr(
    expr: &Expr,
    locals: &[String],
    methods: &[MethodRef],
    strings: &mut Vec<String>,
    file: &mut File,
) -> Result<(), String> {
    if let Some((local, method)) = await_call(expr) {
        emit_await(local, method, locals, methods, file)?;
        return Ok(());
    }
    if let Expr::Call { callee, args } = expr {
        let Expr::Name(name) = callee.as_ref() else {
            return Err("build: unsupported call".into());
        };
        if name != "fprintf" {
            return Err(format!("build: cannot call {name}"));
        }
        if args.len() < 2 {
            return Err("build: fprintf needs a format".into());
        }
        let Expr::Str(fmt) = &args[1] else {
            return Err("build: fprintf format must be a literal".into());
        };
        let mut rest = args[2..].iter();
        for piece in format_pieces(fmt)? {
            match piece {
                Piece::Text(text) => {
                    let id = strings.len();
                    let len = text.len() as i64;
                    strings.push(text);
                    file.lea(reg("rdi"), asm::rip(&format!("S{id}")))
                        .mov(reg("rsi"), imm(len))
                        .call("rt_write");
                }
                Piece::Int => {
                    let arg = rest.next().ok_or("build: fprintf is missing a value")?;
                    let Some((local, method)) = await_call(arg) else {
                        return Err("build: fprintf %d needs an awaited call".into());
                    };
                    emit_await(local, method, locals, methods, file)?;
                    file.mov(reg("rdi"), reg("rax")).call("rt_write_i64");
                }
            }
        }
        return Ok(());
    }
    Err("build: unsupported statement in main".into())
}

fn emit_await(
    local: String,
    method: String,
    locals: &[String],
    methods: &[MethodRef],
    file: &mut File,
) -> Result<(), String> {
    let slot = locals.iter().position(|n| n == &local).ok_or_else(|| format!("build: no local {local}"))?;
    let id = methods.iter().position(|m| m.name == method).ok_or_else(|| format!("build: no method {method}"))?;
    let off = (slot as i64 + 1) * 8;
    file.mov(reg("rdi"), asm::qword_base_disp("rbp", -off))
        .mov(reg("rsi"), imm(id as i64))
        .call("rt_call")
        .mov(reg("rdi"), reg("rax"))
        .call("rt_await");
    Ok(())
}

fn await_call(expr: &Expr) -> Option<(String, String)> {
    let Expr::Call { callee, args } = expr else {
        return None;
    };
    if !args.is_empty() {
        return None;
    }
    let Expr::Field { base, name } = callee.as_ref() else {
        return None;
    };
    if name != "unwrap" {
        return None;
    }
    let Expr::Await(inner) = base.as_ref() else {
        return None;
    };
    let Expr::Call { callee, args } = inner.as_ref() else {
        return None;
    };
    if !args.is_empty() {
        return None;
    }
    let Expr::Field { base, name } = callee.as_ref() else {
        return None;
    };
    let Expr::Name(local) = base.as_ref() else {
        return None;
    };
    Some((local.clone(), name.clone()))
}

enum Piece {
    Text(String),
    Int,
}

fn format_pieces(fmt: &str) -> Result<Vec<Piece>, String> {
    let mut pieces = Vec::new();
    let bytes = fmt.as_bytes();
    let mut i = 0;
    let mut buf = String::new();
    while i < bytes.len() {
        if bytes[i] == b'%' && i + 1 < bytes.len() {
            match bytes[i + 1] {
                b'd' => {
                    if !buf.is_empty() {
                        pieces.push(Piece::Text(std::mem::take(&mut buf)));
                    }
                    pieces.push(Piece::Int);
                }
                b'%' => buf.push('%'),
                other => return Err(format!("build: fprintf %{other} is not in this port")),
            }
            i += 2;
        } else {
            buf.push(bytes[i] as char);
            i += 1;
        }
    }
    if !buf.is_empty() {
        pieces.push(Piece::Text(buf));
    }
    Ok(pieces)
}

fn arm_object(program: &Program) -> Result<crate::amd64::Object, String> {
    let mut actors: Vec<(String, Vec<String>)> = Vec::new();
    let mut methods = Vec::new();
    let mut main = None;
    for item in &program.items {
        match item {
            Item::Actor { name, fields, .. } => {
                actors.push((name.clone(), fields.iter().map(|f| f.name.clone()).collect()));
            }
            Item::Method { type_name, name, body, .. } => {
                let fields = actors.iter().find(|(n, _)| n == type_name).map(|(_, f)| f.clone()).unwrap_or_default();
                let mut steps = Vec::new();
                for stmt in body {
                    match stmt {
                        Stmt::TestBlock(_) => {}
                        Stmt::Assign { target, op, value } => {
                            let Expr::Field { base, name } = target else {
                                return Err("build: a method assigns self.field".into());
                            };
                            let Expr::Name(recv) = base.as_ref() else {
                                return Err("build: a method assigns self.field".into());
                            };
                            if recv != "self" || !matches!(op, AssignOp::PlusEq) {
                                return Err("build: this port lowers self.field +=".into());
                            }
                            let Expr::Int(n) = value else {
                                return Err("build: self.field += needs an integer".into());
                            };
                            let slot = field_slot(&fields, name)? as u16;
                            steps.push(crate::arm64::Step::Load(slot));
                            steps.push(crate::arm64::Step::Add(*n as u16));
                            steps.push(crate::arm64::Step::Store(slot));
                        }
                        Stmt::Expr(Expr::Field { base, name }) => {
                            let Expr::Name(recv) = base.as_ref() else {
                                return Err("build: a method returns self.field".into());
                            };
                            if recv != "self" {
                                return Err("build: a method returns self.field".into());
                            }
                            steps.push(crate::arm64::Step::Load(field_slot(&fields, name)? as u16));
                        }
                        Stmt::Expr(_) => return Err("build: this port lowers self.field in a method".into()),
                    }
                }
                methods.push(crate::arm64::Method { name: format!("{type_name}_{name}"), steps });
            }
            Item::Main(body) => main = Some(body),
            Item::Test(_) | Item::ImportNamed { .. } | Item::ImportModule { .. } | Item::Func { .. } => {}
            Item::Struct { .. } | Item::Enum { .. } => {
                return Err("build: this port lowers actors, integer fields, and fprintf".into());
            }
        }
    }
    let main = main.ok_or("program has no main")?;
    let mut ops = Vec::new();
    let mut locals = Vec::new();
    for stmt in main {
        match stmt {
            Stmt::TestBlock(_) => {}
            Stmt::Assign { target, value, .. } => {
                let Expr::Name(name) = target else { return Err("build: main assigns a local".into()) };
                let Expr::Construct { name: actor, .. } = value else {
                    return Err("build: main assigns an actor".into());
                };
                if !actors.iter().any(|(n, _)| n == actor) {
                    return Err(format!("build: no actor {actor}"));
                }
                ops.push(crate::arm64::Main::Actor(locals.len() as u16));
                locals.push(name.clone());
            }
            Stmt::Expr(expr) => {
                if let Some((local, method)) = await_call(expr) {
                    let slot = locals.iter().position(|n| n == &local).ok_or_else(|| format!("build: no local {local}"))?;
                    let id = methods.iter().position(|m| m.name.ends_with(&format!("_{method}"))).ok_or_else(|| format!("build: no method {method}"))?;
                    ops.push(crate::arm64::Main::Await(slot as u16, id as u16));
                } else if let Expr::Call { args, .. } = expr {
                    let Expr::Str(fmt) = &args[1] else { return Err("build: fprintf format must be a literal".into()) };
                    let mut rest = args[2..].iter();
                    for piece in format_pieces(fmt)? {
                        match piece {
                            Piece::Text(text) => ops.push(crate::arm64::Main::Write(text.into_bytes())),
                            Piece::Int => {
                                let arg = rest.next().ok_or("build: fprintf is missing a value")?;
                                let Some((local, method)) = await_call(arg) else {
                                    return Err("build: fprintf %d needs an awaited call".into());
                                };
                                let slot = locals.iter().position(|n| n == &local).ok_or_else(|| format!("build: no local {local}"))?;
                                let id = methods.iter().position(|m| m.name.ends_with(&format!("_{method}"))).ok_or_else(|| format!("build: no method {method}"))?;
                                ops.push(crate::arm64::Main::AwaitI64(slot as u16, id as u16));
                            }
                        }
                    }
                } else {
                    return Err("build: unsupported statement in main".into());
                }
            }
        }
    }
    crate::arm64::assemble(&methods, &ops)
}

fn field_slot(fields: &[String], name: &str) -> Result<usize, String> {
    fields.iter().position(|n| n == name).ok_or_else(|| format!("build: no field {name}"))
}

fn hal_linux_amd64() -> File {
    let mut f = File::new();
    f.intel().text();
    f.global("_start").label("_start");
    f.alu(Alu::Xor, reg("rbp"), reg("rbp"));
    f.call("rt_start");
    f.mov(reg("rdi"), reg("rax"));
    f.call("hal_exit");
    f.global("hal_write").label("hal_write");
    f.mov(reg("rdx"), reg("rsi"));
    f.mov(reg("rsi"), reg("rdi"));
    f.mov(reg("rdi"), imm(2));
    f.mov(reg("rax"), imm(1));
    f.syscall();
    f.ret();
    f.global("hal_exit").label("hal_exit");
    f.mov(reg("rax"), imm(60));
    f.syscall();
    f.note_gnu_stack();
    f
}

fn runtime_amd64() -> File {
    let mut f = File::new();
    f.intel().text();
    f.global("rt_start").label("rt_start");
    f.call("lang_main");
    f.alu(Alu::Xor, reg("rax"), reg("rax"));
    f.ret();

    f.global("rt_actor").label("rt_actor");
    f.mov(reg("rax"), asm::qword_rip("next_actor"));
    f.lea(reg("rcx"), asm::base_disp("rax", 1));
    f.mov(asm::qword_rip("next_actor"), reg("rcx"));
    f.mov(reg("rdi"), reg("rax"));
    f.alu(Alu::Shl, reg("rdi"), imm(6));
    f.lea(reg("rsi"), asm::rip("actors"));
    f.alu(Alu::Add, reg("rdi"), reg("rsi"));
    f.alu(Alu::Xor, reg("ecx"), reg("ecx"));
    f.label("Lzero");
    f.mov(asm::qword_base_index("rdi", "rcx", 8), imm(0));
    f.unary(Unary::Inc, reg("ecx"));
    f.cmp(reg("ecx"), imm(8));
    f.jump(Cond::B, "Lzero");
    f.ret();

    f.global("rt_load").label("rt_load");
    f.mov(reg("rax"), reg("rdi"));
    f.alu(Alu::Shl, reg("rax"), imm(6));
    f.lea(reg("rcx"), asm::rip("actors"));
    f.alu(Alu::Add, reg("rcx"), reg("rax"));
    f.mov(reg("rax"), asm::qword_base_index("rcx", "rsi", 8));
    f.ret();

    f.global("rt_store").label("rt_store");
    f.mov(reg("rax"), reg("rdi"));
    f.alu(Alu::Shl, reg("rax"), imm(6));
    f.lea(reg("rcx"), asm::rip("actors"));
    f.alu(Alu::Add, reg("rcx"), reg("rax"));
    f.mov(asm::qword_base_index("rcx", "rsi", 8), reg("rdx"));
    f.ret();

    f.global("rt_call").label("rt_call");
    f.mov(reg("rax"), asm::qword_rip("next_promise"));
    f.lea(reg("rcx"), asm::base_disp("rax", 1));
    f.mov(asm::qword_rip("next_promise"), reg("rcx"));
    f.mov(reg("r8"), asm::qword_rip("mbox_w"));
    f.imul(reg("r9"), reg("r8"), imm(24));
    f.lea(reg("r10"), asm::rip("mbox"));
    f.mov(asm::qword_base_index("r10", "r9", 1), reg("rdi"));
    f.mov(asm::qword_base_index_disp("r10", "r9", 1, 8), reg("rsi"));
    f.mov(asm::qword_base_index_disp("r10", "r9", 1, 16), reg("rax"));
    f.unary(Unary::Inc, reg("r8"));
    f.mov(asm::qword_rip("mbox_w"), reg("r8"));
    f.lea(reg("r10"), asm::rip("prom_ready"));
    f.mov(asm::byte_base_index("r10", "rax"), imm(0));
    f.ret();

    f.global("rt_await").label("rt_await");
    prologue(&mut f);
    f.push(reg("rbx"));
    f.push(reg("r12"));
    f.push(reg("r13"));
    f.push(reg("r14"));
    f.push(reg("r15"));
    f.alu(Alu::Sub, reg("rsp"), imm(8));
    f.mov(reg("rbx"), reg("rdi"));
    f.label("Lawait");
    f.lea(reg("rcx"), asm::rip("prom_ready"));
    f.cmp(asm::byte_base_index("rcx", "rbx"), imm(0));
    f.jump(Cond::Ne, "Lready");
    f.mov(reg("r12"), asm::qword_rip("mbox_r"));
    f.cmp(reg("r12"), asm::qword_rip("mbox_w"));
    f.jump(Cond::Ae, "Ldeadlock");
    f.imul(reg("r11"), reg("r12"), imm(24));
    f.lea(reg("rcx"), asm::rip("mbox"));
    f.mov(reg("r13"), asm::qword_base_index("rcx", "r11", 1));
    f.mov(reg("r14"), asm::qword_base_index_disp("rcx", "r11", 1, 8));
    f.mov(reg("r15"), asm::qword_base_index_disp("rcx", "r11", 1, 16));
    f.unary(Unary::Inc, reg("r12"));
    f.mov(asm::qword_rip("mbox_r"), reg("r12"));
    f.lea(reg("rax"), asm::rip("lang_methods"));
    f.mov(reg("rax"), asm::qword_base_index("rax", "r14", 8));
    f.mov(reg("rdi"), reg("r13"));
    f.call_reg("rax");
    f.lea(reg("rcx"), asm::rip("prom_val"));
    f.mov(asm::qword_base_index("rcx", "r15", 8), reg("rax"));
    f.lea(reg("rcx"), asm::rip("prom_ready"));
    f.mov(asm::byte_base_index("rcx", "r15"), imm(1));
    f.jump(Cond::Always, "Lawait");
    f.label("Lready");
    f.lea(reg("rcx"), asm::rip("prom_val"));
    f.mov(reg("rax"), asm::qword_base_index("rcx", "rbx", 8));
    f.alu(Alu::Add, reg("rsp"), imm(8));
    f.pop(reg("r15"));
    f.pop(reg("r14"));
    f.pop(reg("r13"));
    f.pop(reg("r12"));
    f.pop(reg("rbx"));
    f.pop(reg("rbp"));
    f.ret();
    f.label("Ldeadlock");
    f.mov(reg("rdi"), imm(2));
    f.call("hal_exit");

    f.global("rt_write").label("rt_write");
    f.jump(Cond::Always, "hal_write");

    f.global("rt_write_i64").label("rt_write_i64");
    prologue(&mut f);
    f.alu(Alu::Sub, reg("rsp"), imm(32));
    f.mov(reg("rax"), reg("rdi"));
    f.lea(reg("rcx"), asm::base_disp("rbp", -1));
    f.mov(reg("r8"), imm(10));
    f.cmp(reg("rax"), imm(0));
    f.jump(Cond::Ne, "Ldigits");
    f.mov(asm::byte_base("rcx"), imm(48));
    f.mov(reg("rdi"), reg("rcx"));
    f.mov(reg("rsi"), imm(1));
    f.call("rt_write");
    f.jump(Cond::Always, "Lnumdone");
    f.label("Ldigits");
    f.alu(Alu::Xor, reg("rdx"), reg("rdx"));
    f.unary(Unary::Div, reg("r8"));
    f.alu(Alu::Add, reg("rdx"), imm(48));
    f.mov(asm::byte_base("rcx"), reg("dl"));
    f.unary(Unary::Dec, reg("rcx"));
    f.cmp(reg("rax"), imm(0));
    f.jump(Cond::Ne, "Ldigits");
    f.unary(Unary::Inc, reg("rcx"));
    f.mov(reg("rdi"), reg("rcx"));
    f.lea(reg("rsi"), asm::base_disp("rbp", -1));
    f.alu(Alu::Sub, reg("rsi"), reg("rcx"));
    f.unary(Unary::Inc, reg("rsi"));
    f.call("rt_write");
    f.label("Lnumdone");
    f.leave();
    f.ret();

    f.bss().align(8);
    f.label("actors").zero(64 * 64);
    f.label("mbox").zero(128 * 24);
    f.label("prom_val").zero(256 * 8);
    f.label("prom_ready").zero(256);
    f.data().align(8);
    f.label("next_actor").quad(&["0"]);
    f.label("next_promise").quad(&["0"]);
    f.label("mbox_r").quad(&["0"]);
    f.label("mbox_w").quad(&["0"]);
    f.note_gnu_stack();
    f
}

#[cfg(test)]
mod tests {
    use super::{hal_linux_amd64, runtime_amd64};
    use crate::asm::File;
    use crate::link::{Cpu, Os};

    #[test]
    fn arm64_counter_prints_one() {
        let src = include_str!("../examples/01-counter.lang");
        let object = super::arm_object(&crate::parse::parse(src).unwrap()).unwrap();
        let (stderr, code) = crate::arm64::simulate(&object).unwrap();
        assert_eq!(code, 0);
        assert_eq!(stderr, b"count: 1\n");
    }

    #[test]
    fn assembly_round_trips() {
        for file in [hal_linux_amd64(), runtime_amd64()] {
            let text = file.make();
            let again = File::take(&text).expect(&text);
            assert_eq!(again.make(), text);
        }
    }

    #[test]
    fn each_container_has_its_magic() {
        let src = include_str!("../examples/01-counter.lang");
        let checks: &[(Os, Cpu, &[u8])] = &[
            (Os::Linux, Cpu::Amd64, &[0x7F, b'E', b'L', b'F']),
            (Os::Linux, Cpu::Arm64, &[0x7F, b'E', b'L', b'F']),
            (Os::Darwin, Cpu::Amd64, &[0xCF, 0xFA, 0xED, 0xFE]),
            (Os::Darwin, Cpu::Arm64, &[0xCF, 0xFA, 0xED, 0xFE]),
            (Os::Windows, Cpu::Amd64, &[b'M', b'Z']),
            (Os::Windows, Cpu::Arm64, &[b'M', b'Z']),
        ];
        for (os, cpu, magic) in checks {
            let bytes = super::build(src, *os, *cpu, std::path::Path::new("/tmp/the-lang-container-check"));
            bytes.expect("build");
            let file = std::fs::read("/tmp/the-lang-container-check").unwrap();
            assert!(file.starts_with(magic), "{os:?} {cpu:?}");
            if *os == Os::Linux && *cpu == Cpu::Arm64 {
                assert_eq!(u16::from_le_bytes([file[18], file[19]]), 183);
            }
            if *os == Os::Windows {
                assert_eq!(&file[0x80..0x84], b"PE\0\0");
            }
            if *os == Os::Darwin {
                let path = b"/usr/lib/libSystem.B.dylib";
                assert!(file.windows(path.len()).any(|w| w == path));
            }
        }
    }
}
