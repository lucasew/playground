#SingleInstance force

EnvGet, PREFIX, prefix

if (!prefix) {
	run, hell.cmd
	EnvGet, PREFIX, prefix
}

EnvGet, HOMEPATH, HOMEPATH

#r::
InputBox, cmd, HellRun, , , 200 , 100 , , , 30, "cmd"
if (!cmd) {
	return
}
if (cmd == "home") {
	run, explorer %homepath%
	return
}
if (cmd == "hcmd") {
	run, cmd, %homepath%
	return
}

cmd = %prefix%hell.cmd %cmd%
run, %cmd%, %prefix%, max
return

^#x::
    DetectHiddenWindows On
    WinGet, ps, PID, A
    Process, Close, %ps%
    DetectHiddenWindows Off
return

::;shrug::¯\_(ツ)_/¯ 
::;lenny::( ͡° ͜ʖ ͡°)