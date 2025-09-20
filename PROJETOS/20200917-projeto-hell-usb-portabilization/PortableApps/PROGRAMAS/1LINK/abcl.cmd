@echo off
shift
set PARAMS=%* -jar %BINDIR%\abcl.jar
java %PARAMS%