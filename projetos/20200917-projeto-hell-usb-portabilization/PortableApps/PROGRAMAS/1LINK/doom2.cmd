@echo off
shift
set PARAMS=%*
pushd %PREFIX%DADOS\Jogos\DOOM Classic Bundle\DOOM_Classic_2019
__wrapper__ DADOS\Jogos\DOOM Classic Bundle\DOOM_II_Classic_2019\DOOM II.exe
popd