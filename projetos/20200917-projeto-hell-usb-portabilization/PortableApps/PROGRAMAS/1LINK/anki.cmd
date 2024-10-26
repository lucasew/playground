@echo off 
shift
set PARAMS=%*
pushd %BINDIR%\anki2121\
__wrapper__ PROGRAMAS\PortableApps\anki2121\anki.exe
popd