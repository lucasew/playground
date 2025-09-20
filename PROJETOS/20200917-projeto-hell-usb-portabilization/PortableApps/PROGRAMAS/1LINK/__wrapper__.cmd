@echo off
if "%PREFIX%"=="" goto notdefined
if "%*"=="" goto missingParameter

set PROGRAM=%PREFIX%%*
if not exist "%PROGRAM%" goto missingProgram

echo Chamando %PROGRAM% %PARAMS% ...
start "" "%PROGRAM%" %PARAMS%
goto exit

:missingProgram
echo Programa nao encontrado
goto exit

:missingParameter
echo Programa nao especificado
goto exit

:notdefined
echo Nao chamado pelo definirpath, saindo
goto exit

:exit