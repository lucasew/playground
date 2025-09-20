@echo off
shift
set PARAMS=%* -jar %BINDIR%\JDownloader\JDownloader.jar
javaw %PARAMS%