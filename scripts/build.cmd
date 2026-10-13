@echo off
REM Windows build wrapper: runs the PowerShell build script with execution
REM policy bypassed, so it works on default (Restricted) Windows setups.
powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0build.ps1" %*
