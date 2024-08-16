@echo off

REM Go version
set "ver=1.23.0"

REM Look up the Go executable in the PATH.
for %%p in (
    "C:\Program Files\Go\bin",
    "C:\Program Files (x86)\Go\bin",
    "%USERPROFILE%\go\bin"
) do (
    if exist "%%~p\go.exe" (
        set "PATH=%%~p;%PATH%"
        goto :End
    )
)

REM If Go is not found, download and install it.

set "goos=windows"
set "goarch=amd64"

REM Download and install Go.
powershell -Command "Invoke-WebRequest -Uri https://golang.org/dl/go%ver%.%goos%-%goarch%.msi -OutFile go-installer.msi"
msiexec /i go-installer.msi /quiet /norestart
set "PATH=C:\Program Files\Go\bin;%PATH%"

:End
exit /b 0
