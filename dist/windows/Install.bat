@echo off
setlocal

echo ======================================
echo   CalSync - Installation
echo ======================================
echo.

set "APPDIR=%~dp0"
if "%APPDIR:~-1%"=="\" set "APPDIR=%APPDIR:~0,-1%"

set "EXEPATH=%APPDIR%\SyncApp.exe"
set "LOGPATH=%APPDIR%\calsync.log"
set "STARTUPDIR=%APPDATA%\Microsoft\Windows\Start Menu\Programs\Startup"
set "LAUNCHERBAT=%APPDIR%\launcher.bat"
set "STARTUPLINK=%STARTUPDIR%\CalSync.bat"

if not exist "%EXEPATH%" (
  echo Installation failed - SyncApp.exe not found.
  echo Make sure you are running Install.bat from inside the windows folder.
  echo.
  pause
  exit /b 1
)

taskkill /im SyncApp.exe /f >nul 2>&1
del /f "%STARTUPLINK%" >nul 2>&1

echo Installing CalSync...
echo.

(
  echo @echo off
  echo cd /d "%APPDIR%"
  echo start "" "%EXEPATH%"
) > "%LAUNCHERBAT%"

copy /y "%LAUNCHERBAT%" "%STARTUPLINK%" >nul 2>&1
if not exist "%STARTUPLINK%" (
  echo Installation failed - could not write to Startup folder.
  echo Please contact the ACE team.
  echo.
  pause
  exit /b 1
)

echo Starting CalSync...
cd /d "%APPDIR%"
start "" "%EXEPATH%"

timeout /t 4 /nobreak >nul
tasklist /fi "imagename eq SyncApp.exe" 2>nul | find /i "SyncApp.exe" >nul
if errorlevel 1 (
  echo.
  echo The app did not start. Check the log below:
  echo.
  if exist "%LOGPATH%" (
    type "%LOGPATH%"
  ) else (
    echo No log file found.
  )
  echo.
  echo Please contact the ACE team.
  echo.
  pause
  exit /b 1
)

echo Opening the app in your browser...
start "" http://localhost:5001

echo.
echo CalSync installed successfully!
echo.
echo The app will now:
echo   - Start automatically every time you log in
echo   - Sync your calendar every day at 9am
echo   - Catch up automatically if your PC was off
echo.
echo Connect your Salesforce and Google accounts
echo in the browser window that just opened.
echo Then click Sync Now to run your first sync.
echo.
pause
