//! Lower a program to calls on the runtime. The runtime calls the HAL only:
//! `hal_write`, `hal_exit`, `hal_ncpu`, `hal_mem`, `hal_map`, `hal_spawn`, `hal_wait`, `hal_wake`.

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
    params: Vec<String>,
    fields: Vec<String>,
    body: Vec<Stmt>,
}

fn prepare(program: &Program) -> Result<(Vec<(String, Vec<String>)>, Vec<MethodRef>, Vec<Stmt>), String> {
    let mut methods = Vec::new();
    let mut actors: Vec<(String, Vec<String>)> = Vec::new();
    let mut main = None;
    for item in &program.items {
        match item {
            Item::Actor { name, fields, .. } => {
                actors.push((name.clone(), fields.iter().map(|f| f.name.clone()).collect()));
            }
            Item::Method { type_name, name, params, body } => {
                let fields = actors
                    .iter()
                    .find(|(n, _)| n == type_name)
                    .map(|(_, f)| f.clone())
                    .unwrap_or_default();
                methods.push(MethodRef {
                    type_name: type_name.clone(),
                    name: name.clone(),
                    params: params.clone(),
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
    Ok((actors, methods, main))
}

fn lower(program: &Program) -> Result<File, String> {
    let (actors, methods, main) = prepare(program)?;
    let mut file = File::new();
    file.intel().text();
    let mut names = Vec::new();
    for i in 0..methods.len() {
        let symbol = format!("m{i}_{}_{}", methods[i].type_name, methods[i].name);
        names.push(symbol.clone());
        let body = methods[i].body.clone();
        let fields = methods[i].fields.clone();
        let params = methods[i].params.clone();
        file.global(&symbol).label(&symbol);
        prologue(&mut file);
        file.push(reg("rbx"))
            .push(reg("r12"))
            .push(reg("r13"))
            .push(reg("r14"))
            .push(reg("r15"))
            .mov(reg("rbx"), reg("rdi"))
            .mov(reg("r12"), reg("rsi"))
            .mov(reg("r13"), reg("rdx"));
        lower_method(&body, &fields, &params, &methods, &mut file)?;
        file.pop(reg("r15")).pop(reg("r14")).pop(reg("r13")).pop(reg("r12")).pop(reg("rbx")).pop(reg("rbp")).ret();
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

fn lower_method(
    body: &[Stmt],
    fields: &[String],
    params: &[String],
    methods: &[MethodRef],
    file: &mut File,
) -> Result<(), String> {
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
                if recv != "self" || !matches!(op, AssignOp::PlusEq) {
                    return Err("build: this port lowers self.field +=".into());
                }
                let slot = field_slot(fields, name)? as i64;
                if let Some(call) = unwrap_await(value) {
                    emit_param_await(&call, params, methods, file)?;
                    file.push(reg("rax"))
                        .mov(reg("rdi"), reg("rbx"))
                        .mov(reg("rsi"), imm(slot))
                        .call("rt_load")
                        .pop(reg("rcx"))
                        .alu(Alu::Add, reg("rax"), reg("rcx"));
                } else {
                    file.mov(reg("rdi"), reg("rbx")).mov(reg("rsi"), imm(slot)).call("rt_load");
                    match value {
                        Expr::Int(n) => {
                            file.alu(Alu::Add, reg("rax"), imm(*n));
                        }
                        Expr::Name(param) => {
                            file.alu(Alu::Add, reg("rax"), reg(param_reg(params, param)?));
                        }
                        _ => return Err("build: self.field += needs an integer or an awaited call".into()),
                    }
                }
                file.mov(reg("rdx"), reg("rax"))
                    .mov(reg("rdi"), reg("rbx"))
                    .mov(reg("rsi"), imm(slot))
                    .call("rt_store");
            }
            Stmt::Expr(Expr::Field { base, name }) => {
                let Expr::Name(recv) = base.as_ref() else {
                    return Err("build: a method returns self.field".into());
                };
                if recv != "self" {
                    return Err("build: a method returns self.field".into());
                }
                let slot = field_slot(fields, name)? as i64;
                file.mov(reg("rdi"), reg("rbx")).mov(reg("rsi"), imm(slot)).call("rt_load");
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
    if let Some(call) = unwrap_await(expr) {
        emit_await(call.actor, call.method, call.args, locals, methods, file)?;
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
                    let Some(call) = unwrap_await(arg) else {
                        return Err("build: fprintf %d needs an awaited call".into());
                    };
                    emit_await(call.actor, call.method, call.args, locals, methods, file)?;
                    file.mov(reg("rdi"), reg("rax")).call("rt_write_i64");
                }
            }
        }
        return Ok(());
    }
    Err("build: unsupported statement in main".into())
}

fn emit_await(
    local: &str,
    method: &str,
    args: &[Expr],
    locals: &[String],
    methods: &[MethodRef],
    file: &mut File,
) -> Result<(), String> {
    let slot = locals.iter().position(|n| n == local).ok_or_else(|| format!("build: no local {local}"))?;
    let id = method_id(methods, method)?;
    let off = (slot as i64 + 1) * 8;
    if args.len() > 2 {
        return Err("build: a call takes two arguments".into());
    }
    file.mov(reg("rdi"), asm::qword_base_disp("rbp", -off)).mov(reg("rsi"), imm(id as i64));
    emit_arg(file, "rdx", args.first(), ArgPlace::Local(locals))?;
    emit_arg(file, "rcx", args.get(1), ArgPlace::Local(locals))?;
    file.call("rt_call").mov(reg("rdi"), reg("rax")).call("rt_await");
    Ok(())
}

fn emit_param_await(call: &Invoke<'_>, params: &[String], methods: &[MethodRef], file: &mut File) -> Result<(), String> {
    if call.args.len() > 2 {
        return Err("build: a call takes two arguments".into());
    }
    let id = method_id(methods, call.method)?;
    file.mov(reg("rdi"), reg(param_reg(params, call.actor)?)).mov(reg("rsi"), imm(id as i64));
    emit_arg(file, "rdx", call.args.first(), ArgPlace::Param(params))?;
    emit_arg(file, "rcx", call.args.get(1), ArgPlace::Param(params))?;
    file.call("rt_call").mov(reg("rdi"), reg("rax")).call("rt_await");
    Ok(())
}

enum ArgPlace<'a> {
    Local(&'a [String]),
    Param(&'a [String]),
}

fn emit_arg(file: &mut File, dst: &str, arg: Option<&Expr>, place: ArgPlace<'_>) -> Result<(), String> {
    match arg {
        None => {
            file.alu(Alu::Xor, reg(dst), reg(dst));
        }
        Some(Expr::Int(n)) => {
            file.mov(reg(dst), imm(*n));
        }
        Some(Expr::Name(name)) => match place {
            ArgPlace::Param(params) => {
                file.mov(reg(dst), reg(param_reg(params, name)?));
            }
            ArgPlace::Local(locals) => {
                let slot = locals.iter().position(|n| n == name).ok_or_else(|| format!("build: no local {name}"))?;
                let off = (slot as i64 + 1) * 8;
                file.mov(reg(dst), asm::qword_base_disp("rbp", -off));
            }
        },
        Some(_) => return Err("build: a call argument is a local or an integer".into()),
    }
    Ok(())
}

fn param_reg(params: &[String], name: &str) -> Result<&'static str, String> {
    match params.iter().position(|p| p == name) {
        Some(0) => Ok("r12"),
        Some(1) => Ok("r13"),
        Some(_) => Err("build: a call takes two arguments".into()),
        None => Err(format!("build: no value {name}")),
    }
}

fn method_id(methods: &[MethodRef], name: &str) -> Result<usize, String> {
    let mut found = None;
    for (i, method) in methods.iter().enumerate() {
        if method.name == name {
            if found.is_some() {
                return Err(format!("build: method {name} is ambiguous"));
            }
            found = Some(i);
        }
    }
    found.ok_or_else(|| format!("build: no method {name}"))
}

struct Invoke<'a> {
    actor: &'a str,
    method: &'a str,
    args: &'a [Expr],
}

fn unwrap_await(expr: &Expr) -> Option<Invoke<'_>> {
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
    let Expr::Field { base, name } = callee.as_ref() else {
        return None;
    };
    let Expr::Name(actor) = base.as_ref() else {
        return None;
    };
    Some(Invoke { actor, method: name, args })
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
    use crate::arm64::{Arg, Main, Step};
    let (actors, methods, main) = prepare(program)?;
    let mut lowered = Vec::new();
    for method in &methods {
        let mut steps = Vec::new();
        for stmt in &method.body {
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
                    let slot = field_slot(&method.fields, name)? as u16;
                    if let Some(call) = unwrap_await(value) {
                        if call.args.len() > 1 {
                            return Err("build: a method passes one argument".into());
                        }
                        let actor = param_index(&method.params, call.actor)?;
                        let id = method_id(&methods, call.method)? as u16;
                        let arg = match call.args.first() {
                            None => Arg::None,
                            Some(Expr::Int(n)) => Arg::Imm(fit_u16(*n)?),
                            Some(Expr::Name(n)) => Arg::Param(param_index(&method.params, n)?),
                            Some(_) => return Err("build: a call argument is a local or an integer".into()),
                        };
                        steps.push(Step::AddAwait { slot, actor, method: id, arg });
                        steps.push(Step::Store(slot));
                    } else {
                        steps.push(Step::Load(slot));
                        match value {
                            Expr::Int(n) => steps.push(Step::Add(fit_u16(*n)?)),
                            Expr::Name(n) => steps.push(Step::AddArg(param_index(&method.params, n)?)),
                            _ => return Err("build: self.field += needs an integer or an awaited call".into()),
                        }
                        steps.push(Step::Store(slot));
                    }
                }
                Stmt::Expr(Expr::Field { base, name }) => {
                    let Expr::Name(recv) = base.as_ref() else {
                        return Err("build: a method returns self.field".into());
                    };
                    if recv != "self" {
                        return Err("build: a method returns self.field".into());
                    }
                    steps.push(Step::Load(field_slot(&method.fields, name)? as u16));
                }
                Stmt::Expr(_) => return Err("build: this port lowers self.field in a method".into()),
            }
        }
        lowered.push(crate::arm64::Method { name: format!("{}_{}", method.type_name, method.name), steps });
    }
    let mut ops = Vec::new();
    let mut locals = Vec::new();
    for stmt in &main {
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
                ops.push(Main::Actor(locals.len() as u16));
                locals.push(name.clone());
            }
            Stmt::Expr(expr) => {
                if let Some(call) = unwrap_await(expr) {
                    ops.push(arm_await(&call, &locals, &methods, false)?);
                } else if let Expr::Call { args, .. } = expr {
                    let Expr::Str(fmt) = &args[1] else { return Err("build: fprintf format must be a literal".into()) };
                    let mut rest = args[2..].iter();
                    for piece in format_pieces(fmt)? {
                        match piece {
                            Piece::Text(text) => ops.push(Main::Write(text.into_bytes())),
                            Piece::Int => {
                                let arg = rest.next().ok_or("build: fprintf is missing a value")?;
                                let Some(call) = unwrap_await(arg) else {
                                    return Err("build: fprintf %d needs an awaited call".into());
                                };
                                ops.push(arm_await(&call, &locals, &methods, true)?);
                            }
                        }
                    }
                } else {
                    return Err("build: unsupported statement in main".into());
                }
            }
        }
    }
    crate::arm64::assemble(&lowered, &ops)
}

