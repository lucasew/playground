@echo off
shift
set PARAMS=%*
rem https://www.reddit.com/r/RimWorld/comments/4psazi/an_option_to_have_savegames_within_the_game_folder/
set PARAMS=%PARAMS% -savedatafolder=%PREFIX%DADOS\Lucas\AppData\Rimworld
__wrapper__ DADOS\Jogos\RimWorld.v1.3.3060\RimWorldWin64.exe