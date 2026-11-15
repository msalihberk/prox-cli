/* Copyright 2026 Mustafa Salih Berk

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License. */

package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"prox-cli/core"
	"runtime"
	"strings"
	"time"
)

type AICommand struct{}

type GeminiRequest struct {
	Contents          []GeminiContent          `json:"contents"`
	SystemInstruction *GeminiSystemInstruction `json:"systemInstruction,omitempty"`
}

type GeminiSystemInstruction struct {
	Parts []GeminiPart `json:"parts"`
}

type GeminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []GeminiPart `json:"parts"`
}

type GeminiPart struct {
	Text string `json:"text"`
}

type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

const (
	geminiEndpoint         = "https://generativelanguage.googleapis.com/v1beta/models/"
	maxGeminiResponseBytes = 4 << 20
)

func queryGemini(contents []GeminiContent, systemInstruction string) (string, error) {
	return queryGeminiContext(context.Background(), contents, systemInstruction)
}

func queryGeminiContext(ctx context.Context, contents []GeminiContent, systemInstruction string) (string, error) {
	apiKey := os.Getenv("PROX_API_KEY")
	if apiKey == "" {
		return "", errors.New("PROX_API_KEY environment variable is not set. Please set it before using AI features")
	}
	config, err := core.LoadConfig()
	if err != nil {
		return "", err
	}

	reqBody := GeminiRequest{
		Contents: contents,
	}
	if systemInstruction != "" {
		reqBody.SystemInstruction = &GeminiSystemInstruction{
			Parts: []GeminiPart{
				{Text: systemInstruction},
			},
		}
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	requestCtx, cancel := context.WithTimeout(ctx, time.Duration(config.AIRequestTimeoutSeconds)*time.Second)
	defer cancel()

	endpoint := geminiEndpoint + url.PathEscape(config.AIModel) + ":generateContent"
	return sendGeminiRequest(requestCtx, http.DefaultClient, endpoint, apiKey, jsonData)
}

func sendGeminiRequest(ctx context.Context, client *http.Client, endpoint, apiKey string, jsonData []byte) (string, error) {
	requestURL, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}
	query := requestURL.Query()
	query.Set("key", apiKey)
	requestURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL.String(), bytes.NewReader(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return "", fmt.Errorf("AI request canceled: %w", context.Canceled)
		}
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return "", fmt.Errorf("AI request timed out: %w", context.DeadlineExceeded)
		}
		return "", errors.New("AI request failed; check network connectivity and try again")
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, maxGeminiResponseBytes+1))
	if err != nil {
		return "", err
	}
	if len(bodyBytes) > maxGeminiResponseBytes {
		return "", fmt.Errorf("AI response exceeded the %d-byte limit", maxGeminiResponseBytes)
	}

	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case http.StatusUnauthorized, http.StatusForbidden:
			return "", errors.New("AI API rejected the request; check the API key and its permissions")
		case http.StatusTooManyRequests:
			return "", errors.New("AI API rate limit or quota reached; try again later")
		default:
			if resp.StatusCode >= http.StatusInternalServerError {
				return "", fmt.Errorf("AI service returned HTTP %d; try again later", resp.StatusCode)
			}
			return "", fmt.Errorf("AI API request failed with HTTP %d", resp.StatusCode)
		}
	}

	var geminiResp GeminiResponse
	if err := json.Unmarshal(bodyBytes, &geminiResp); err != nil {
		return "", err
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("empty response received from AI model")
	}

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}

func printSlowText(text string) {
	if core.IsPiped() {
		if err := writeAIOutput(text); err != nil {
			core.PrintError("Failed to write AI output: %v", err)
		}
		return
	}
	for _, ch := range text {
		fmt.Printf("%c", ch)
		time.Sleep(1 * time.Millisecond)
	}
	fmt.Println()
}

func formatAIOutput(text string) string {
	return strings.TrimRight(text, "\r\n") + "\n"
}

func writeAIOutput(text string) error {
	_, err := fmt.Fprint(os.Stdout, formatAIOutput(text))
	return err
}

