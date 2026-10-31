use std::process::ExitCode;

use usage_rs::{Args, Cli, Subcommands, ValueEnum};

/// the-lang toolchain.
#[derive(Cli)]
#[usage(
    bin = "the-lang",
    version = "0.1.0",
    unknown_flags = "error",
    disable_help_subcommand,
    disable_version_flag
)]
struct App {
    /// Print version.
    #[usage(short = 'V', long, global, action = usage_rs::ArgAction::Version)]
    version: bool,

    #[usage(subcommand)]
    command: Command,
}

#[derive(Subcommands)]
enum Command {
    /// Write one native executable.
    Build(Build),
    /// Start the program.
    Run(Source),
    /// Run each top-level test block.
    Test(Source),
    /// Run the generators the source names.
    Generate(Source),
}

#[derive(Args)]
#[usage(unknown_flags = "error")]
struct Build {
    /// File `build` writes.
    ///
    /// `binary` is one executable. `c-shared` is one C shared library.
    #[usage(long, value_enum, default = "binary")]
    artifact: Artifact,

    /// Target operating system.
    ///
    /// Absent means the host. The names are `linux`, `darwin`, and `windows`.
    #[usage(long, value_enum)]
    os: Option<TargetOs>,

    /// Target architecture.
    ///
    /// Absent means the host. The names are `amd64` and `arm64`.
    #[usage(long, value_enum)]
    arch: Option<TargetArch>,

    /// File path.
    ///
    /// Absent means the base name of the source path, in the current directory.
    #[usage(long)]
    output: Option<String>,

    /// Source file.
    path: String,
}

#[derive(Args)]
#[usage(unknown_flags = "error")]
struct Source {
    /// Source file.
    path: String,
}

#[derive(Clone, Copy, ValueEnum)]
enum TargetOs {
    Linux,
    Darwin,
    Windows,
}

#[derive(Clone, Copy, ValueEnum)]
enum TargetArch {
    #[usage(name = "amd64")]
    Amd64,
    #[usage(name = "arm64")]
    Arm64,
}

#[derive(Clone, Copy, ValueEnum)]
enum Artifact {
    /// One executable.
    Binary,
    /// One C shared library.
    #[usage(name = "c-shared")]
    CShared,
}

fn main() -> ExitCode {
    let app = App::parse();
    let _ = app.version;
    match app.command {
        Command::Build(build) => match build_cmd(&build) {
            Ok(()) => ExitCode::from(0),
            Err(e) => {
                eprintln!("{e}");
                ExitCode::from(1)
            }
        }
        Command::Generate(_) => {
            eprintln!("generate: the marker grammar is later");
            ExitCode::from(1)
        }
        Command::Run(source) => run_path(&source.path, false),
        Command::Test(source) => run_path(&source.path, true),
    }
}

fn build_cmd(build: &Build) -> Result<(), String> {
    let os = build.os.unwrap_or_else(host_os);
    let arch = build.arch.unwrap_or_else(host_arch);
    let output = build.output.clone().unwrap_or_else(|| {
        std::path::Path::new(&build.path)
            .file_stem()
            .and_then(|s| s.to_str())
            .unwrap_or("a.out")
            .to_string()
    });
    match (os, arch, build.artifact) {
        (TargetOs::Linux, TargetArch::Amd64, Artifact::Binary) => {
            let src = std::fs::read_to_string(&build.path).map_err(|e| format!("{}: {e}", build.path))?;
            the_lang::build_linux_amd64(&src, std::path::Path::new(&output))
        }
        (_, _, Artifact::CShared) => Err("build: c-shared is not in this port".into()),
        (os, arch, _) => Err(format!("build: no port for {os}/{arch}")),
    }
}

fn host_os() -> TargetOs {
    match std::env::consts::OS {
        "macos" => TargetOs::Darwin,
        "windows" => TargetOs::Windows,
        _ => TargetOs::Linux,
    }
}

fn host_arch() -> TargetArch {
    match std::env::consts::ARCH {
        "aarch64" => TargetArch::Arm64,
        _ => TargetArch::Amd64,
    }
}

impl std::fmt::Display for TargetOs {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.write_str(match self {
            TargetOs::Linux => "linux",
            TargetOs::Darwin => "darwin",
            TargetOs::Windows => "windows",
        })
    }
}

impl std::fmt::Display for TargetArch {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.write_str(match self {
            TargetArch::Amd64 => "amd64",
            TargetArch::Arm64 => "arm64",
        })
    }
}

fn run_path(path: &str, testing: bool) -> ExitCode {
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
