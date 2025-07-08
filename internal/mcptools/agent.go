package mcptools

import (
	"context"
	"fmt"
	"neuronc2/internal/utils"
	"neuronc2/pkg/models"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

func ListAgents(c2 ServerInterface) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		startTime := time.Now()

		agents := c2.GetAllAgents()
		var agentList []models.AgentJSON

		for id, agent := range agents {
			agentList = append(agentList, models.AgentJSON{
				ID:       id,
				Hostname: agent.Hostname,
				Username: agent.Username,
				OS:       agent.OS,
				Arch:     agent.Arch,
				Status:   "connected",
				LastSeen: agent.LastSeen,
			})
		}

		data := map[string]interface{}{
			"agents": agentList,
			"count":  len(agentList),
		}

		return utils.FormatJSONResponse("list_agents", data, nil, startTime)
	}
}

func ListAllAgents(c2 ServerInterface) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		startTime := time.Now()

		queries := c2.GetQueries()
		agents, err := queries.GetAllAgents()
		if err != nil {
			return utils.FormatJSONResponse("list_all_agents", nil, err, startTime)
		}

		connectedAgents := c2.GetAllAgents()

		// Enrich with connection status
		for i := range agents {
			if _, connected := connectedAgents[agents[i].AgentID]; connected {
				agents[i].Status = "connected"
			} else {
				agents[i].Status = "disconnected"
			}
		}

		data := map[string]interface{}{
			"agents": agents,
			"count":  len(agents),
			"summary": map[string]int{
				"total":        len(agents),
				"connected":    len(connectedAgents),
				"disconnected": len(agents) - len(connectedAgents),
			},
		}

		return utils.FormatJSONResponse("list_all_agents", data, nil, startTime)
	}
}

func GetAgentInfo(c2 ServerInterface) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		startTime := time.Now()

		agentID, ok := req.Params.Arguments["agent_id"].(string)
		if !ok {
			return utils.FormatJSONResponse("get_agent_info", nil, fmt.Errorf("agent_id parameter missing"), startTime)
		}

		queries := c2.GetQueries()
		agent, err := queries.GetAgentInfo(agentID)
		if err != nil {
			return utils.FormatJSONResponse("get_agent_info", nil, err, startTime)
		}

		connectedAgents := c2.GetAllAgents()
		if _, connected := connectedAgents[agentID]; connected {
			agent.Status = "connected"
		} else {
			agent.Status = "disconnected"
		}

		data := map[string]interface{}{
			"agent": agent,
		}

		return utils.FormatJSONResponse("get_agent_info", data, nil, startTime)
	}
}
