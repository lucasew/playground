@echo off
set SAVEPATH=%CSIDL_MYDOCUMENTS%\WatchDogs2\

mkdir %SAVEPATH%

(
echo [Settings]
echo AppID=2688
echo PlayerName=Player
echo SavePath=%SAVEPATH%
echo UnlockDLC=true
echo UplayID=c91c91c9-1c91-c91c-91c9-1c91c91c91c9
) > "%PREFIX%DADOS\Jogos\Watch Dogs 2\bin\CPY.ini"

shift
set PARAMS=%*
pushd %PREFIX%DADOS\Jogos\Watch Dogs 2
start __wrapper__ DADOS\Jogos\Watch Dogs 2\bin\WatchDogs2.exe
rem start __wrapper__ DADOS\Jogos\Watch Dogs 2\EAC.exe
popd