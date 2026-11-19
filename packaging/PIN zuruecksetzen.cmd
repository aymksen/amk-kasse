@echo off
rem Clears the OBK PIN. The next visit to the OBK area asks for a new one.
start "" "%~dp0AMK-Kasse.exe" -reset-pin
