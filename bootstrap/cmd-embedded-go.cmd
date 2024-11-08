@echo off
setlocal enabledelayedexpansion

if "%~1" == "update-me" (
  curl.exe --fail --location --output %TEMP%\cmd-%~nx0 https://raw.githubusercontent.com/knaka/src/main/go/go-embedded.cmd || exit /b !ERRORLEVEL!
  move /y %TEMP%\cmd-%~nx0 %~f0
  exit /b 0
)

@REM All releases - The Go Programming Language https://go.dev/dl/
set required_min_major_ver=1
set required_min_minor_ver=23
set ver=!required_min_major_ver!.!required_min_minor_ver!.1

set exit_code=1

:unique_temp_loop
set "temp_dir_path=%TEMP%\%~n0-%RANDOM%"
if exist "!temp_dir_path!" goto unique_temp_loop
mkdir "!temp_dir_path!" || goto :exit
call :to_short_path "!temp_dir_path!"
set temp_dir_spath=!short_path!

set goroot_dir_spath=

call :to_short_path "C:\Program Files"
set program_files_spath=!short_path!
call :to_short_path "%USERPROFILE%"
set user_profile_spath=!short_path!

@REM Command in %PATH%
where go >nul 2>&1 
if !ERRORLEVEL! == 0 (
  for /F "usebackq tokens=*" %%p in (`where go`) do (
    call :to_short_path "%%p"
    set cmd_spath=!short_path!
    call :set_proper_goroot_dir_spath !cmd_spath!
    if !goroot_dir_spath! neq "" (
      goto :found_goroot
    )
  )
)

@REM Trivial installation paths
@REM for /D %%d in (!program_files_spath!\go !user_profile_spath!\sdk\go*) do (
for /D %%d in (!user_profile_spath!\sdk\go*) do (
  set cmd_spath=%%d\bin\go.exe
  if exist !cmd_spath! (
    call :set_proper_goroot_dir_spath !cmd_spath!
    if !goroot_dir_spath! neq "" (
      goto :found_goroot
    )
  )
)

@REM Download if not found
set goos=windows
if "%PROCESSOR_ARCHITECTURE%"=="AMD64" (
  set goarch=amd64
) else if "%PROCESSOR_ARCHITEW6432%"=="AMD64" (
  set goarch=amd64
) else (
  goto :exit
)
set sdk_dir_spath=!user_profile_spath!\sdk
if not exist !sdk_dir_spath! (
  mkdir !sdk_dir_spath! || goto :exit
)
set zip_spath=!temp_dir_spath!\go.zip
echo Downloading Go SDK: !ver! >&2
curl.exe --fail --location -o !zip_spath! "https://go.dev/dl/go!ver!.%goos%-%goarch%.zip" || goto :exit
cd !sdk_dir_spath! || goto :exit
unzip -q !zip_spath! || goto :exit
move /y !sdk_dir_spath!\go !sdk_dir_spath!\go!ver! || goto :exit
set goroot_dir_spath=!sdk_dir_spath!\go!ver!

:found_goroot

set go_cmd_spath=!goroot_dir_spath!\bin\go.exe
set GOROOT=

if not defined GOPATH (
  set GOPATH=!user_profile_spath!\go
)

if not exist !GOPATH!\bin (
  mkdir !GOPATH!\bin
)

set "name=embedded-%~f0"
set "name=!name: =_!"
set "name=!name:\=_!"
set "name=!name::=_!"
set "name=!name:/=_!"
if exist !GOPATH!\bin\!name!.exe (
  xcopy /l /d /y !GOPATH!\bin\!name!.exe "%~f0" | findstr /b /c:"1 " >nul 2>&1
  if !ERRORLEVEL! == 0 (
    goto :execute
  )
)

:build
echo Using Go compiler: !go_cmd_spath! >&2
for /f "usebackq tokens=1 delims=:" %%i in (`findstr /n /b :embed_53c8fd5 "%~f0"`) do set n=%%i
set tempfile=!temp_dir_spath!\!name!.go
more +%n% "%~f0" > !tempfile!

!go_cmd_spath! build -o !GOPATH!\bin\!name!.exe !tempfile! || goto :exit
del /q !temp_dir_spath!

:execute
!GOPATH!\bin\!name!.exe %* || goto :exit
set exit_code=0

:exit
if exist !temp_dir_spath! (
  del /q !temp_dir_spath!
)
exit /b !exit_code!

:to_short_path
set "input_path=%~1"
for %%i in ("%input_path%") do set "short_path=%%~si"
exit /b
goto :eof

:set_proper_goroot_dir_spath
for /F "usebackq tokens=*" %%v in (`%1 env GOVERSION`) do (
  set version=%%v
  set major=!version:~2,1!
  set minor=!version:~4,2!
  if !major! geq !required_min_major_ver! (
  if !minor! geq !required_min_minor_ver! (
    for /F "useback tokens=*" %%p in (`%1 env GOROOT`) do (
      call :to_short_path "%%p"
      set goroot_dir_spath=!short_path!
    )
  )
)
exit /b
goto :eof

endlocal

:embed_53c8fd5
package main

import (
    "os"
)

func main() {
    println("Hello, World!")
    for i, arg := range os.Args {
        println(i, arg)
    }
    cwd, _ := os.Getwd()
    println("CWD:", cwd)
}
