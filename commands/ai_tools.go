package commands

import (
	"strings"

	"prox-cli/core"
)

func proxAgentTools() []core.AgentTool {
	configCommand := AIConfigCommand{}
	sessionsCommand := AISessionsCommand{}
	return []core.AgentTool{
		{
			Name:        "ai-config",
			Description: configCommand.Description(),
			Usage:       "status | model <name> | timeout <seconds> | key remove",
			RequiresApproval: func(args []string) bool {
				if len(args) == 0 {
					return false
				}
				return args[0] != "status"
			},
			Execute: configCommand.ExecuteAgentTool,
		},
		{
			Name:        "ai-sessions",
			Description: sessionsCommand.Description(),
			Usage:       "list | rename <id> <name> | export <id> <file> | delete <id> | clear",
			RequiresApproval: func(args []string) bool {
				return sessionActionRequiresApproval(args)
			},
			Execute: sessionsCommand.ExecuteAgentTool,
		},
	}
}

func agentToolInstructions(tools []core.AgentTool) string {
	var instructions strings.Builder
	instructions.WriteString("Available local Prox tools:\n")
	for _, tool := range tools {
		instructions.WriteString("- ")
		instructions.WriteString(tool.Name)
		instructions.WriteString(": ")
		instructions.WriteString(tool.Description)
		instructions.WriteString(". Supported actions: ")
		instructions.WriteString(tool.Usage)
		instructions.WriteString(".\n")
	}
	instructions.WriteString("When a built-in Prox tool is the best way to satisfy the user's request, respond with one exact line: PROX_TOOL: <tool-name> <action and arguments>. Put any brief explanation on separate lines. Never invent tool names or actions. Only use tools when the user requested the operation. Read-only status/list actions may run directly; all settings changes, exports, renames, and deletions require explicit TUI approval. API-key setup is not available as an agent tool; tell the user to run 'prox ai-config key set' directly. Shell commands must continue to use PROX_EXECUTE and require their separate review.")
	return instructions.String()
}
