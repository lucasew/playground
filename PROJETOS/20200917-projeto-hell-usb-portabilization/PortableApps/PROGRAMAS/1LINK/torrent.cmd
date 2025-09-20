@echo off
shift
(
echo {
echo   "AutoStart": true,
echo   "DisableEncryption": false,
echo   "DownloadDirectory": "%prefix%\\DADOS\\Lucas\\Downloads\\TORRENTS",
echo   "EnableUpload": true,
echo   "EnableSeeding": true,
echo   "IncomingPort": 50007
echo }
) > %TEMP%\cloud-torrent.json
set PARAMS=%* --config-path %TEMP%\cloud-torrent.json
__wrapper__ PortableApps\PROGRAMAS\cloudtorrent\cloudtorrent.exe