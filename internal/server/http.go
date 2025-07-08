package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"neuronc2/internal/agent"
	"neuronc2/internal/auth"
	"neuronc2/internal/database"
	"neuronc2/internal/utils"
	"neuronc2/pkg/models"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (c2 *C2Server) HandleActivation(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.ActivationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Validate token
	token, err := c2.tokenManager.ValidateToken(req.Token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Generate agent credentials
	agentID := fmt.Sprintf("agent-%s", utils.GenerateRandomString(8))
	apiKey := auth.GenerateAPIKey()

	// Create agent record using transaction
	tx, err := c2.db.Begin()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Update token usage
	if err := c2.queries.IncrementTokenUsage(token.ID); err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Create agent record
	agentRecord := &database.AgentRecord{
		AgentID:     agentID,
		APIKey:      apiKey,
		Hostname:    req.Metadata["hostname"],
		Username:    req.Metadata["username"],
		OS:          req.Metadata["os"],
		Arch:        req.Metadata["arch"],
		ActivatedAt: time.Now(),
		LastSeen:    time.Now(),
	}

	if err := c2.queries.CreateAgent(agentRecord); err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Record token usage
	if err := c2.queries.RecordTokenUsage(token.ID, agentID); err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Send response
	resp := models.ActivationResponse{
		AgentID: agentID,
		APIKey:  apiKey,
		Status:  "activated",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

	log.Printf("[+] Agent %s activated with token %s (Hostname: %s, Username: %s)",
		agentID, req.Token[:8]+"...", req.Metadata["hostname"], req.Metadata["username"])
}

func (c2 *C2Server) HandleAgent(w http.ResponseWriter, r *http.Request) {
	// API key validation
	apiKey := r.Header.Get("X-API-Key")
	if apiKey == "" {
		http.Error(w, "Missing API key", http.StatusUnauthorized)
		return
	}

	// Verify agent credentials
	agentRecord, err := c2.queries.GetAgentByAPIKey(apiKey)
	if err == sql.ErrNoRows {
		http.Error(w, "Invalid API key", http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade failed:", err)
		return
	}

	// Receive initial agent information
	var hello map[string]string
	if err := conn.ReadJSON(&hello); err != nil {
		log.Println("Failed to get agent metadata:", err)
		conn.Close()
		return
	}

	// Create agent instance
	agent := agent.New(agentRecord.AgentID, conn, apiKey)
	agent.UpdateInfo(hello["hostname"], hello["username"], hello["os"], hello["arch"])

	// Update last seen timestamp
	c2.queries.UpdateAgentLastSeen(agentRecord.AgentID)

	// Register agent
	c2.AddAgent(agent)

	log.Printf("[+] Agent %s connected (authenticated)\n", agentRecord.AgentID)

	// Handle agent connection
	go func() {
		defer func() {
			c2.RemoveAgent(agentRecord.AgentID)
			agent.Close()
			log.Printf("[-] Agent %s disconnected\n", agentRecord.AgentID)
		}()

		c2.agentHandler.HandleConnection(agent, *agentRecord)
	}()
}
