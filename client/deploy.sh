#!/bin/bash

# Verificar si se proporcionó el token
if [ -z "$1" ]; then
    echo "Uso: ./deploy.sh DEPLOY-TOKEN [output.exe]"
    echo "Ejemplo: ./deploy.sh DEPLOY-abc123def456"
    exit 1
fi

TOKEN=$1
OUTPUT=${2:-agent.exe}

echo "Compilando agente con token: $TOKEN"

# Configurar entorno para cross-compilation a Windows
export GOOS=windows
export GOARCH=amd64
export CGO_ENABLED=0

# Compilar con flags para ocultar consola
go build -ldflags "-H windowsgui -s -w -X main.deploymentToken=$TOKEN" -o $OUTPUT agent.go

if [ $? -ne 0 ]; then
    echo "¡La compilación falló!"
    exit 1
fi

echo "¡Compilación exitosa!"
echo "Output: $OUTPUT" 