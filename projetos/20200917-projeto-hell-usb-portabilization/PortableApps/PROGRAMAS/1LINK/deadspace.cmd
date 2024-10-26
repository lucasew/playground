@echo off
shift
set PARAMS=%*

pushd %PREFIX%DADOS\Jogos\Dead Space
call __wrapper__ DADOS\Jogos\Dead Space\Dead Space.exe
popd