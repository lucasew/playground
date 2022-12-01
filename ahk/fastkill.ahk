^#x::
    WinGet, ps, ProcessName, A
    Run, taskkill -im %ps%
    return

^#u::
    Run, taskkill -im hl.exe
    return