@echo off
setlocal enabledelayedexpansion

@REM All releases - The Go Programming Language https://go.dev/dl/
set "ver=1.23.1"

set "exit_code=1"

:unique_temp_loop
set "temp_dir_path=%TEMP%\%~n0-%RANDOM%"
if exist "!temp_dir_path!" goto unique_temp_loop
mkdir "!temp_dir_path!" || goto :exit

@REM Command in %PATH%
where go >nul 2>&1
if !ERRORLEVEL! == 0 (
    set "go_cmd_path=go"
    goto found_go_cmd
)
@REM Trivial installation paths
set "dirs=%USERPROFILE%\go\go!ver!\bin;%USERPROFILE%\sdk\go!ver!\bin;\Program Files\Go\bin"
for %%d in ("!dirs:;=" "!") do (
    if exist "%%d\go.exe" (
        set "go_cmd_path=%%d\go.exe"
        goto found_go_cmd
    )
)
@REM Download if not found
set "goos=windows"
if "%PROCESSOR_ARCHITECTURE%"=="AMD64" (
    set "goarch=amd64"
) else if "%PROCESSOR_ARCHITEW6432%"=="AMD64" (
    set "goarch=amd64"
) else (
    goto :exit
)
set "sdk_dir_path=%USERPROFILE%\sdk"
if not exist "!sdk_dir_path!" (
    mkdir "!sdk_dir_path!" || goto :exit
)
set "zip_path=!temp_dir_path!\go.zip"
curl.exe --fail --location -o "!zip_path!" "https://go.dev/dl/go!ver!.%goos%-%goarch%.zip" || goto :exit
cd "!sdk_dir_path!" || goto :exit
tar.exe -xf "!zip_path!" || goto :exit
move /y "!sdk_dir_path!\go" "!sdk_dir_path!\go!ver!" || goto :exit
set "go_cmd_path=!sdk_dir_path!\go!ver!\bin\go.exe"

:found_go_cmd
if not defined GOPATH (
    set "GOPATH=%USERPROFILE%\go"
)
if not exist "!GOPATH!\bin" (
    mkdir !%GOPATH!\bin"
)
set "name=%~n0"
if exist "!GOPATH!\bin\embedded-!name!.exe" (
    xcopy /l /d /y "!GOPATH!\bin\embedded-!name!.exe" "%~f0" | findstr /b /c:"1 " >nul 2>&1
    if !ERRORLEVEL! == 0 (
        goto :execute
    )
)

:build
for /f "usebackq tokens=1 delims=:" %%i in (`findstr /n /b :embed_53c8fd5 "%~f0"`) do set n=%%i
set "tempfile=!temp_dir_path!\!name!.go"
more +%n% "%~f0" > "!tempfile!"

!go_cmd_path! build -o !GOPATH!\bin\embedded-!name!.exe "!tempfile!" || goto :exit
del /q "!temp_dir_path!"
goto :execute

:execute
!GOPATH!\bin\embedded-!name!.exe %* || goto :exit
set "exit_code=0"

:exit
if exist "!temp_dir_path!" (
    del /q "!temp_dir_path!"
)
exit /b !exit_code!

endlocal

:embed_53c8fd5
embed_fce761e
