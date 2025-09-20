@echo off
shift
set PARAMS=%*
pushd %PREFIX%DADOS\Jogos\Need For Speed Underground 2
__wrapper__ DADOS\Jogos\Need For Speed Underground 2\speed2.exe
popd