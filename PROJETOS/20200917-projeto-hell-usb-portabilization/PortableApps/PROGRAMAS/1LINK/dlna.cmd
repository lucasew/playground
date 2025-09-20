@echo off
shift
rem set PARAMS=%* -path %PREFIX%DADOS\Lucas\Videos
rclone serve dlna %PREFIX%DADOS\Lucas\Videos --name "pobreflix local" -vv