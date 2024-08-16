@echo off

cd /d %~dp0

call install-go.cmd

go gobin-install.go %*
