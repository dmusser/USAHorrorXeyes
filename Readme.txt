SWITCHABLE XEYES - BUILD INSTRUCTIONS
=====================================

This file contains platform-specific build instructions for building on each target platform.

PREREQUISITES
-------------
- Go 1.19 or later installed
- Git (optional, for version control)

WINDOWS
-------
Build on Windows using Command Prompt or PowerShell:

1. Open Command Prompt or PowerShell
2. Navigate to the project directory
3. Create output directory:
   mkdir dist
4. Build the executable:
   go build -ldflags="-H windowsgui -s -w" -o dist/xeyes.exe xeyes.go

The -H windowsgui flag creates a Windows GUI application (no console window).

LINUX
-----
Build on Linux using terminal:

1. Open terminal
2. Navigate to the project directory
3. Create output directory:
   mkdir -p dist
4. Build the executable:
   go build -ldflags="-s -w" -o dist/xeyes-linux xeyes.go
5. Make executable (if needed):
   chmod +x dist/xeyes-linux

Note: Requires X11 or Wayland display server for GUI functionality.

MACOS
-----
Build on macOS using terminal:

1. Open Terminal
2. Navigate to the project directory
3. Create output directory:
   mkdir -p dist
4. Build the executable:
   go build -ldflags="-s -w" -o dist/xeyes-macos xeyes.go
5. Make executable (if needed):
   chmod +x dist/xeyes-macos

Note: macOS may require developer tools (Xcode Command Line Tools) for CGO compilation.

RUNNING THE APPLICATION
-----------------------

Windows:
  dist\xeyes.exe

Linux:
  ./dist/xeyes-linux

macOS:
  ./dist/xeyes-macos

CONTROLS
--------
- Single click + drag: Move window
- Double-click: Switch between normal/creepy mode  
- Right-click or F key: Freeze/unfreeze eyes
- ESC key: Exit

TROUBLESHOOTING
---------------

Windows:
- If you get "missing DLL" errors, make sure you have Visual C++ Redistributable installed
- If the window appears but eyes don't work, check Windows graphics drivers

Linux:
- If you get display errors, ensure X11 or Wayland is running
- For "permission denied" errors, run: chmod +x dist/xeyes-linux
- Install graphics libraries if needed: sudo apt install libgl1-mesa-dev

macOS:
- If build fails with CGO errors, install Xcode Command Line Tools:
  xcode-select --install
- If the app won't run, you may need to allow it in System Preferences > Security

LDFLAGS EXPLANATION
-------------------
-s: Strip symbol table (smaller executable)
-w: Strip debug info (smaller executable)  
-H windowsgui: Windows GUI subsystem (Windows only)

FILE SIZES (APPROXIMATE)
-------------------------
Windows: ~9-12 MB
Linux: ~10-14 MB  
macOS: ~12-16 MB

The executables are self-contained with no external dependencies.