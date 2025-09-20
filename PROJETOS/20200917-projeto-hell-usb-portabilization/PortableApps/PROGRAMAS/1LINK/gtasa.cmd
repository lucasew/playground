@echo off
rem configure gta path
rem https://www.gtagarage.com/mods/show.php?id=3861
"%CSIDL_MYDOCUMENTS%\GTA San Andreas User Files\SGPE.exe"

shift
set PARAMS=%*
pushd %PREFIX%DADOS\Jogos\GTA San Andreas
__wrapper__ DADOS\Jogos\GTA San Andreas\gta_sa.exe
popd