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
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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

func queryGemini(contents []GeminiContent, systemInstruction string) (string, error) {
	apiKey := os.Getenv("PROX_API_KEY")
	if apiKey == "" {
		return "", errors.New("PROX_API_KEY environment variable is not set. Please set it before using AI features")
	}

	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=" + apiKey

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

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
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
		fmt.Print(text)
		return
	}
	for _, ch := range text {
		fmt.Printf("%c", ch)
		time.Sleep(12 * time.Millisecond)
	}
	fmt.Println()
}

func extractExecutionDirective(text string) (string, bool) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return "", false
	}

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
		if candidate != "" && looksLikeCommand(candidate) {
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
			if candidate != "" && looksLikeCommand(candidate) {
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
		if looksLikeCommand(candidate) {
			return candidate, true
		}
	}
	return "", false
}

func looksLikeCommand(text string) bool {
	candidate := strings.TrimSpace(text)
	if candidate == "" {
		return false
	}
	if strings.Contains(strings.ToLower(candidate), "here is") || strings.Contains(strings.ToLower(candidate), "you should") {
		return false
	}
	if strings.Contains(candidate, " ") || strings.Contains(candidate, "\t") {
		return true
	}
	return strings.Contains(candidate, "\\") || strings.Contains(candidate, "/") || strings.HasPrefix(candidate, "prox ") || strings.HasPrefix(candidate, "git ") || strings.HasPrefix(candidate, "ls ") || strings.HasPrefix(candidate, "dir ") || strings.HasPrefix(candidate, "curl ") || strings.HasPrefix(candidate, "wget ")
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
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
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
	// Full screen / Clear screen
	if !core.IsPiped() {
		fmt.Print("\033[H\033[2J")
		fmt.Println(core.GetColorizedBanner())
		fmt.Println()

		// Progressive booting animations
		time.Sleep(150 * time.Millisecond)
		core.PrintInfo("Initializing Prox Copilot Engine...")
		time.Sleep(200 * time.Millisecond)
		core.PrintSuccess("Loaded AI cognitive models successfully.")
		time.Sleep(150 * time.Millisecond)
		core.PrintSuccess("Established connection to generative API service.")
		time.Sleep(200 * time.Millisecond)
		core.PrintInfo("Agent Safe Mode: Active. Command execution requires explicit approval.")
		fmt.Println()
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("%s %s ", core.StyleText("prox-agent", core.ColorLightCyan+core.ColorBold), core.StyleText("›", core.ColorLightGreen+core.ColorBold))
		input, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println()
				return nil
			}
			return err
		}
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}
		if strings.EqualFold(input, "exit") || strings.EqualFold(input, "quit") {
			core.PrintInfo("Agent session ended. Goodbye!")
			return nil
		}

		systemInstruction := "You are Prox Agent, a terminal-native developer assistant. " +
			"This is a real command-capable agent session. " +
			"Important: if the user requests a shell command or an action that must be executed, reply with EXACTLY this format: PROX_EXECUTE: <command>. " +
			"If no command should be executed, answer normally and do not include an execution token. " +
			"Keep replies concise, practical, and developer-focused. " +
			"Never include markdown fences when returning a command. " +
			"CRITICAL: The user is running on the following OS: " + runtime.GOOS + ". " +
			"Make sure all generated commands are fully compatible with " + runtime.GOOS + "."

		spinnerStop := core.ShowSpinner("Agent is thinking")
		result, err := queryGemini([]GeminiContent{
			{
				Parts: []GeminiPart{
					{Text: "User request: " + prepareAgentRequest(input)},
				},
			},
		}, systemInstruction)
		spinnerStop <- true
		if err != nil {
			core.PrintError("%v", err)
			continue
		}

		cleaned := strings.TrimSpace(result)
		if cleaned == "" {
			continue
		}

		explanation, cmd, hasCmd := splitAgentResponse(cleaned)
		if hasCmd {
			if explanation != "" {
				fmt.Println()
				printSlowText(explanation)
			}
			box := core.DrawBox("PROPOSED COMMAND", cmd, core.ColorLightCyan)
			fmt.Println("\n" + box)
			if core.PromptConfirm("Run the proposed command?", false) {
				core.PrintSuccess("Running command: %s", cmd)
				if err := runShellCommand(cmd); err != nil {
					core.PrintError("Command execution failed: %v", err)
				}
			} else {
				core.PrintWarning("Command execution cancelled by user.")
			}
		} else {
			fmt.Println()
			printSlowText(cleaned)
		}
	}
}

func explainCommand(parser *core.Parser) error {
	var inputText string

	if core.IsPiped() {
		stat, err := os.Stdin.Stat()
		if err == nil && (stat.Mode()&os.ModeCharDevice) == 0 {
			bytesData, err := io.ReadAll(os.Stdin)
			if err == nil {
				inputText = string(bytesData)
			}
		}
	}

	if inputText == "" {
		prompt, ok := parser.Pos(1)
		if ok && prompt != "" {
			inputText = prompt
		}
	}

	if inputText == "" {
		return errors.New("ai explain requires an input string or piped data stream")
	}

	systemInstruction := "You are a cybersecurity expert and systems developer. Analyze the given log, error message, payload," +
		" or code snippet. Explain what it means, identify any potential security risks or errors, and provide a brief actionable" +
		"solution. Keep it concise."

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
		fmt.Print(cleanedResult)
	} else {
		core.PrintSuccess("Analysis Result:")
		printSlowText(cleanedResult)
	}
	return nil
}
func basicQuestionCommand(systemInstruction string, userPrompt string) error {

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
		fmt.Print(cleanedResult)
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
	err := basicQuestionCommand(systemInstruction, prompt)
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

	systemInstruction := "You are a CLI assistant. Based on the user request, identify and suggest relevant " +
		"built-in prox modules or commands that can accomplish the task. Output ONLY the names of the modules or commands " +
		"and necessary arguments, nothing else. No explanations, no text before or after.If you don't know, say so. This is tool's usage guide: " +
		core.GetAllHelpTexts()
	err := basicQuestionCommand(systemInstruction, prompt)
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
	help += "\n  explain [text]   : Analyze and explain logs, code, or payloads (Supports piping)"
	help += "\n  agent            : Start an interactive agent session"
	return help
}
func (v AICommand) SubCommands() []string {
	return []string{"cmd", "explain", "find", "agent", "help"}
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
