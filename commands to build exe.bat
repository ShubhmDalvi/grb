@echo off
echo ==============================
echo   Building grb - Clipboard Manager
echo ==============================

REM Step 1: Initialize module (only once)
IF NOT EXIST go.mod (
    echo Initializing Go module...
    go mod init grb
) ELSE (
    echo go.mod already exists, skipping init.
)

REM Step 2: Install dependencies
echo Installing dependencies...
go get github.com/spf13/cobra@latest
go get github.com/atotto/clipboard@latest
go get github.com/fatih/color@latest
go get go.etcd.io/bbolt@latest
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/bubbles@latest
go get github.com/sahilm/fuzzy@latest
go get github.com/olekukonko/tablewriter

REM Step 3: Tidy modules
echo Tidying modules...
go mod tidy

REM Step 4: Build the executable
echo Building grb.exe...
go build -o grb.exe

IF %ERRORLEVEL% NEQ 0 (
    echo.
    echo ❌ Build failed! Check the error above.
    echo.
    pause
    exit /b %ERRORLEVEL%
) ELSE (
    echo.
    echo ✔ Build successful! You can now run .\grb.exe
    echo.
    pause
)
