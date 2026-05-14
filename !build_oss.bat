@echo off
echo Building Evidence-Capture (OSS Version)...

:: iconリソースの生成 (rsrcがインストールされている前提)
rsrc -manifest main.manifest -ico "icon\favicon.ico" -o cmd\capture\rsrc.syso

:: OSS版としてビルド (-tags pro を付けない)
go build -ldflags "-H windowsgui -s -w" -o Evidence-Capture-OSS.exe .\cmd\capture

if %ERRORLEVEL% neq 0 (
    echo [ERROR] Build failed!
    pause
    exit /b %ERRORLEVEL%
)

echo Build Complete: Evidence-Capture-OSS.exe
pause