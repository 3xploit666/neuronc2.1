@echo off
REM build_agent.bat

if "%~1"=="" (
    echo Usage: build_agent.bat DEPLOY-TOKEN [output.exe]
    echo Example: build_agent.bat DEPLOY-abc123def456
    exit /b 1
)

set TOKEN=%1
set OUTPUT=%2
if "%OUTPUT%"=="" set OUTPUT=agent.exe

echo Building agent with token: %TOKEN%

REM Configurar entorno
set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=0

REM Compilar con flags para ocultar consola
go build -ldflags "-H windowsgui -s -w -X main.deploymentToken=%TOKEN%" -o %OUTPUT% agent.go

if errorlevel 1 (
    echo Build failed!
    exit /b 1
)

echo Build successful!
echo Output: %OUTPUT%