fn arm_await(call: &Invoke<'_>, locals: &[String], methods: &[MethodRef], print: bool) -> Result<crate::arm64::Main, String> {
    use crate::arm64::{Arg, Main};
    if call.args.len() > 2 {
        return Err("build: a call takes two arguments".into());
    }
    let slot = locals.iter().position(|n| n == call.actor).ok_or_else(|| format!("build: no local {}", call.actor))?;
    let id = method_id(methods, call.method)? as u16;
    let mut args = [Arg::None, Arg::None];
    for (i, arg) in call.args.iter().enumerate() {
        args[i] = match arg {
            Expr::Int(n) => Arg::Imm(fit_u16(*n)?),
            Expr::Name(name) => {
                let local = locals.iter().position(|n| n == name).ok_or_else(|| format!("build: no local {name}"))?;
                Arg::Local(local as u8)
            }
            _ => return Err("build: a call argument is a local or an integer".into()),
        };
    }
    Ok(Main::Await { slot: slot as u16, method: id, arg0: args[0], arg1: args[1], print })
}

fn param_index(params: &[String], name: &str) -> Result<u8, String> {
    match params.iter().position(|p| p == name) {
        Some(i) if i < 2 => Ok(i as u8),
        Some(_) => Err("build: a call takes two arguments".into()),
        None => Err(format!("build: no value {name}")),
    }
}

fn fit_u16(n: i64) -> Result<u16, String> {
    u16::try_from(n).map_err(|_| "build: integer does not fit this port".into())
}

