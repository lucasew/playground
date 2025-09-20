@echo off
shift
set PARAMS=%*
pushd %PREFIX%\DADOS\Jogos\hercules
__wrapper__ DADOS\Jogos\hercules\hercules.exe
popd