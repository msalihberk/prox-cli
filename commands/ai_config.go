package commands

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"prox-cli/core"
)

type AIConfigCommand struct{}

func (c AIConfigCommand) Execute(args []string) error {
	parser := core.New(args, false)
	if err := parser.Parse(); err != nil {
		return err
	}
	args = parser.AllPos()
	if len(args) == 0 || args[0] == "help" {
		_, err := fmt.Fprintln(os.Stdout, c.Help())
		return err
	}
	return executeAIConfig(args, os.Stdout, false, true)
}

func (c AIConfigCommand) ExecuteAgentTool(_ context.Context, args []string, approved bool) (string, error) {
	if aiConfigActionRequiresApproval(args) && !approved {
		return "", errors.New("this AI configuration action requires user approval")
	}
	var output bytes.Buffer
	if err := executeAIConfig(args, &output, approved, false); err != nil {
		return "", err
	}
	return strings.TrimSpace(output.String()), nil
}

func aiConfigActionRequiresApproval(args []string) bool {
	if len(args) == 0 {
		return false
	}
	switch args[0] {
	case "model", "timeout":
		return true
	case "key":
		return len(args) > 1 && args[1] == "remove"
	default:
		return false
	}
}

func executeAIConfig(args []string, output io.Writer, approved, allowInteractiveKeySetup bool) error {
	if len(args) == 0 || args[0] == "help" {
		_, err := fmt.Fprintln(output, (AIConfigCommand{}).Help())
		return err
	}
	if args[0] == "status" {
		return showAIStatus(args[1:], output)
	}
	return configureAI(args, output, approved, allowInteractiveKeySetup)
}

func (c AIConfigCommand) Description() string {
	return "Inspect AI configuration and manage model, timeout, and API key settings"
}

func (c AIConfigCommand) Help() string {
	return "Usage: prox ai config <status|model|timeout|key> [value]\n" +
		"  status             Show AI configuration without revealing the API key\n" +
		"  model <name>       Set the Gemini model name\n" +
		"  timeout <seconds>  Set request timeout (5-120 seconds)\n" +
		"  key set            Save PROX_API_KEY for your user\n" +
		"  key remove         Remove the saved PROX_API_KEY (asks for confirmation)\n" +
		"API keys are not written to the Prox config file."
}

func (c AIConfigCommand) SubCommands() []string {
	return []string{"status", "model", "timeout", "key", "help"}
}

func showAIStatus(args []string, output io.Writer) error {
	if len(args) != 0 {
		return errors.New("usage: prox ai status")
	}
	config, err := core.LoadConfig()
	if err != nil {
		return err
	}
	keyStatus := "not configured"
	if os.Getenv("PROX_API_KEY") != "" {
		keyStatus = "configured (value hidden)"
	}
	store, err := core.NewHistoryStore()
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(output,
		"AI provider: Google Gemini\nModel: %s\nRequest timeout: %d seconds\nAPI key: %s\nSession history: %s\nProject context: not collected automatically\n",
		config.AIModel,
		config.AIRequestTimeoutSeconds,
		keyStatus,
		store.Path,
	)
	return err
}

func configureAI(args []string, output io.Writer, approved, allowInteractiveKeySetup bool) error {
	if len(args) == 0 || args[0] == "help" {
		return errors.New("usage: prox ai config <model|timeout|key> <value>")
	}
	if args[0] == "key" {
		if len(args) != 2 {
			return errors.New("usage: prox ai config key <set|remove>")
		}
		switch args[1] {
		case "set":
			if !allowInteractiveKeySetup {
				return errors.New("API key setup is interactive; run 'prox ai-config key set' directly")
			}
			return core.SetupEnvironment()
		case "remove":
			if !approved && !core.PromptConfirm("Remove the saved PROX_API_KEY?", false) {
				_, err := fmt.Fprintln(output, "API key removal cancelled")
				return err
			}
			if err := core.RemoveAPIKey(); err != nil {
				return err
			}
			_, err := fmt.Fprintln(output, "Saved PROX_API_KEY removed. The current shell may retain its inherited value; unset it there or restart the terminal.")
			return err
		default:
			return errors.New("usage: prox ai config key <set|remove>")
		}
	}

	config, err := core.LoadConfig()
	if err != nil {
		return err
	}
	switch args[0] {
	case "model":
		if len(args) != 2 || !validAIModel(args[1]) {
			return errors.New("usage: prox ai config model <name> (letters, numbers, '.', '_' and '-' only)")
		}
		config.AIModel = args[1]
	case "timeout":
		if len(args) != 2 {
			return errors.New("usage: prox ai config timeout <seconds> (5-120)")
		}
		seconds, err := parseAIRequestTimeout(args[1])
		if err != nil {
			return err
		}
		config.AIRequestTimeoutSeconds = seconds
	default:
		return fmt.Errorf("unknown AI setting %q; run 'prox ai config help' for usage", args[0])
	}
	if err := core.SaveConfig(config); err != nil {
		return err
	}
	_, err = fmt.Fprintf(output, "AI %s updated\n", args[0])
	return err
}

func parseAIRequestTimeout(value string) (int, error) {
	seconds, err := strconv.Atoi(value)
	if err != nil || seconds < 5 || seconds > 120 {
		return 0, errors.New("AI request timeout must be between 5 and 120 seconds")
	}
	return seconds, nil
}

func validAIModel(model string) bool {
	if model == "" || len(model) > 128 {
		return false
	}
	for _, char := range model {
		if (char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			strings.ContainsRune("._-", char) {
			continue
		}
		return false
	}
	return true
}

func init() {
	core.Register("ai-config", AIConfigCommand{})
}
