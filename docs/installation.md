# Installation Guide

## Mac

### Step-by-step

> **Important:** The first time you install, macOS will block the installer because it was not downloaded from the App Store. This is normal and safe — follow the steps below to allow it.

1. Extract the zip file you received
2. Move the `CalSync-Share` folder into your **Applications** folder
3. Open the `mac` folder inside it
4. Double-click **`Install.command`**
   - A Terminal window will open briefly
   - macOS may show a popup saying it was blocked — this is expected
5. Open **System Settings** → **Privacy & Security**
6. Scroll down until you see: *"Install.command was blocked from use because it is not from an identified developer"*
7. Click **Open Anyway**
8. Enter your Mac password if prompted
9. A Terminal window will run the installer — your browser will open automatically when it's done
10. Follow the setup steps in your browser (see [Setup Guide](./setup.md))

### What the installer does

- Removes macOS security quarantine flags from the app files
- Digitally signs the app binary so it can run
- Registers the app as a **Login Item** (`~/Library/LaunchAgents/com.ace.calsync.plist`) so it starts automatically every time you log in
- Starts the app immediately and opens the browser setup page

### Uninstalling on Mac

1. Open the `mac` folder
2. Double-click **`Uninstall.command`**
   - Follow the same Privacy & Security steps if macOS blocks it
3. The app is removed from startup — your calendar data is untouched

---

## Windows

### Step-by-step

1. Extract the zip file you received
2. Open the `windows` folder
3. Double-click **`Install.bat`**
   - A command prompt window will open and run the installer
   - Your browser will open automatically when it's done
4. Follow the setup steps in your browser (see [Setup Guide](./setup.md))

### What the installer does

- Creates a small launcher script and places it in your Windows **Startup folder** (`%APPDATA%\Microsoft\Windows\Start Menu\Programs\Startup\`) so the app starts automatically every time you log in
- Starts the app immediately and opens the browser setup page

### Uninstalling on Windows

1. Open the `windows` folder
2. Double-click **`Uninstall.bat`**
3. The app is removed from startup — your calendar data is untouched

---

## Troubleshooting Installation

### Mac: The app didn't open in my browser

Check the log file for errors:
```
CalSync-Share/mac/calsync.log
```

### Mac: "Install.command was blocked" option doesn't appear

Wait a moment and scroll down fully in Privacy & Security. If it still doesn't appear, try right-clicking `Install.command` and selecting **Open**.

### Windows: The command prompt closed immediately

Right-click `Install.bat` and choose **Run as administrator**, then check the log:
```
CalSync-Share\windows\calsync.log
```

### The app is running but nothing syncs

See the [Troubleshooting guide](./troubleshooting.md).
