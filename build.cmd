@echo off
rem Builds dist\AMK-Kasse.exe and the 32-bit version.
rem Needs Node.js and Go, or the portable Go in .tools\go.
setlocal
cd /d "%~dp0"
if exist ".tools\go\bin\go.exe" (
  set "PATH=%~dp0.tools\go\bin;%PATH%"
  set "GOPATH=%~dp0.tools\gopath"
  set "GOCACHE=%~dp0.tools\gocache"
  set "GOTOOLCHAIN=local"
)

pushd web
call npm ci || goto :failed
call npm run build || goto :failed
popd

pushd server
go test ./... || goto :failed
go build -trimpath -ldflags "-s -w -H windowsgui" -o ..\dist\AMK-Kasse.exe . || goto :failed
set "GOARCH=386"
go build -trimpath -ldflags "-s -w -H windowsgui" -o ..\dist\AMK-Kasse-32bit.exe . || goto :failed
popd

copy /y packaging\*.cmd dist >nul
copy /y docs\image-credits.md dist >nul
echo Done: dist\AMK-Kasse.exe
exit /b 0

:failed
echo Build failed.
exit /b 1
