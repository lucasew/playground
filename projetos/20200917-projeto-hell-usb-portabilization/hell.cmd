@rem - ignorar o lixo que o bloco de notas deixa
@echo off
set "def_homedir=%homedir%"
set "def_homepath=%homepath%"
set "def_username=%username%"
set "def_temp=%tmp%"

rem Alguns scripts usam a variável HOSTNAME pra saber o hostname da máquina
set HOSTNAME=HD

rem Cores no vim do bash
set TERM=xterm-256color

rem Suporte a reload sem reiniciar a sessão
if ["%def_path%"] == [""] (
    set "def_path=%PATH%"
)
set "PATH=%def_path%"

rem cd to the folder where this script is
rem https://stackoverflow.com/questions/672693/windows-batch-file-starting-directory-when-run-as-admin
pushd %~dp0

rem --- Reduzir BO com letra de unidade
set "prefix=%~dp0"

rem --- Mandar as coisas que iriam para a pasta do usuário no pendrive
set "homedir=%prefix%DADOS\Lucas"
set "HOME=%HOMEDIR%"
set XDG_CONFIG_HOME=%home%\.config
set USERPROFILE=%home%
set HOMEPATH=%home%
set _JAVA_OPTIONS=-Duser.home=%home%

rem --- AppData
set APPDATA=%home%\AppData
set LOCALAPPDATA=%APPDATA%\Local

set CSIDL_MYMUSIC=%home%\Music
set CSIDL_MYVIDEO=%home%\Videos
set CSIDL_MYDOCUMENTS=%home%\Documents

set BINDIR=%PREFIX%PortableApps\PROGRAMAS\

rem --- Nome de usuário
set USER=lucasew
set USERNAME=%USER%

set path=%BINDIR%1LINK;%path%
set path=%BINDIR%;%path%
set path=%BINDIR%GridMove;%path%
set path=%BINDIR%LiveWire;%path%
set path=%BINDIR%UniversalAdbDriver-master;%path%
set path=%BINDIR%vscode;%path%
set path=%BINDIR%ahk;%path%
set path=%BINDIR%argouml-0.34;%path%
set path=%BINDIR%boole;%path%
set path=%BINDIR%ffmpeg;%path%
set path=%BINDIR%git\bin;%path%
set path=%BINDIR%gnirehtet;%path%
set path=%BINDIR%graphviz\bin;%path%
set path=%BINDIR%simpledlna-1.0;%path%
set path=%BINDIR%odin;%path%
set path=%BINDIR%pascalzim;%path%
set path=%BINDIR%platform-tools;%path%
set path=%BINDIR%7z;%path%
set path=%BINDIR%AnyDesk;%path%
set path=%BINDIR%ahk;%path%
set path=%BINDIR%cmake;%path%
set path=%BINDIR%Fritzing;%path%
set path=%BINDIR%curl\bin;%path%
set path=%BINDIR%ffmpeg\bin;%path%
set path=%BINDIR%gcc\bin;%path%
set path=%BINDIR%git\bin;%path%
set path=%BINDIR%graphviz\bin;%path%
set path=%BINDIR%nasm;%path%
set path=%BINDIR%nasm\rdoff;%path%
set path=%BINDIR%nircmd;%path%
set path=%BINDIR%pwsh;%path%
set path=%BINDIR%sqlite;%path%
set path=%BINDIR%Sysinternals;%path%
set path=%BINDIR%ultradefrag-portable-7.1.1.i386;%path%
set path=%BINDIR%zsnesw151;%path%
set path=%BINDIR%python382;%path%
set path=%BINDIR%qemu;%path%
set path=%BINDIR%clisp-2.49;%path%
set path=%BINDIR%racket76;%path%
set path=%BINDIR%neovim\bin;%path%
set path=%BINDIR%wxMEdit;%path%
set path=%BINDIR%tar\bin;%path%
set path=%BINDIR%ImgBurn;%path%
set path=%BINDIR%PCSX2;%path%
set path=%BINDIR%v;%path%
set path=%BINDIR%docker;%path%
set path=%BINDIR%ipmiutil;%path%
set path=%BINDIR%tcltk87-8.7a3-1.tcl87.Win10.x86_64\bin;%path%
set path=%BINDIR%obsidian;%path%

rem -- flutter
set ENABLE_FLUTTER_DESKTOP=true
set path=%BINDIR%flutter\bin;%path%

rem -- sbcl
set path=%BINDIR%sbcl2;%path%
set SBCL_HOME=%BINDIR%sbcl2

rem -- elixir
set path=%BINDIR%elixir\bin;%path%
set path=%BINDIR%erlang\erts-10.7\bin;%path%


rem -- ruby
set path=%BINDIR%ruby265\bin;%path%

rem -- restic
set RESTIC_REPOSITORY=rclone:drive:/BACKUPS/restic
set RESTIC_CACHE_DIR=%def_temp%

rem -- supertuxkart
set path=%BINDIR%stk;%path%
set SUPERTUXKART_DATADIR=%BINDIR%stk

rem --- rust
set RUST_DEFAULT_TOOLCHAIN=x86_64-pc-windows-gnu

rem -- Python
set PYTHONPATH=%BINDIR%python382\Lib\site-packages
set path=%BINDIR%python382\Scripts;%path%

rem --- Android
set "ANDROID_HOME=%BINDIR%android-sdk"
set path=%ANDROID_HOME%\tools\bin;%path%

rem --- Wasmer
set WASMER_DIR=%home%\.wasmer

rem --- Go
set GOPATH=%USERPROFILE%\go
set "GOROOT=%BINDIR%go"
set path=%BINDIR%go\bin;%path%
set path=%home%\go\bin;%path%

rem --- Clang/LLVM
set path=%BINDIR%llvm\bin;%path%

rem --- Temp
set TEMP=%AppData%\Local\Temp
set TMP=%AppData%\Local\Temp

rem --- Java
set "JAVA_HOME=%BINDIR%java8"
set path=%BINDIR%java8\bin;%path%

rem --- Node
set path=%BINDIR%node;%path%
set NODE_PATH=%BINDIR%node\node_modules;%NODE_PATH%

set "CARGO_HOME=%PREFIX%DADOS\Lucas\scoop\persist\rustup\.cargo"
set "RUSTUP_HOME=%PREFIX%DADOS\Lucas\scoop\persist\rustup\.rustup"

popd

set CMDLINE=%*

rem Pega o texto retornado pela fatia do cmdline, que se vazio passa direto
set "expr=%CMDLINE:~0,1%"

rem e checa se não passou direto kkkk
rem enginering :v
if "%expr:~0,1%" == "~" goto end
title Executando %1 via hell

%CMDLINE%

if %ERRORLEVEL% == 0 (
    color 27
    echo Comando terminado com sucesso
) else (
    color 47
    echo Comando terminado com codigo de erro %ERRORLEVEL%
)
:end
