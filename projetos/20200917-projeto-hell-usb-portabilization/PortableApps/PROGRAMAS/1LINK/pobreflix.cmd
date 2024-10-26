@echo off
shift
rem set PARAMS=%* -path %PREFIX%DADOS\Lucas\Videos
rclone serve dlna pobreflix: --name "pobreflix cloud" -vv --addr :7830