fn field_slot(fields: &[String], name: &str) -> Result<usize, String> {
    fields.iter().position(|n| n == name).ok_or_else(|| format!("build: no field {name}"))
}

fn hal_linux_amd64() -> File {
    let mut f = File::new();
    f.intel().text();

    // rdi = pointer, rsi = length. Writes to stderr.
    f.global("hal_write").label("hal_write");
    f.mov(reg("rdx"), reg("rsi"));
    f.mov(reg("rsi"), reg("rdi"));
    f.mov(reg("rdi"), imm(2));
    f.mov(reg("rax"), imm(1));
    f.syscall();
    f.ret();

    // rdi = status. Ends every scheduler thread.
    f.global("hal_exit").label("hal_exit");
    f.mov(reg("rax"), imm(231));
    f.syscall();

    // rax = number of CPUs in this process's affinity. At least 1.
    f.global("hal_ncpu").label("hal_ncpu");
    f.alu(Alu::Xor, reg("rdi"), reg("rdi"));
    f.mov(reg("rsi"), imm(128));
    f.lea(reg("rdx"), asm::rip("hal_cpu_mask"));
    f.mov(reg("rax"), imm(204));
    f.syscall();
    f.mov(reg("r11"), reg("rax"));
    f.mov(reg("rcx"), reg("rax"));
    f.alu(Alu::Shr, reg("rcx"), imm(63));
    f.cmp(reg("rcx"), imm(0));
    f.jump(Cond::Ne, "HalNcpuOne");
    f.cmp(reg("r11"), imm(0));
    f.jump(Cond::E, "HalNcpuOne");
    f.lea(reg("r9"), asm::rip("hal_cpu_mask"));
    f.alu(Alu::Xor, reg("rcx"), reg("rcx"));
    f.alu(Alu::Xor, reg("rsi"), reg("rsi"));
    f.label("HalNcpuByte");
    f.cmp(reg("rsi"), reg("r11"));
    f.jump(Cond::Ae, "HalNcpuDone");
    f.alu(Alu::Xor, reg("eax"), reg("eax"));
    f.mov(reg("al"), asm::byte_base_index("r9", "rsi"));
    f.label("HalNcpuBit");
    f.cmp(reg("eax"), imm(0));
    f.jump(Cond::E, "HalNcpuNext");
    f.alu(Alu::Shr, reg("eax"), imm(1));
    f.jump(Cond::Ae, "HalNcpuNo");
    f.unary(Unary::Inc, reg("rcx"));
    f.label("HalNcpuNo");
    f.cmp(reg("eax"), imm(0));
    f.jump(Cond::Ne, "HalNcpuBit");
    f.label("HalNcpuNext");
    f.unary(Unary::Inc, reg("rsi"));
    f.jump(Cond::Always, "HalNcpuByte");
    f.label("HalNcpuDone");
    f.cmp(reg("rcx"), imm(0));
    f.jump(Cond::E, "HalNcpuOne");
    f.mov(reg("rax"), reg("rcx"));
    f.ret();
    f.label("HalNcpuOne");
    f.mov(reg("rax"), imm(1));
    f.ret();

    // rdi = length. rax = address, or a negative errno.
    f.global("hal_map").label("hal_map");
    f.mov(reg("rsi"), reg("rdi"));
    f.alu(Alu::Xor, reg("rdi"), reg("rdi"));
    f.mov(reg("rdx"), imm(3));
    f.mov(reg("r10"), imm(0x22));
    f.mov(reg("r8"), imm(-1));
    f.alu(Alu::Xor, reg("r9"), reg("r9"));
    f.mov(reg("rax"), imm(9));
    f.syscall();
    f.ret();

    // rdi = child stack top, rsi = entry. Child jumps to the entry. Parent returns the tid.
    f.global("hal_spawn").label("hal_spawn");
    f.push(reg("r12"));
    f.mov(reg("r12"), reg("rsi"));
    f.mov(reg("rsi"), reg("rdi"));
    f.mov(reg("rdi"), imm(0x50f00));
    f.alu(Alu::Xor, reg("rdx"), reg("rdx"));
    f.alu(Alu::Xor, reg("r10"), reg("r10"));
    f.alu(Alu::Xor, reg("r8"), reg("r8"));
    f.mov(reg("rax"), imm(56));
    f.syscall();
    f.cmp(reg("rax"), imm(0));
    f.jump(Cond::Ne, "HalSpawnParent");
    f.call_reg("r12");
    f.label("HalSpawnParent");
    f.pop(reg("r12"));
    f.ret();

    // rdi = word, rsi = expected value. Returns when the word changes.
    f.global("hal_wait").label("hal_wait");
    f.mov(reg("rdx"), reg("rsi"));
    f.alu(Alu::Xor, reg("rsi"), reg("rsi"));
    f.alu(Alu::Xor, reg("r10"), reg("r10"));
    f.mov(reg("rax"), imm(202));
    f.syscall();
    f.ret();

    // rdi = word. Wakes one waiter.
    f.global("hal_wake").label("hal_wake");
    f.mov(reg("rsi"), imm(1));
    f.mov(reg("rdx"), imm(1));
    f.alu(Alu::Xor, reg("r10"), reg("r10"));
    f.mov(reg("rax"), imm(202));
    f.syscall();
    f.ret();

    emit_hal_mem(&mut f);
    f.rodata();
    f.label("HalCgFile").ascii(b"/proc/self/cgroup\0");
    f.label("HalInfoFile").ascii(b"/proc/meminfo\0");
    f.label("HalRoot").ascii(b"/sys/fs/cgroup\0");
    f.label("HalCur").ascii(b"/memory.current\0");
    f.label("HalMax").ascii(b"/memory.max\0");
    f.label("HalHigh").ascii(b"/memory.high\0");
    f.label("HalAvail").ascii(b"MemAvailable:\0");
    f.bss().label("hal_cpu_mask").zero(128);
    f.label("hal_buf").zero(4096);
    f.label("hal_path").zero(1024);
    f.label("hal_file").zero(1024);
    f.note_gnu_stack();
    f
}

