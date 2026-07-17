@echo off

rem Builds the final application.
rem Run inside res directory.

cd ..
echo Building...
go build -trimpath -ldflags "-s -w -H=windowsgui"

C:\app-portable\Go\gsa.exe id3fit.exe

pause
