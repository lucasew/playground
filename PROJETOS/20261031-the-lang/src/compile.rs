//! Lower a program to calls on the runtime. The runtime calls `hal_write` and `hal_exit` only.

use std::path::Path;
use std::process::Command;

use crate::ast::*;
use crate::parse;

pub fn build_linux_amd64(src: &str, output: &Path) -> Result<(), String> {
    let program = parse::parse(src)?;
    let asm = lower(&program)?;
    let dir = std::env::temp_dir().join(format!("the-lang-{}", std::process::id()));
    std::fs::create_dir_all(&dir).map_err(|e| e.to_string())?;
    let prog = dir.join("prog.s");
    let hal = dir.join("hal.s");
    let rt = dir.join("rt.s");
    std::fs::write(&prog, asm).map_err(|e| e.to_string())?;
    std::fs::write(&hal, include_str!("../rt/hal_linux_amd64.s")).map_err(|e| e.to_string())?;
    std::fs::write(&rt, include_str!("../rt/runtime_amd64.s")).map_err(|e| e.to_string())?;
    let prog_o = dir.join("prog.o");
    let hal_o = dir.join("hal.o");
    let rt_o = dir.join("rt.o");
    assemble(&prog, &prog_o)?;
    assemble(&hal, &hal_o)?;
    assemble(&rt, &rt_o)?;
    if let Some(parent) = output.parent() {
        if !parent.as_os_str().is_empty() {
            std::fs::create_dir_all(parent).map_err(|e| e.to_string())?;
        }
    }
    let status = Command::new("ld")
        .args(["-nostdlib", "-static", "-e", "_start", "-o"])
        .arg(output)
        .arg(&prog_o)
        .arg(&rt_o)
        .arg(&hal_o)
        .output()
        .map_err(|e| format!("ld: {e}"))?;
    if !status.status.success() {
        return Err(format!("ld: {}", String::from_utf8_lossy(&status.stderr)));
    }
    Ok(())
}

fn assemble(src: &Path, obj: &Path) -> Result<(), String> {
    let status = Command::new("as")
        .arg("-o")
        .arg(obj)
        .arg(src)
        .output()
        .map_err(|e| format!("as: {e}"))?;
    if status.status.success() {
        Ok(())
    } else {
        Err(format!("as: {}", String::from_utf8_lossy(&status.stderr)))
    }
}

struct MethodRef {
    type_name: String,
    name: String,
    fields: Vec<String>,
    body: Vec<Stmt>,
}

fn lower(program: &Program) -> Result<String, String> {
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
    let mut out = String::from(".intel_syntax noprefix\n.text\n");
    for (i, method) in methods.iter().enumerate() {
        out.push_str(&format!("# {}.{}\n.global m{i}\nm{i}:\n", method.type_name, method.name));
        out.push_str("    push rbp\n    mov rbp, rsp\n    push rbx\n    mov rbx, rdi\n");
        lower_method(&method.body, &method.fields, &mut out)?;
        out.push_str("    pop rbx\n    pop rbp\n    ret\n");
    }
    out.push_str(".global lang_methods\nlang_methods:\n");
    for i in 0..methods.len() {
        out.push_str(&format!("    .quad m{i}\n"));
    }
    if methods.is_empty() {
        out.push_str("    .quad 0\n");
    }
    out.push_str(".global lang_main\nlang_main:\n");
    out.push_str("    push rbp\n    mov rbp, rsp\n    sub rsp, 64\n");
    lower_main(&main, &actors, &methods, &mut out)?;
    out.push_str("    leave\n    ret\n");
    out.push_str(".section .note.GNU-stack,\"\",@progbits\n");
    Ok(out)
}

fn lower_method(body: &[Stmt], fields: &[String], out: &mut String) -> Result<(), String> {
    if body.is_empty() {
        out.push_str("    xor rax, rax\n");
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
                let slot = field_slot(fields, name)?;
                match op {
                    AssignOp::PlusEq => {
                        let Expr::Int(n) = value else {
                            return Err("build: self.field += needs an integer".into());
                        };
                        out.push_str(&format!(
                            "    mov rdi, rbx\n    mov rsi, {slot}\n    call rt_load\n    add rax, {n}\n    mov rdx, rax\n    mov rdi, rbx\n    mov rsi, {slot}\n    call rt_store\n"
                        ));
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
                let slot = field_slot(fields, name)?;
                out.push_str(&format!(
                    "    mov rdi, rbx\n    mov rsi, {slot}\n    call rt_load\n"
                ));
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
    out: &mut String,
) -> Result<(), String> {
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
                let slot = locals.len();
                locals.push(name.clone());
                let off = (slot as i32 + 1) * 8;
                out.push_str(&format!("    call rt_actor\n    mov qword ptr [rbp-{off}], rax\n"));
            }
            Stmt::Expr(expr) => lower_main_expr(expr, &locals, methods, &mut strings, out)?,
        }
    }
    if !strings.is_empty() {
        out.push_str(".section .rodata\n");
        for (i, s) in strings.iter().enumerate() {
            out.push_str(&format!("S{i}:\n    .ascii \""));
            for b in s.bytes() {
                out.push_str(&format!("\\x{b:02x}"));
            }
            out.push_str("\"\n");
        }
        out.push_str(".text\n");
    }
    Ok(())
}

fn lower_main_expr(
    expr: &Expr,
    locals: &[String],
    methods: &[MethodRef],
    strings: &mut Vec<String>,
    out: &mut String,
) -> Result<(), String> {
    if let Some((local, method)) = await_call(expr) {
        let slot = locals.iter().position(|n| n == &local).ok_or_else(|| format!("build: no local {local}"))?;
        let id = methods
            .iter()
            .position(|m| m.name == method)
            .ok_or_else(|| format!("build: no method {method}"))?;
        let off = (slot as i32 + 1) * 8;
        out.push_str(&format!(
            "    mov rdi, qword ptr [rbp-{off}]\n    mov rsi, {id}\n    call rt_call\n    mov rdi, rax\n    call rt_await\n"
        ));
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
                    let len = text.len();
                    strings.push(text);
                    out.push_str(&format!(
                        "    lea rdi, [rip+S{id}]\n    mov rsi, {len}\n    call rt_write\n"
                    ));
                }
                Piece::Int => {
                    let arg = rest.next().ok_or("build: fprintf is missing a value")?;
                    if let Some((local, method)) = await_call(arg) {
                        let slot = locals.iter().position(|n| n == &local).ok_or_else(|| format!("build: no local {local}"))?;
                        let id = methods.iter().position(|m| m.name == method).ok_or_else(|| format!("build: no method {method}"))?;
                        let off = (slot as i32 + 1) * 8;
                        out.push_str(&format!(
                            "    mov rdi, qword ptr [rbp-{off}]\n    mov rsi, {id}\n    call rt_call\n    mov rdi, rax\n    call rt_await\n    mov rdi, rax\n    call rt_write_i64\n"
                        ));
                    } else {
                        return Err("build: fprintf %d needs an awaited call".into());
                    }
                }
            }
        }
        return Ok(());
    }
    Err("build: unsupported statement in main".into())
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

fn field_slot(fields: &[String], name: &str) -> Result<usize, String> {
    fields.iter().position(|n| n == name).ok_or_else(|| format!("build: no field {name}"))
}