fn emit_hal_mem(f: &mut File) {
    // rax = bytes still available. Cgroup v2 limit minus usage, else MemAvailable.
    f.global("hal_mem").label("hal_mem");
    f.push(reg("r12"));
    f.push(reg("r13"));
    f.push(reg("r14"));
    f.push(reg("r15"));
    f.mov(reg("r12"), imm(-1));
    f.mov(reg("r15"), imm(64));
    f.lea(reg("rdi"), asm::rip("HalCgFile"));
    f.call("hal_read");
    f.cmp(reg("rax"), imm(-1));
    f.jump(Cond::E, "HalMemHost");
    f.lea(reg("r8"), asm::rip("hal_buf"));
    f.label("HalFind");
    f.mov(reg("al"), asm::byte_base("r8"));
    f.cmp(reg("al"), imm(0));
    f.jump(Cond::E, "HalMemHost");
    f.cmp(reg("al"), imm(b'0' as i64));
    f.jump(Cond::Ne, "HalFindNext");
    f.mov(reg("rsi"), reg("r8"));
    f.unary(Unary::Inc, reg("rsi"));
    f.mov(reg("al"), asm::byte_base("rsi"));
    f.cmp(reg("al"), imm(b':' as i64));
    f.jump(Cond::Ne, "HalFindNext");
    f.unary(Unary::Inc, reg("rsi"));
    f.mov(reg("al"), asm::byte_base("rsi"));
    f.cmp(reg("al"), imm(b':' as i64));
    f.jump(Cond::Ne, "HalFindNext");
    f.unary(Unary::Inc, reg("rsi"));
    f.mov(reg("r8"), reg("rsi"));
    f.jump(Cond::Always, "HalFound");
    f.label("HalFindNext");
    f.unary(Unary::Inc, reg("r8"));
    f.jump(Cond::Always, "HalFind");
    f.label("HalFound");
    f.lea(reg("rsi"), asm::rip("HalRoot"));
    f.lea(reg("rdi"), asm::rip("hal_path"));
    f.call("hal_copy");
    f.label("HalPathCopy");
    f.mov(reg("al"), asm::byte_base("r8"));
    f.cmp(reg("al"), imm(10));
    f.jump(Cond::E, "HalPathEnd");
    f.cmp(reg("al"), imm(0));
    f.jump(Cond::E, "HalPathEnd");
    f.lea(reg("r10"), asm::rip("hal_path"));
    f.alu(Alu::Add, reg("r10"), imm(1000));
    f.cmp(reg("rdi"), reg("r10"));
    f.jump(Cond::Ae, "HalPathEnd");
    f.mov(asm::byte_base("rdi"), reg("al"));
    f.unary(Unary::Inc, reg("r8"));
    f.unary(Unary::Inc, reg("rdi"));
    f.jump(Cond::Always, "HalPathCopy");
    f.label("HalPathEnd");
    f.mov(asm::byte_base("rdi"), imm(0));
    f.lea(reg("rdi"), asm::rip("HalRoot"));
    f.call("hal_strlen");
    f.mov(reg("r14"), reg("rax"));
    f.label("HalWalk");
    f.cmp(reg("r15"), imm(0));
    f.jump(Cond::E, "HalAfterWalk");
    f.unary(Unary::Dec, reg("r15"));
    f.call("hal_cgdir");
    f.lea(reg("rdi"), asm::rip("hal_path"));
    f.call("hal_strlen");
    f.cmp(reg("r14"), reg("rax"));
    f.jump(Cond::B, "HalStripGo");
    f.jump(Cond::Always, "HalAfterWalk");
    f.label("HalStripGo");
    f.call("hal_strip");
    f.jump(Cond::Always, "HalWalk");
    f.label("HalAfterWalk");
    f.call("hal_memavail");
    f.cmp(reg("r12"), imm(-1));
    f.jump(Cond::Ne, "HalCap");
    f.cmp(reg("rax"), imm(-1));
    f.jump(Cond::Ne, "HalMemOut");
    f.alu(Alu::Xor, reg("rax"), reg("rax"));
    f.jump(Cond::Always, "HalMemOut");
    f.label("HalCap");
    f.cmp(reg("rax"), imm(-1));
    f.jump(Cond::E, "HalUseBest");
    f.cmp(reg("rax"), reg("r12"));
    f.jump(Cond::Ae, "HalUseBest");
    f.mov(reg("r12"), reg("rax"));
    f.label("HalUseBest");
    f.mov(reg("rax"), reg("r12"));
    f.jump(Cond::Always, "HalMemOut");
    f.label("HalMemHost");
    f.call("hal_memavail");
    f.cmp(reg("rax"), imm(-1));
    f.jump(Cond::Ne, "HalMemOut");
    f.alu(Alu::Xor, reg("rax"), reg("rax"));
    f.label("HalMemOut");
    f.pop(reg("r15"));
    f.pop(reg("r14"));
    f.pop(reg("r13"));
    f.pop(reg("r12"));
    f.ret();

    f.global("hal_cgdir").label("hal_cgdir");
    f.lea(reg("rdi"), asm::rip("HalCur"));
    f.call("hal_join");
    f.lea(reg("rdi"), asm::rip("hal_file"));
    f.call("hal_read");
    f.cmp(reg("rax"), imm(-1));
    f.jump(Cond::E, "HalCgRet");
    f.lea(reg("rdi"), asm::rip("hal_buf"));
    f.call("hal_u64");
    f.cmp(reg("rax"), imm(-1));
    f.jump(Cond::E, "HalCgRet");
    f.mov(reg("r13"), reg("rax"));
    f.lea(reg("rdi"), asm::rip("HalMax"));
    f.call("hal_limit");
    f.lea(reg("rdi"), asm::rip("HalHigh"));
    f.call("hal_limit");
    f.label("HalCgRet");
    f.ret();

    f.global("hal_limit").label("hal_limit");
    f.call("hal_join");
    f.lea(reg("rdi"), asm::rip("hal_file"));
    f.call("hal_read");
    f.cmp(reg("rax"), imm(-1));
    f.jump(Cond::E, "HalLimRet");
    f.lea(reg("rdi"), asm::rip("hal_buf"));
    f.call("hal_u64");
    f.cmp(reg("rax"), imm(-1));
    f.jump(Cond::E, "HalLimRet");
    f.cmp(reg("r13"), reg("rax"));
    f.jump(Cond::B, "HalLimSub");
    f.alu(Alu::Xor, reg("rax"), reg("rax"));
    f.jump(Cond::Always, "HalLimHave");
    f.label("HalLimSub");
    f.alu(Alu::Sub, reg("rax"), reg("r13"));
    f.label("HalLimHave");
    f.cmp(reg("rax"), reg("r12"));
    f.jump(Cond::Ae, "HalLimRet");
    f.mov(reg("r12"), reg("rax"));
    f.label("HalLimRet");
    f.ret();

    f.global("hal_memavail").label("hal_memavail");
    f.lea(reg("rdi"), asm::rip("HalInfoFile"));
    f.call("hal_read");
    f.cmp(reg("rax"), imm(-1));
    f.jump(Cond::E, "HalInfoFail");
    f.lea(reg("r8"), asm::rip("hal_buf"));
    f.label("HalInfoScan");
    f.mov(reg("al"), asm::byte_base("r8"));
    f.cmp(reg("al"), imm(0));
    f.jump(Cond::E, "HalInfoFail");
    f.mov(reg("rdi"), reg("r8"));
    f.lea(reg("rsi"), asm::rip("HalAvail"));
    f.call("hal_prefix");
    f.cmp(reg("rax"), imm(1));
    f.jump(Cond::E, "HalInfoParse");
    f.unary(Unary::Inc, reg("r8"));
    f.jump(Cond::Always, "HalInfoScan");
    f.label("HalInfoParse");
    f.alu(Alu::Add, reg("r8"), imm(13));
    f.mov(reg("rdi"), reg("r8"));
    f.call("hal_u64");
    f.cmp(reg("rax"), imm(-1));
    f.jump(Cond::E, "HalInfoFail");
    f.alu(Alu::Shl, reg("rax"), imm(10));
    f.ret();
    f.label("HalInfoFail");
    f.mov(reg("rax"), imm(-1));
    f.ret();

    f.global("hal_read").label("hal_read");
    f.mov(reg("r9"), reg("rdi"));
    f.mov(reg("rdi"), imm(-100));
    f.mov(reg("rsi"), reg("r9"));
    f.alu(Alu::Xor, reg("rdx"), reg("rdx"));
    f.alu(Alu::Xor, reg("r10"), reg("r10"));
    f.mov(reg("rax"), imm(257));
    f.syscall();
    f.mov(reg("rcx"), reg("rax"));
    f.alu(Alu::Shr, reg("rcx"), imm(63));
    f.cmp(reg("rcx"), imm(0));
    f.jump(Cond::Ne, "HalReadFail");
    f.mov(reg("r8"), reg("rax"));
    f.mov(reg("rdi"), reg("r8"));
    f.lea(reg("rsi"), asm::rip("hal_buf"));
    f.mov(reg("rdx"), imm(4095));
    f.alu(Alu::Xor, reg("rax"), reg("rax"));
    f.syscall();
    f.mov(reg("r9"), reg("rax"));
    f.mov(reg("rdi"), reg("r8"));
    f.mov(reg("rax"), imm(3));
    f.syscall();
    f.mov(reg("rcx"), reg("r9"));
    f.alu(Alu::Shr, reg("rcx"), imm(63));
    f.cmp(reg("rcx"), imm(0));
    f.jump(Cond::Ne, "HalReadFail");
    f.lea(reg("rdi"), asm::rip("hal_buf"));
    f.alu(Alu::Add, reg("rdi"), reg("r9"));
    f.mov(asm::byte_base("rdi"), imm(0));
    f.mov(reg("rax"), reg("r9"));
    f.ret();
    f.label("HalReadFail");
    f.mov(reg("rax"), imm(-1));
    f.ret();

    f.global("hal_u64").label("hal_u64");
    f.label("HalU64Skip");
    f.mov(reg("al"), asm::byte_base("rdi"));
    f.cmp(reg("al"), imm(b' ' as i64));
    f.jump(Cond::E, "HalU64Adv");
    f.cmp(reg("al"), imm(9));
    f.jump(Cond::E, "HalU64Adv");
    f.jump(Cond::Always, "HalU64Start");
    f.label("HalU64Adv");
    f.unary(Unary::Inc, reg("rdi"));
    f.jump(Cond::Always, "HalU64Skip");
    f.label("HalU64Start");
    f.cmp(reg("al"), imm(b'0' as i64));
    f.jump(Cond::B, "HalU64None");
    f.cmp(reg("al"), imm(b'9' as i64 + 1));
    f.jump(Cond::Ae, "HalU64None");
    f.alu(Alu::Xor, reg("rax"), reg("rax"));
    f.label("HalU64Dig");
    f.alu(Alu::Xor, reg("edx"), reg("edx"));
    f.mov(reg("dl"), asm::byte_base("rdi"));
    f.cmp(reg("dl"), imm(b'0' as i64));
    f.jump(Cond::B, "HalU64Ok");
    f.cmp(reg("dl"), imm(b'9' as i64 + 1));
    f.jump(Cond::Ae, "HalU64Ok");
    f.mov(reg("rcx"), reg("rax"));
    f.alu(Alu::Shl, reg("rax"), imm(3));
    f.alu(Alu::Shl, reg("rcx"), imm(1));
    f.alu(Alu::Add, reg("rax"), reg("rcx"));
    f.alu(Alu::Sub, reg("edx"), imm(b'0' as i64));
    f.alu(Alu::Add, reg("rax"), reg("rdx"));
    f.unary(Unary::Inc, reg("rdi"));
    f.jump(Cond::Always, "HalU64Dig");
    f.label("HalU64None");
    f.mov(reg("rax"), imm(-1));
    f.ret();
    f.label("HalU64Ok");
    f.ret();

    f.global("hal_copy").label("hal_copy");
    f.label("HalCopyLoop");
    f.mov(reg("al"), asm::byte_base("rsi"));
    f.mov(asm::byte_base("rdi"), reg("al"));
    f.cmp(reg("al"), imm(0));
    f.jump(Cond::E, "HalCopyDone");
    f.unary(Unary::Inc, reg("rsi"));
    f.unary(Unary::Inc, reg("rdi"));
    f.jump(Cond::Always, "HalCopyLoop");
    f.label("HalCopyDone");
    f.ret();

    f.global("hal_join").label("hal_join");
    f.mov(reg("r8"), reg("rdi"));
    f.lea(reg("rsi"), asm::rip("hal_path"));
    f.lea(reg("rdi"), asm::rip("hal_file"));
    f.call("hal_copy");
    f.mov(reg("rsi"), reg("r8"));
    f.call("hal_copy");
    f.ret();

    f.global("hal_strlen").label("hal_strlen");
    f.alu(Alu::Xor, reg("rax"), reg("rax"));
    f.label("HalLen");
    f.cmp(asm::byte_base_index("rdi", "rax"), imm(0));
    f.jump(Cond::E, "HalLenDone");
    f.unary(Unary::Inc, reg("rax"));
    f.jump(Cond::Always, "HalLen");
    f.label("HalLenDone");
    f.ret();

    f.global("hal_strip").label("hal_strip");
    f.lea(reg("rdi"), asm::rip("hal_path"));
    f.call("hal_strlen");
    f.mov(reg("rcx"), reg("rax"));
    f.label("HalStrip");
    f.cmp(reg("rcx"), imm(0));
    f.jump(Cond::E, "HalStripDone");
    f.unary(Unary::Dec, reg("rcx"));
    f.cmp(asm::byte_base_index("rdi", "rcx"), imm(b'/' as i64));
    f.jump(Cond::Ne, "HalStrip");
    f.mov(asm::byte_base_index("rdi", "rcx"), imm(0));
    f.label("HalStripDone");
    f.ret();

    f.global("hal_prefix").label("hal_prefix");
    f.label("HalPre");
    f.mov(reg("al"), asm::byte_base("rsi"));
    f.cmp(reg("al"), imm(0));
    f.jump(Cond::E, "HalPreYes");
    f.cmp(reg("al"), asm::byte_base("rdi"));
    f.jump(Cond::Ne, "HalPreNo");
    f.unary(Unary::Inc, reg("rsi"));
    f.unary(Unary::Inc, reg("rdi"));
    f.jump(Cond::Always, "HalPre");
    f.label("HalPreYes");
    f.mov(reg("rax"), imm(1));
    f.ret();
    f.label("HalPreNo");
    f.alu(Alu::Xor, reg("rax"), reg("rax"));
    f.ret();
}

