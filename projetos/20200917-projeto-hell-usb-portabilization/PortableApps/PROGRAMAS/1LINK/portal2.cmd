@echo off
shift
set PARAMS=%*

pushd %PREFIX%DADOS\Jogos\Portal 2 Complete\Portal 2
call __wrapper__ DADOS\Jogos\Portal 2 Complete\Portal 2\Portal 2 Launcher.exe
popd