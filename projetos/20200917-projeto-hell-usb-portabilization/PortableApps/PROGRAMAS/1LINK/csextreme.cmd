@echo off
shift
set PARAMS=%*
pushd %PREFIX%DADOS\Jogos\Counter-Strike Xtreme V6\
call __wrapper__ DADOS\Jogos\Counter-Strike Xtreme V6\Counter Strike Xtreme.exe
popd