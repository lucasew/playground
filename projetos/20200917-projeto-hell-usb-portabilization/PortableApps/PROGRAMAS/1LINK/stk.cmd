@echo off
shift
set PARAMS=%*
pushd %BINDIR%\stk
__wrapper__ PortableApps\PROGRAMAS\stk\supertuxkart.exe
popd