func extractExecutionDirective(text string) (string, bool) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return "", false
	}
	trimmed = strings.TrimSpace(strings.ReplaceAll(trimmed, "\r", ""))
	markers := []string{"PROX_EXECUTE:", "PROX_RUN:", "EXECUTE:"}
	for _, marker := range markers {
		idx := strings.Index(strings.ToUpper(trimmed), marker)
		if idx < 0 {
			continue
		}
		candidate := trimmed[idx+len(marker):]
		candidate = strings.TrimSpace(candidate)
		candidate = strings.Trim(candidate, "`\n\r")
		candidate = strings.TrimPrefix(candidate, "```bash")
		candidate = strings.TrimPrefix(candidate, "```")
		candidate = strings.TrimSuffix(candidate, "```")
		candidate = strings.TrimSpace(candidate)
		if candidate != "" {
			return candidate, true
		}
	}
	return "", false
}

func extractCommandCandidate(text string) (string, bool) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return "", false
	}
	if candidate, ok := extractExecutionDirective(trimmed); ok {
		return candidate, true
	}
	patterns := []string{
		"command:",
		"run:",
		"execute:",
		"here is the command",
		"use this command",
		"recommended command",
	}
	for _, pattern := range patterns {
		if idx := strings.Index(strings.ToLower(trimmed), pattern); idx >= 0 {
			candidate := trimmed[idx+len(pattern):]
			candidate = strings.TrimSpace(candidate)
			candidate = strings.Trim(candidate, "`\n\r")
			candidate = strings.TrimPrefix(candidate, "```bash")
			candidate = strings.TrimPrefix(candidate, "```")
			candidate = strings.TrimSuffix(candidate, "```")
			candidate = strings.TrimSpace(candidate)
			if candidate != "" {
				return candidate, true
			}
		}
	}

	lines := strings.Split(trimmed, "\n")
	for _, line := range lines {
		candidate := strings.TrimSpace(line)
		candidate = strings.Trim(candidate, "`\n\r")
		if candidate == "" {
			continue
		}
		return candidate, true
	}
	return "", false
}

func splitAgentResponse(response string) (string, string, bool) {
	cmd, ok := extractExecutionDirective(response)
	if !ok {
		return response, "", false
	}
	// Let's strip the command instruction lines from response
	lines := strings.Split(response, "\n")
	var explanationLines []string
	markers := []string{"PROX_EXECUTE:", "PROX_RUN:", "EXECUTE:"}
	for _, line := range lines {
		isMarkerLine := false
		for _, m := range markers {
			if strings.Contains(strings.ToUpper(line), m) {
				isMarkerLine = true
				break
			}
		}
		if !isMarkerLine {
			explanationLines = append(explanationLines, line)
		}
	}
	explanation := strings.TrimSpace(strings.Join(explanationLines, "\n"))
	return explanation, cmd, true
}

func runShellCommand(command string) error {
	output, err := executeShellCommand(context.Background(), command)
	if output != "" {
		fmt.Fprintln(os.Stdout, output)
	}
	return err
}

func executeShellCommand(ctx context.Context, command string) (string, error) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "pwsh", "-NoProfile", "-NoLogo", "-Command", command)
	} else {
		cmd = exec.CommandContext(ctx, "bash", "-c", command)
	}
	output, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(output)), err
}

func maybeExecuteSuggestedCommand(response string) {
	if core.IsPiped() {
		return
	}
	explanation, cmd, ok := splitAgentResponse(response)
	if !ok {
		return
	}
	if explanation != "" {
		fmt.Println()
		printSlowText(explanation)
	}
	box := core.DrawBox("PROPOSED COMMAND", cmd, core.ColorLightCyan)
	fmt.Println("\n" + box)
	if !core.PromptConfirm("Run the proposed command?", false) {
		core.PrintWarning("Command execution cancelled by user.")
		return
	}
	core.PrintSuccess("Running command: %s", cmd)
	if err := runShellCommand(cmd); err != nil {
		core.PrintError("Command execution failed: %v", err)
	}
}

func prepareAgentRequest(input string) string {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return trimmed
	}
	lower := strings.ToLower(trimmed)
	if strings.Contains(lower, "run ") || strings.Contains(lower, "execute ") || strings.Contains(lower, "open ") || strings.Contains(lower, "install ") || strings.Contains(lower, "list files") || strings.Contains(lower, "scan ") || strings.Contains(lower, "check ") || strings.Contains(lower, "start ") || strings.Contains(lower, "restart ") || strings.HasPrefix(lower, "prox ") || strings.HasPrefix(lower, "git ") || strings.HasPrefix(lower, "ls ") || strings.HasPrefix(lower, "dir ") {
		return "EXECUTION_REQUEST: " + trimmed
	}
	return trimmed
}

