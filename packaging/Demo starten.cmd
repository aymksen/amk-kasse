@echo off
rem Starts the till with invented sales in a separate folder. Real data is not touched.
start "" "%~dp0AMK-Kasse.exe" -demo -windowed -port 8418
