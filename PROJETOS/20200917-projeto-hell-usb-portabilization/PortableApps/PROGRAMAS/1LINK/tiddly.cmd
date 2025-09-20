@echo off
shift
set PARAMS=%*
pushd %PREFIX%\PortableApps\PROGRAMAS\TiddlyDesktop-win64-v0.0.14
call __wrapper__ PortableApps\PROGRAMAS\TiddlyDesktop-win64-v0.0.14\nw.exe
popd