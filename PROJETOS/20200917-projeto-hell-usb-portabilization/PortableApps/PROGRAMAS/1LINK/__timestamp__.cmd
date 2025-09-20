set _DATEF=%date:/=_%
set _TIMEF=%TIME:,=%
set _TIMEF=%_TIMEF::=_%
set TIMESTAMP=%_DATEF% - %_TIMEF%
