@echo off

shift
set PARAMS=%*
pushd %PREFIX%DADOS\Jogos\Grand Theft Auto V
call __wrapper__ DADOS\Jogos\Grand Theft Auto V\PlayGTAV.exe
popd