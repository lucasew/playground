@echo off
shift
set PARAMS=%*

pushd %PREFIX%DADOS\Jogos\The Elder Scrolls V Skyrim Special Edition
__wrapper__ DADOS\Jogos\The Elder Scrolls V Skyrim Special Edition\SkyrimSELauncher.exe
popd