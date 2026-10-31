mod amd64;
mod arm64;
mod asm;
mod ast;
mod compile;
mod link;
mod eval;
mod parse;
mod runtime;

pub use compile::{build, build_linux_amd64};
pub use link::{Cpu, Os};

pub use eval::{run_source, Outcome};

pub fn run_file(path: &str, testing: bool) -> Result<Outcome, String> {
    let src = std::fs::read_to_string(path).map_err(|e| format!("{path}: {e}"))?;
    run_source(&src, testing)
}
