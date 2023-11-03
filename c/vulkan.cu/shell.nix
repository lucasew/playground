{ pkgs ? import <nixpkgs> { config.allowUnfree = true; }}:

let
  llvmPackages = pkgs.llvmPackages_16;
  spirv-llvm-translator = pkgs.spirv-llvm-translator.override {
    inherit (llvmPackages) llvm;
  };
  compiladorEOSCaraio = pkgs.buildEnv {
    name = "gambiarra";
    extraOutputsToInstall = ["lib" "dev"];
    paths = with llvmPackages; [
      bintools-unwrapped
      # libllvm
      lld
      libcxxabi
      clang-unwrapped
      libstdcxxClang
    ];
  };
  ahfodasse = pkgs.writeShellScriptBin "llvm-config" ''
    echo ${compiladorEOSCaraio}
  '';
  chipStar = llvmPackages.stdenv.mkDerivation rec {
    pname = "chipStar";
    version = "1.0";

    src = pkgs.fetchFromGitHub {
      owner = "CHIP-SPV";
      repo = "chipStar";
      rev = "v${version}";
      hash = "sha256-6VW8/NJsRX5ebjK7am7HcxlJUaga0U3ACsoonpZgjRg=";
    };

    nativeBuildInputs = with pkgs; [
      pkg-config
      cmake
      which
      ahfodasse
    ];

    buildInputs = [
      spirv-llvm-translator
    ];
  };
in

pkgs.mkShell {
  buildInputs = with pkgs; [ cudatoolkit chipStar ];
}
