use std::env;
use std::process::ExitCode;

fn main() -> ExitCode {
    let mut args: Vec<String> = env::args().skip(1).collect();
    if args.iter().any(|a| a == "--help" || a == "-h") {
        println!("the-lang <command> [path]");
        println!("commands: build, run, test, generate");
        return ExitCode::SUCCESS;
    }
    if args.iter().any(|a| a == "--version") {
        println!("the-lang 0.1.0");
        return ExitCode::SUCCESS;
    }
    args.retain(|a| a != "--");
    let Some(command) = args.first().cloned() else {
        eprintln!("usage: the-lang <command> [path]");
        return ExitCode::from(2);
    };
    let path = args.get(1).map(String::as_str);
    match command.as_str() {
        "run" => run_path(path, false),
        "test" => run_path(path, true),
        "build" => {
            eprintln!("build: native codegen is later");
            ExitCode::from(1)
        }
        "generate" => {
            eprintln!("generate: the marker grammar is later");
            ExitCode::from(1)
        }
        _ => {
            eprintln!("unknown command {command}");
            ExitCode::from(2)
        }
    }
}

fn run_path(path: Option<&str>, testing: bool) -> ExitCode {
    let Some(path) = path else {
        eprintln!("expected a source file");
        return ExitCode::from(2);
    };
    match the_lang::run_file(path, testing) {
        Ok(out) => {
            eprint!("{}", out.stderr);
            ExitCode::from(out.status as u8)
        }
        Err(e) => {
            eprintln!("{e}");
            ExitCode::from(1)
        }
    }
}
