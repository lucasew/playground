@echo off
shift
set PARAMS=%*
pushd %PREFIX%DADOS\Jogos\FlatOut 2
__wrapper__ DADOS\Jogos\FlatOut 2\FlatOut2.exe
popd