func startInteractiveAgent() error {
	return core.RunAgentTUIWithTools(requestAgentResponse, executeShellCommand, assessCommandRisk, proxAgentTools())
}

func explainCommand(parser *core.Parser) error {
	var inputText string

	projectContext := parser.GetAlias("project").Found
	fileInput := parser.GetAlias("f", "file")
	if projectContext {
		if fileInput.Found {
			return errors.New("use either --project or --file, not both")
		}
		if core.IsPiped() {
			return errors.New("--project requires an interactive terminal so you can review and approve the context")
		}
		stat, err := os.Stdin.Stat()
		if err != nil {
			return fmt.Errorf("inspect standard input: %w", err)
		}
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			return errors.New("--project cannot be combined with piped input")
		}
		positionals := parser.AllPos()
		if len(positionals) < 2 {
			return errors.New("ai explain --project requires a request describing what to analyze")
		}
		projectContextText, err := collectProjectContext()
		if err != nil {
			return err
		}
		approved, err := reviewProjectContext(os.Stdout, projectContextText, func() bool {
			return core.PromptConfirm("Send this project context and your request to Google Gemini?", false)
		})
		if err != nil {
			return fmt.Errorf("show project context preview: %w", err)
		}
		if !approved {
			core.PrintWarning("Project context was not sent.")
			return nil
		}
		inputText = projectExplainInput(strings.Join(positionals[1:], " "), projectContextText)
	} else {
		stat, statErr := os.Stdin.Stat()
		if statErr != nil {
			return fmt.Errorf("inspect standard input: %w", statErr)
		}
		stdinIsPiped := (stat.Mode() & os.ModeCharDevice) == 0
		if stdinIsPiped {
			if fileInput.Found {
				return errors.New("use either --file or piped input, not both")
			}
			input, err := readAIInput(os.Stdin)
			if err != nil {
				return fmt.Errorf("read piped input: %w", err)
			}
			inputText = input
		}

		if fileInput.Found {
			if stdinIsPiped {
				return errors.New("use either --file or piped input, not both")
			}
			source, err := os.Open(fileInput.Value)
			if err != nil {
				return fmt.Errorf("open input file: %w", err)
			}
			defer source.Close()
			inputText, err = readAIInput(source)
			if err != nil {
				return fmt.Errorf("read input file: %w", err)
			}
		}

		if inputText == "" && !stdinIsPiped && !fileInput.Found {
			positionals := parser.AllPos()
			if len(positionals) > 1 {
				inputText = strings.Join(positionals[1:], " ")
			}
		}
	}

	if inputText == "" {
		return errors.New("ai explain requires text, --file <path>, or piped input")
	}

	systemInstruction := "You are a cybersecurity expert and systems developer. Analyze the given log, error message, payload," +
		" or code snippet. Explain what it means, identify any potential security risks or errors, and provide a brief actionable" +
		"solution. Keep it concise. Use only the input supplied by the user; do not request or assume access to other files."

	if !core.IsPiped() {
		core.PrintMessage("Analyzing...")
	}

	result, err := queryGemini([]GeminiContent{
		{
			Parts: []GeminiPart{
				{Text: "Input to analyze:\n" + inputText},
			},
		},
	}, systemInstruction)
	if err != nil {
		return errors.New("failed to analyze input: " + err.Error())
	}

	cleanedResult := strings.TrimSpace(result)

	if core.IsPiped() {
		if err := writeAIOutput(cleanedResult); err != nil {
			return fmt.Errorf("failed to write AI output: %w", err)
		}
	} else {
		core.PrintSuccess("Analysis Result:")
		printSlowText(cleanedResult)
	}
	return nil
}

const maxAIInputBytes = 1 << 20