fn runtime_amd64() -> File {
    let mut f = File::new();
    f.intel().text();
    f.global("_start").label("_start");
    f.alu(Alu::Xor, reg("rbp"), reg("rbp"));
    f.mov(reg("rbx"), imm(-1));
    f.call("spawn_ms");
    f.call("rt_start");
    f.mov(reg("rdi"), reg("rax"));
    f.call("hal_exit");

    f.global("rt_start").label("rt_start");
    f.call("hal_mem");
    f.mov(asm::qword_rip("mem_avail"), reg("rax"));
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

    f.global("spawn_ms").label("spawn_ms");
    f.call("hal_ncpu");
    f.mov(reg("r12"), reg("rax"));
    f.cmp(reg("r12"), imm(2));
    f.jump(Cond::B, "MsDone");
    f.unary(Unary::Dec, reg("r12"));
    f.mov(reg("rdi"), reg("r12"));
    f.alu(Alu::Shl, reg("rdi"), imm(13));
    f.call("hal_map");
    f.mov(reg("rcx"), reg("rax"));
    f.alu(Alu::Shr, reg("rcx"), imm(63));
    f.cmp(reg("rcx"), imm(0));
    f.jump(Cond::Ne, "MsDone");
    f.mov(reg("r15"), reg("rax"));
    f.label("MsSpawn");
    f.cmp(reg("r12"), imm(0));
    f.jump(Cond::E, "MsDone");
    f.unary(Unary::Dec, reg("r12"));
    f.mov(reg("rax"), reg("r12"));
    f.alu(Alu::Shl, reg("rax"), imm(13));
    f.mov(reg("rsi"), reg("r15"));
    f.alu(Alu::Add, reg("rsi"), reg("rax"));
    f.lea(reg("rsi"), asm::base_disp("rsi", 8192));
    f.mov(reg("rdi"), reg("rsi"));
    f.lea(reg("rsi"), asm::rip("m_idle"));
    f.call("hal_spawn");
    f.jump(Cond::Always, "MsSpawn");
    f.label("MsDone");
    f.ret();

    f.global("m_idle").label("m_idle");
    f.alu(Alu::Sub, reg("rsp"), imm(8));
    f.label("Mloop");
    f.call("rt_lock");
    f.call("rt_find");
    f.cmp(reg("rax"), imm(0));
    f.jump(Cond::E, "Msleep");
    f.call("rt_unlock");
    f.call("rt_exec");
    f.jump(Cond::Always, "Mloop");
    f.label("Msleep");
    f.mov(reg("r13"), asm::qword_rip("sched_seq"));
    f.call("rt_unlock");
    f.call("rt_park");
    f.jump(Cond::Always, "Mloop");

    f.global("rt_lock").label("rt_lock");
    f.push(reg("rax"));
    f.label("Lspin");
    f.mov(reg("rax"), imm(1));
    f.xchg(reg("rax"), asm::qword_rip("sched_lock"));
    f.cmp(reg("rax"), imm(0));
    f.jump(Cond::Ne, "Lspin");
    f.pop(reg("rax"));
    f.ret();

    f.global("rt_unlock").label("rt_unlock");
    f.mov(asm::qword_rip("sched_lock"), imm(0));
    f.ret();

    f.global("rt_wake").label("rt_wake");
    f.lea(reg("rdi"), asm::rip("sched_seq"));
    f.jump(Cond::Always, "hal_wake");

    f.global("rt_park").label("rt_park");
    f.lea(reg("rdi"), asm::rip("sched_seq"));
    f.mov(reg("rsi"), reg("r13"));
    f.jump(Cond::Always, "hal_wait");

    f.global("rt_bump").label("rt_bump");
    f.mov(reg("rcx"), asm::qword_rip("sched_seq"));
    f.alu(Alu::Add, reg("rcx"), imm(1));
    f.mov(asm::qword_rip("sched_seq"), reg("rcx"));
    f.ret();

    f.global("rt_find").label("rt_find");
    f.mov(reg("r10"), asm::qword_rip("mbox_w"));
    f.alu(Alu::Sub, reg("r10"), asm::qword_rip("mbox_r"));
    f.label("Ftry");
    f.cmp(reg("r10"), imm(0));
    f.jump(Cond::E, "Fnone");
    f.mov(reg("rsi"), asm::qword_rip("mbox_r"));
    f.imul(reg("rdi"), reg("rsi"), imm(40));
    f.lea(reg("rcx"), asm::rip("mbox"));
    f.mov(reg("r13"), asm::qword_base_index("rcx", "rdi", 1));
    f.lea(reg("rdx"), asm::rip("actor_state"));
    f.cmp(asm::byte_base_index("rdx", "r13"), imm(0));
    f.jump(Cond::E, "Fpop");
    f.mov(reg("rsi"), asm::qword_rip("mbox_w"));
    f.imul(reg("rsi"), reg("rsi"), imm(40));
    f.mov(reg("rax"), asm::qword_base_index("rcx", "rdi", 1));
    f.mov(asm::qword_base_index("rcx", "rsi", 1), reg("rax"));
    f.mov(reg("rax"), asm::qword_base_index_disp("rcx", "rdi", 1, 8));
    f.mov(asm::qword_base_index_disp("rcx", "rsi", 1, 8), reg("rax"));
    f.mov(reg("rax"), asm::qword_base_index_disp("rcx", "rdi", 1, 16));
    f.mov(asm::qword_base_index_disp("rcx", "rsi", 1, 16), reg("rax"));
    f.mov(reg("rax"), asm::qword_base_index_disp("rcx", "rdi", 1, 24));
    f.mov(asm::qword_base_index_disp("rcx", "rsi", 1, 24), reg("rax"));
    f.mov(reg("rax"), asm::qword_base_index_disp("rcx", "rdi", 1, 32));
    f.mov(asm::qword_base_index_disp("rcx", "rsi", 1, 32), reg("rax"));
    f.mov(reg("rax"), asm::qword_rip("mbox_r"));
    f.alu(Alu::Add, reg("rax"), imm(1));
    f.mov(asm::qword_rip("mbox_r"), reg("rax"));
    f.mov(reg("rax"), asm::qword_rip("mbox_w"));
    f.alu(Alu::Add, reg("rax"), imm(1));
    f.mov(asm::qword_rip("mbox_w"), reg("rax"));
    f.unary(Unary::Dec, reg("r10"));
    f.jump(Cond::Always, "Ftry");
    f.label("Fpop");
    f.mov(reg("r14"), asm::qword_base_index_disp("rcx", "rdi", 1, 8));
    f.mov(reg("r15"), asm::qword_base_index_disp("rcx", "rdi", 1, 16));
    f.mov(reg("r8"), asm::qword_base_index_disp("rcx", "rdi", 1, 24));
    f.mov(reg("r9"), asm::qword_base_index_disp("rcx", "rdi", 1, 32));
    f.mov(reg("rax"), asm::qword_rip("mbox_r"));
    f.alu(Alu::Add, reg("rax"), imm(1));
    f.mov(asm::qword_rip("mbox_r"), reg("rax"));
    f.mov(asm::byte_base_index("rdx", "r13"), imm(1));
    f.mov(reg("rax"), asm::qword_rip("running"));
    f.alu(Alu::Add, reg("rax"), imm(1));
    f.mov(asm::qword_rip("running"), reg("rax"));
    f.mov(reg("rax"), imm(1));
    f.ret();
    f.label("Fnone");
    f.alu(Alu::Xor, reg("rax"), reg("rax"));
    f.ret();

    f.global("rt_exec").label("rt_exec");
    f.lea(reg("rax"), asm::rip("lang_methods"));
    f.mov(reg("rax"), asm::qword_base_index("rax", "r14", 8));
    f.mov(reg("rdi"), reg("r13"));
    f.mov(reg("rsi"), reg("r8"));
    f.mov(reg("rdx"), reg("r9"));
    f.call_reg("rax");
    f.call("rt_lock");
    f.lea(reg("rcx"), asm::rip("prom_val"));
    f.mov(asm::qword_base_index("rcx", "r15", 8), reg("rax"));
    f.lea(reg("rcx"), asm::rip("prom_ready"));
    f.mov(asm::byte_base_index("rcx", "r15"), imm(1));
    f.lea(reg("rcx"), asm::rip("actor_state"));
    f.mov(asm::byte_base_index("rcx", "r13"), imm(0));
    f.mov(reg("rcx"), asm::qword_rip("running"));
    f.alu(Alu::Sub, reg("rcx"), imm(1));
    f.mov(asm::qword_rip("running"), reg("rcx"));
    f.call("rt_bump");
    f.call("rt_unlock");
    f.call("rt_wake");
    f.ret();

    f.global("rt_call").label("rt_call");
    f.call("rt_lock");
    f.mov(reg("r11"), reg("rcx"));
    f.mov(reg("rax"), asm::qword_rip("next_promise"));
    f.lea(reg("rcx"), asm::base_disp("rax", 1));
    f.mov(asm::qword_rip("next_promise"), reg("rcx"));
    f.mov(reg("r8"), asm::qword_rip("mbox_w"));
    f.imul(reg("r9"), reg("r8"), imm(40));
    f.lea(reg("r10"), asm::rip("mbox"));
    f.mov(asm::qword_base_index("r10", "r9", 1), reg("rdi"));
    f.mov(asm::qword_base_index_disp("r10", "r9", 1, 8), reg("rsi"));
    f.mov(asm::qword_base_index_disp("r10", "r9", 1, 16), reg("rax"));
    f.mov(asm::qword_base_index_disp("r10", "r9", 1, 24), reg("rdx"));
    f.mov(asm::qword_base_index_disp("r10", "r9", 1, 32), reg("r11"));
    f.unary(Unary::Inc, reg("r8"));
    f.mov(asm::qword_rip("mbox_w"), reg("r8"));
    f.lea(reg("r10"), asm::rip("prom_ready"));
    f.mov(asm::byte_base_index("r10", "rax"), imm(0));
    f.call("rt_bump");
    f.call("rt_unlock");
    f.push(reg("rax"));
    f.call("rt_wake");
    f.pop(reg("rax"));
    f.ret();

    f.global("rt_await").label("rt_await");
    prologue(&mut f);
    f.push(reg("rbx"));
    f.push(reg("r12"));
    f.push(reg("r13"));
    f.push(reg("r14"));
    f.push(reg("r15"));
    f.alu(Alu::Sub, reg("rsp"), imm(8));
    f.mov(reg("r12"), reg("rbx"));
    f.mov(reg("rbx"), reg("rdi"));
    f.cmp(reg("r12"), imm(-1));
    f.jump(Cond::E, "AwaitMark");
    f.lea(reg("rcx"), asm::rip("actor_state"));
    f.mov(asm::byte_base_index("rcx", "r12"), imm(2));
    f.label("AwaitMark");
    f.label("AwaitLoop");
    f.call("rt_lock");
    f.lea(reg("rcx"), asm::rip("prom_ready"));
    f.cmp(asm::byte_base_index("rcx", "rbx"), imm(0));
    f.jump(Cond::Ne, "AwaitReady");
    f.call("rt_find");
    f.cmp(reg("rax"), imm(0));
    f.jump(Cond::E, "AwaitIdle");
    f.call("rt_unlock");
    f.call("rt_exec");
    f.jump(Cond::Always, "AwaitLoop");
    f.label("AwaitIdle");
    f.mov(reg("rax"), asm::qword_rip("running"));
    f.cmp(reg("rax"), imm(0));
    f.jump(Cond::E, "AwaitDead");
    f.mov(reg("r13"), asm::qword_rip("sched_seq"));
    f.call("rt_unlock");
    f.call("rt_park");
    f.jump(Cond::Always, "AwaitLoop");
    f.label("AwaitReady");
    f.cmp(reg("r12"), imm(-1));
    f.jump(Cond::E, "AwaitValue");
    f.lea(reg("rcx"), asm::rip("actor_state"));
    f.mov(asm::byte_base_index("rcx", "r12"), imm(1));
    f.label("AwaitValue");
    f.lea(reg("rcx"), asm::rip("prom_val"));
    f.mov(reg("rax"), asm::qword_base_index("rcx", "rbx", 8));
    f.call("rt_unlock");
    f.alu(Alu::Add, reg("rsp"), imm(8));
    f.pop(reg("r15"));
    f.pop(reg("r14"));
    f.pop(reg("r13"));
    f.pop(reg("r12"));
    f.pop(reg("rbx"));
    f.pop(reg("rbp"));
    f.ret();
    f.label("AwaitDead");
    f.call("rt_unlock");
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
    f.label("mbox").zero(128 * 40);
    f.label("prom_val").zero(256 * 8);
    f.label("prom_ready").zero(256);
    f.label("actor_state").zero(64);
    f.data().align(8);
    f.label("next_actor").quad(&["0"]);
    f.label("next_promise").quad(&["0"]);
    f.label("mbox_r").quad(&["0"]);
    f.label("mbox_w").quad(&["0"]);
    f.label("sched_lock").quad(&["0"]);
    f.label("sched_seq").quad(&["0"]);
    f.label("running").quad(&["0"]);
    f.label("mem_avail").quad(&["0"]);
    f.note_gnu_stack();
    f
}

