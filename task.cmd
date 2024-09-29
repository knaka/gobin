@echo off
setlocal enabledelayedexpansion
if "%~1" == "update-me" (
  curl.exe --fail --location --output %TEMP%\task_cmd-%~nx0 https://raw.githubusercontent.com/knaka/src/main/task.cmd || exit /b !ERRORLEVEL!
  move /y %TEMP%\task_cmd-%~nx0 %~f0
  exit /b 0
)

@REM BusyBox for Windows https://frippery.org/busybox/index.html
@REM Index of /files/busybox https://frippery.org/files/busybox/?C=M;O=D
set ver=FRP-5398-g89ae34445
if "%PROCESSOR_ARCHITECTURE%" == "x86" (
  echo "WARNING: Your environment is 32-bit. Not all features are supported." >&2
  set arch=32
) else if "%PROCESSOR_ARCHITECTURE%" == "AMD64" (
  set arch=64u
) else if "%PROCESSOR_ARCHITECTURE%" == "ARM64" (
  set arch=64a
) else (
  exit /b 1 
)
set cmd_name=busybox-w!arch!-!ver!.exe
set bin_dir_path=%USERPROFILE%\.bin
if not exist !bin_dir_path! (
  mkdir "!bin_dir_path!"
)
set cmd_path=!bin_dir_path!\!cmd_name!
if not exist !cmd_path! (
  curl.exe --fail --location --output "!cmd_path!" https://frippery.org/files/busybox/!cmd_name! || exit /b !ERRORLEVEL!
)

set "ARG0=%~f0"
set "ARG0BASE=%~n0"
set script_dir_path=%~dp0
set script_name=%~n0
set sh_dir_path=!script_dir_path!
set env_file_path=!script_dir_path!\.env.sh.cmd
if exist !env_file_path! (
  call !env_file_path!
)
set BB_GLOBBING=0
!cmd_path! sh !sh_dir_path!\!script_name!.sh %* || exit /b !ERRORLEVEL!
endlocal
