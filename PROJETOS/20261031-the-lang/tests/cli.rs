use std::process::{Command, Output};

fn bin() -> Command {
    let mut exe = std::env::current_exe().expect("test executable");
    exe.pop();
    exe.pop();
    exe.push("the-lang");
    Command::new(exe)
}

fn run(args: &[&str]) -> Output {
    bin().args(args).output().expect("spawn the-lang")
}

#[test]
fn help_lists_commands() {
    let out = run(&["--help"]);
    assert!(out.status.success());
    let stdout = String::from_utf8(out.stdout).unwrap();
    assert!(stdout.contains("build"), "{stdout}");
    assert!(stdout.contains("run"), "{stdout}");
    assert!(stdout.contains("test"), "{stdout}");
    assert!(stdout.contains("generate"), "{stdout}");
    assert!(!stdout.contains("--os"), "{stdout}");
    assert!(!stdout.contains("--artifact"), "{stdout}");
    assert!(out.stderr.is_empty());
}

#[test]
fn version_is_plain() {
    let out = run(&["--version"]);
    assert!(out.status.success());
    assert_eq!(String::from_utf8(out.stdout).unwrap(), "the-lang 0.1.0\n");
    assert!(out.stderr.is_empty());
}

#[test]
fn version_is_valid_on_a_command() {
    let out = run(&["run", "--version"]);
    assert!(out.status.success(), "{}", String::from_utf8_lossy(&out.stderr));
    assert_eq!(String::from_utf8(out.stdout).unwrap(), "the-lang 0.1.0\n");
}

#[test]
fn missing_command_is_usage() {
    let out = run(&[]);
    assert_eq!(out.status.code(), Some(2));
    assert!(!out.stderr.is_empty());
}

#[test]
fn unknown_command_is_usage() {
    let out = run(&["fmt"]);
    assert_eq!(out.status.code(), Some(2));
}

#[test]
fn unknown_flag_is_usage() {
    let out = run(&["run", "--nope", "examples/01-counter.lang"]);
    assert_eq!(out.status.code(), Some(2));
}

#[test]
fn missing_path_is_usage() {
    let out = run(&["run"]);
    assert_eq!(out.status.code(), Some(2));
}

#[test]
fn bad_os_is_usage() {
    let out = run(&["build", "--os", "plan9", "examples/01-counter.lang"]);
    assert_eq!(out.status.code(), Some(2));
}

#[test]
fn os_on_run_is_usage() {
    let out = run(&["run", "--os", "darwin", "examples/01-counter.lang"]);
    assert_eq!(out.status.code(), Some(2));
}

#[test]
fn arch_on_test_is_usage() {
    let out = run(&["test", "--arch", "arm64", "examples/04-test.lang"]);
    assert_eq!(out.status.code(), Some(2));
}

#[test]
fn build_accepts_target_after_path() {
    let out = run(&[
        "build",
        "examples/01-counter.lang",
        "--os",
        "darwin",
        "--arch",
        "arm64",
        "--artifact",
        "c-shared",
    ]);
    assert_eq!(out.status.code(), Some(1));
    let err = String::from_utf8(out.stderr).unwrap();
    assert!(err.contains("codegen"), "{err}");
}

#[test]
fn artifact_on_test_is_usage() {
    let out = run(&["test", "--artifact", "binary", "examples/04-test.lang"]);
    assert_eq!(out.status.code(), Some(2));
}

#[test]
fn bad_artifact_is_usage() {
    let out = run(&["build", "--artifact", "pie", "examples/01-counter.lang"]);
    assert_eq!(out.status.code(), Some(2));
}

#[test]
fn lone_dash_is_a_path() {
    let out = run(&["run", "-"]);
    assert_eq!(out.status.code(), Some(1));
    let err = String::from_utf8(out.stderr).unwrap();
    assert!(err.contains('-'), "{err}");
}

#[test]
fn build_reports_codegen_later() {
    let out = run(&["build", "examples/01-counter.lang"]);
    assert_eq!(out.status.code(), Some(1));
    assert!(out.stdout.is_empty());
    assert!(String::from_utf8(out.stderr).unwrap().contains("codegen"));
}

#[test]
fn generate_reports_marker_later() {
    let out = run(&["generate", "examples/01-counter.lang"]);
    assert_eq!(out.status.code(), Some(1));
    assert!(String::from_utf8(out.stderr).unwrap().contains("marker"));
}