#[cfg(test)]
mod tests {
    use super::{hal_linux_amd64, runtime_amd64};
    use crate::asm::{self, Alu, File};
    use crate::link::{self, Cpu, Os};

    #[test]
    fn arm64_counter_prints_one() {
        let src = include_str!("../examples/01-counter.lang");
        let object = super::arm_object(&crate::parse::parse(src).unwrap()).unwrap();
        let (stderr, code) = crate::arm64::simulate(&object).unwrap();
        assert_eq!(code, 0);
        assert_eq!(stderr, b"count: 1\n");
    }

    #[test]
    fn arm64_workers_sum_across_actors() {
        let src = include_str!("../examples/05-workers.lang");
        let object = super::arm_object(&crate::parse::parse(src).unwrap()).unwrap();
        let (stderr, code) = crate::arm64::simulate(&object).unwrap();
        assert_eq!(code, 0, "{stderr:?}");
        assert_eq!(stderr, b"4\n11\n16\n5 7\n");
    }

    #[test]
    fn hal_mem_matches_cgroup() {
        let mut probe = File::new();
        probe.intel().text();
        probe.global("_start").label("_start");
        probe.alu(Alu::Xor, asm::reg("rbp"), asm::reg("rbp"));
        probe.call("hal_mem");
        probe.alu(Alu::Sub, asm::reg("rsp"), asm::imm(8));
        probe.mov(asm::qword_base_disp("rsp", 0), asm::reg("rax"));
        probe.mov(asm::reg("rdi"), asm::reg("rsp"));
        probe.mov(asm::reg("rsi"), asm::imm(8));
        probe.call("hal_write");
        probe.alu(Alu::Xor, asm::reg("rdi"), asm::reg("rdi"));
        probe.call("hal_exit");
        let object = crate::amd64::Object::assemble(&[hal_linux_amd64(), probe]).unwrap();
        let bytes = link::write(&object, Os::Linux, Cpu::Amd64).unwrap();
        let path = std::path::Path::new("/tmp/the-lang-mem");
        std::fs::write(path, &bytes).unwrap();
        #[cfg(unix)]
        {
            use std::os::unix::fs::PermissionsExt;
            std::fs::set_permissions(path, std::fs::Permissions::from_mode(0o755)).unwrap();
        }
        let expect = available_memory();
        let out = std::process::Command::new(path).output().unwrap();
        assert!(out.status.success(), "{}", String::from_utf8_lossy(&out.stderr));
        assert!(out.stderr.len() >= 8, "{:?}", out.stderr);
        let got = u64::from_le_bytes(out.stderr[..8].try_into().unwrap());
        let slack = 1024 * 1024 * 1024;
        let low = expect.saturating_sub(slack);
        let high = expect.saturating_add(slack);
        assert!(got >= low && got <= high, "hal_mem {got} expected {expect}");
    }

