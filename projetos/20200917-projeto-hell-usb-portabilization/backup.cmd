@echo off
pushd %~dp0

call hell.cmd
dir %BINDIR% > %TMP%\bin.txt
dir %PREFIX%DADOS\Jogos > %TMP%\jogos.txt

hell.cmd restic backup -vvvv -H hd --cleanup-cache ^
-e node_modules ^
-e .ccls-cache ^
-e .cache ^
-e dist ^
-e plugged ^
-e elpa ^
-e .log ^
-e .jar ^
%PREFIX%backup.cmd ^
%PREFIX%hell.cmd ^
%PREFIX%hellexec.ahk ^
%PREFIX%DADOS\Lucas\.emacs.d\init.el ^
%PREFIX%DADOS\Lucas\.config ^
%PREFIX%DADOS\Lucas\IdeaProjects ^
%PREFIX%DADOS\Lucas\CLionProjects ^
%PREFIX%DADOS\Lucas\AppData\.minecraft ^
%PREFIX%DADOS\Lucas\AppData\Local\LumaEmu_SteamCloud ^
%PREFIX%DADOS\Lucas\AppData\Rimworld ^
%PREFIX%DADOS\Lucas\AppData\Roaming\SmartSteamEmu ^
%PREFIX%DADOS\Lucas\CODIGOS ^
%PREFIX%DADOS\Lucas\Documents ^
"%PREFIX%\DADOS\Lucas\AppData\Roaming\Goldberg SocialClub Emu Saves" ^
%BINDIR%\1LINK ^
%TMP%\bin.txt ^
%TMP%\jogos.txt

popd