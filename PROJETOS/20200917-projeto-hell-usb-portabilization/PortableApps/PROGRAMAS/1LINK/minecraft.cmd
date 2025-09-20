@echo off
shift
set PARAMS=%* -jar %BINDIR%\java_minecraft_shiginima.jar
javaw %PARAMS%