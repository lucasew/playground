@echo off

shift
set PARAMS=%*
pushd %PREFIX%DADOS\Jogos\Scrap Mechanic v0.4.6.578\Release
call __wrapper__ DADOS\Jogos\Scrap Mechanic v0.4.6.578\Release\ScrapMechanic.exe
popd