func readAIInput(reader io.Reader) (string, error) {
	data, err := io.ReadAll(io.LimitReader(reader, maxAIInputBytes+1))
	if err != nil {
		return "", err
	}
	if len(data) > maxAIInputBytes {
		return "", fmt.Errorf("input exceeds the %d-byte limit", maxAIInputBytes)
	}
	return string(data), nil
}
func agentSystemInstruction() string {
	return "You are Prox Agent, a terminal-native developer assistant. " +
		"This is a real command-capable agent session. " +
		"Important: if the user requests a shell command or an action that must be executed, reply with EXACTLY this format: PROX_EXECUTE: <command>. " +
		"If no command should be executed, answer normally and do not include an execution token. " +
		"Keep replies concise, practical, and developer-focused. " +
		"Use the supplied recent conversation context to resolve follow-up questions without repeating prior answers. " +
		"Never include markdown fences when returning a command. " +
		"CRITICAL: The user is running on the following OS: " + runtime.GOOS + ". " +
		"Make sure all generated commands are fully compatible with " + runtime.GOOS + ". " +
		agentToolInstructions(proxAgentTools())
}

func requestAgentResponse(ctx context.Context, userPrompt string, history []core.AgentMessage) (string, error) {
	contents := make([]GeminiContent, 0, len(history)+1)
	for _, message := range history {
		role := "user"
		if message.Role == "assistant" {
			role = "model"
		}
		contents = append(contents, GeminiContent{
			Role:  role,
			Parts: []GeminiPart{{Text: message.Content}},
		})
	}
	if len(history) == 0 || history[len(history)-1].Role != "user" || history[len(history)-1].Content != userPrompt {
		contents = append(contents, GeminiContent{
			Role:  "user",
			Parts: []GeminiPart{{Text: "User Request: " + prepareAgentRequest(userPrompt)}},
		})
	}

	result, err := queryGeminiContext(ctx, contents, agentSystemInstruction())
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(result), nil
}

func assessCommandRisk(ctx context.Context, command string) (core.RiskAssessment, error) {
	systemInstruction := "You are a command risk assessment layer. Analyze the proposed terminal command for data loss, privilege escalation, credential exposure, destructive behavior, network impact, and irreversible changes. " +
		"Return ONLY valid JSON in this exact shape: {\"level\":\"low|medium|high\",\"reason\":\"brief explanation\"}. " +
		"Do not execute the command and do not include markdown."
	result, err := queryGeminiContext(ctx, []GeminiContent{{
		Role:  "user",
		Parts: []GeminiPart{{Text: "Proposed command:\n" + command}},
	}}, systemInstruction)
	if err != nil {
		return core.RiskAssessment{}, err
	}

	cleaned := strings.TrimSpace(result)
	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)
	var assessment core.RiskAssessment
	if err := json.Unmarshal([]byte(cleaned), &assessment); err != nil {
		return core.RiskAssessment{}, fmt.Errorf("invalid risk assessment: %w", err)
	}
	assessment.Level = strings.ToLower(strings.TrimSpace(assessment.Level))
	if assessment.Level != "low" && assessment.Level != "medium" && assessment.Level != "high" {
		return core.RiskAssessment{}, fmt.Errorf("invalid risk level %q", assessment.Level)
	}
	return assessment, nil
}