    fn available_memory() -> u64 {
        let host = mem_available();
        let cgroup = cgroup_available();
        match (cgroup, host) {
            (Some(c), Some(h)) => c.min(h),
            (Some(c), None) => c,
            (None, Some(h)) => h,
            (None, None) => 0,
        }
    }

    fn mem_available() -> Option<u64> {
        let text = std::fs::read_to_string("/proc/meminfo").ok()?;
        for line in text.lines() {
            let Some(rest) = line.strip_prefix("MemAvailable:") else { continue };
            let kb: u64 = rest.split_whitespace().next()?.parse().ok()?;
            return Some(kb * 1024);
        }
        None
    }

    fn cgroup_available() -> Option<u64> {
        let text = std::fs::read_to_string("/proc/self/cgroup").ok()?;
        let path = text.lines().find_map(|line| line.strip_prefix("0::"))?;
        let root = "/sys/fs/cgroup";
        let mut dir = format!("{root}{path}");
        while dir.ends_with('/') && dir.len() > root.len() {
            dir.pop();
        }
        let mut best = None;
        loop {
            if let Some(avail) = dir_available(&dir) {
                best = Some(best.map(|old: u64| old.min(avail)).unwrap_or(avail));
            }
            if dir.len() <= root.len() {
                break;
            }
            let Some(slash) = dir.rfind('/') else { break };
            if slash < root.len() {
                break;
            }
            dir.truncate(slash);
        }
        best
    }

    fn dir_available(dir: &str) -> Option<u64> {
        let current = read_limit(&format!("{dir}/memory.current"))?;
        let mut best = None;
        for name in ["memory.max", "memory.high"] {
            let Some(limit) = read_limit(&format!("{dir}/{name}")) else { continue };
            let avail = limit.saturating_sub(current);
            best = Some(best.map(|old: u64| old.min(avail)).unwrap_or(avail));
        }
        best
    }

    fn read_limit(path: &str) -> Option<u64> {
        let text = std::fs::read_to_string(path).ok()?;
        let text = text.trim();
        if text.is_empty() || text == "max" {
            return None;
        }
        text.parse().ok()
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
