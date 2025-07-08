package mcptools

import (
	"context"
	"neuronc2/internal/utils"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

func GetDatabaseStats(c2 ServerInterface) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		startTime := time.Now()

		queries := c2.GetQueries()
		stats, err := queries.GetDatabaseStats()
		if err != nil {
			return utils.FormatJSONResponse("get_database_stats", nil, err, startTime)
		}

		// Add connected agents count
		stats["connected_agents"] = len(c2.GetAllAgents())

		return utils.FormatJSONResponse("get_database_stats", stats, nil, startTime)
	}
}
