mod ast;
mod eval;
mod parse;

pub use eval::{run_source, Outcome};

pub fn run_file(path: &str, testing: bool) -> Result<Outcome, String> {
    let src = std::fs::read_to_string(path).map_err(|e| format!("{path}: {e}"))?;
    run_source(&src, testing)
}
