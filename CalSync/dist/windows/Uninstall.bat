@echo off

echo ======================================
echo   CalSync - Uninstall
echo ======================================
echo.

set "STARTUPLINK=%APPDATA%\Microsoft\Windows\Start Menu\Programs\Startup\CalSync.bat"

taskkill /im SyncApp.exe /f >nul 2>&1
del /f "%STARTUPLINK%" >nul 2>&1
schtasks /delete /tn "CalSync" /f >nul 2>&1
schtasks /delete /tn "CalSync Weekly" /f >nul 2>&1

echo CalSync has been uninstalled.
echo.
echo   - The app will no longer start on login
echo   - Scheduled syncs have been stopped
echo   - Your calendar data and sync history are untouched
echo.
echo To reinstall later, run Install.bat.
echo.
pause