func TUIChat(userPrompt string) string {
	result, err := requestAgentResponse(context.Background(), userPrompt, nil)
	if err != nil {
		return "Agent error: " + err.Error()
	}

	if !core.IsPiped() {
		core.PrintMessage("Thinking...")
	}

	cleanedResult := strings.TrimSpace(result)
	cleanedResult = strings.TrimPrefix(cleanedResult, "```bash")
	cleanedResult = strings.TrimPrefix(cleanedResult, "```")
	cleanedResult = strings.TrimSuffix(cleanedResult, "```")
	cleanedResult = strings.TrimSpace(cleanedResult)

	return cleanedResult
}
func BasicQuestionCommand(systemInstruction string, userPrompt string) error {

	if !core.IsPiped() {
		core.PrintMessage("Thinking...")
	}

	result, err := queryGemini([]GeminiContent{
		{
			Parts: []GeminiPart{
				{Text: "User Request: " + userPrompt},
			},
		},
	}, systemInstruction)
	if err != nil {
		return err
	}

	cleanedResult := strings.TrimSpace(result)
	cleanedResult = strings.TrimPrefix(cleanedResult, "```bash")
	cleanedResult = strings.TrimPrefix(cleanedResult, "```")
	cleanedResult = strings.TrimSuffix(cleanedResult, "```")
	cleanedResult = strings.TrimSpace(cleanedResult)

	if core.IsPiped() {
		if err := writeAIOutput(cleanedResult); err != nil {
			return fmt.Errorf("failed to write AI output: %w", err)
		}
	} else {
		core.PrintSuccess("AI Response:")
		printSlowText(cleanedResult)
		maybeExecuteSuggestedCommand(cleanedResult)
	}
	return nil
}
func cmdCommand(parser *core.Parser) error {
	prompt, ok := parser.Pos(1)
	if !ok || prompt == "" {
		return errors.New("ai cmd requires a prompt string (e.g., prox ai cmd \"list all files over 50mb\")")
	}

	systemInstruction := "You are a precise CLI assistant. Convert the user request into a single one-liner terminal command. " +
		"Output ONLY the raw executable command, nothing else. No markdown formatting, no code blocks, no explanations, " +
		"no text before or after."
	err := BasicQuestionCommand(systemInstruction, prompt)
	if err != nil {
		return errors.New("failed to generate command: " + err.Error())
	}
	return nil
}
func findCommand(parser *core.Parser) error {
	prompt, ok := parser.Pos(1)
	if !ok || prompt == "" {
		return errors.New("ai find requires a prompt string (e.g., prox ai find \"scan for open ports\")")
	}

	systemInstruction := "You are a CLI assistant. Based on the user request, identify relevant built-in prox modules or commands. " +
		"For each suggestion, give the command, a brief reason it fits, and one accurate example based only on the supplied usage guide. " +
		"Do not invent unsupported flags. If no command fits, say so. This is the tool usage guide: " +
		core.GetAllHelpTexts()
	err := BasicQuestionCommand(systemInstruction, prompt)
	if err != nil {
		return errors.New("failed to find relevant commands: " + err.Error())
	}
	return nil
}
func (v AICommand) Execute(args []string) error {
	parser := core.New(args, false)
	parser.Parse()

	subCommand, ok := parser.Pos(0)
	if !ok || subCommand == "help" || parser.GetAlias("h", "help").Found {
		core.PrintInfo("%s", v.Help())
		return nil
	}

	switch subCommand {
	case "cmd":
		return cmdCommand(parser)
	case "explain":
		return explainCommand(parser)
	case "find":
		return findCommand(parser)
	case "agent":
		return startInteractiveAgent()
	case "sessions":
		return core.CommandRegistry["ai-sessions"].Execute(parser.AllPos()[1:])
	case "status":
		return core.CommandRegistry["ai-config"].Execute(append([]string{"status"}, parser.AllPos()[1:]...))
	case "config":
		return core.CommandRegistry["ai-config"].Execute(parser.AllPos()[1:])
	default:
		return fmt.Errorf("unknown ai sub-command '%s'. Try 'prox ai help' for usage info", subCommand)
	}
}

func (v AICommand) Description() string {
	return "Leverage AI to generate terminal commands, analyze logs, and run an interactive agent session"
}
func (v AICommand) Help() string {
	help := "Usage: prox ai <command> [arguments]"
	help += "\n  cmd <prompt>     : Convert natural language to a one-liner terminal command"
	help += "\n  find <prompt>    : Find builtin prox modules for a specific task or command"
	help += "\n  explain [text]   : Analyze text, --file input, or piped data (maximum 1 MiB)"
	help += "\n  explain --project: Preview and explicitly approve a limited Git/project summary"
	help += "\n                    before sending it to the configured AI provider."
	help += "\n  agent            : Start an interactive agent session"
	help += "\n  sessions         : List, resume, rename, export, or delete agent sessions"
	help += "\n  status           : Show AI configuration and API-key availability"
	help += "\n  config           : Configure the model, timeout, and API key"
	return help
}
func (v AICommand) SubCommands() []string {
	return []string{"cmd", "explain", "find", "agent", "sessions", "status", "config", "--file", "--project", "help"}
}
func init() {
	core.Register("ai", AICommand{})
	core.Register("agent", AgentCommand{})
}

type AgentCommand struct{}

func (a AgentCommand) Execute(args []string) error {
	if len(args) > 0 && (args[0] == "help" || args[0] == "--help" || args[0] == "-h") {
		core.PrintInfo("%s", a.Help())
		return nil
	}
	return startInteractiveAgent()
}

func (a AgentCommand) Description() string {
	return "Start an interactive AI assistant session"
}

func (a AgentCommand) Help() string {
	return "Usage: prox agent\n  Starts an interactive agent chat session that can suggest and run approved commands."
}

func (a AgentCommand) SubCommands() []string {
	return []string{"help"}
}
