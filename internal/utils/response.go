package utils

import (
	"encoding/json"
	"neuronc2/pkg/models"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

func FormatJSONResponse(command string, data interface{}, err error, startTime time.Time) (*mcp.CallToolResult, error) {
	resp := models.MCPResponse{
		Success: err == nil,
		Data:    data,
	}

	resp.Metadata.Timestamp = time.Now()
	resp.Metadata.Command = command
	resp.Metadata.Duration = time.Since(startTime).Milliseconds()

	if err != nil {
		resp.Error = err.Error()
		resp.Message = "Operation failed"
	} else {
		resp.Message = "Operation successful"
	}

	jsonData, _ := json.MarshalIndent(resp, "", "  ")
	return mcp.NewToolResultText(string(jsonData)), nil
}

func Truncate(s string, max int) string {
	if len(s) > max {
		return s[:max] + "..."
	}
	return s
}
