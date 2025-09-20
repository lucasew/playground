@echo off

shift
set PARAMS=%*
pushd %PREFIX%DADOS\Jogos\Watch Dogs\bin
call __wrapper__ DADOS\Jogos\Watch Dogs\bin\watch_dogs.exe
popd