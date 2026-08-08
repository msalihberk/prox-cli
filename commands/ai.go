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
	Contents []GeminiContent `json:"contents"`
}

type GeminiContent struct {
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

func queryGemini(prompt string) (string, error) {
	apiKey := os.Getenv("PROX_API_KEY")
	if apiKey == "" {
		return "", errors.New("PROX_API_KEY environment variable is not set. Please set it before using AI features")
	}

	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=" + apiKey

	reqBody := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{Text: prompt},
				},
			},
		},
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

func extractCommandCandidate(text string) (string, bool) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return "", false
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

func promptForConfirmation(promptText string) bool {
	if core.IsPiped() {
		return false
	}
	fmt.Printf("\n%s [Y/n]: ", promptText)
	reader := bufio.NewReader(os.Stdin)
	answer, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return false
	}
	answer = strings.TrimSpace(strings.ToLower(answer))
	return answer == "" || answer == "y" || answer == "yes"
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
	candidate, ok := extractCommandCandidate(response)
	if !ok {
		return
	}
	if !promptForConfirmation("AI suggested a command to run") {
		core.PrintWarning("Command execution cancelled by user.")
		return
	}
	core.PrintSuccess("Running command: %s", candidate)
	if err := runShellCommand(candidate); err != nil {
		core.PrintError("Command execution failed: %v", err)
	}
}

func startInteractiveAgent() error {
	core.PrintSuccess("Starting interactive agent session")
	core.PrintInfo("Type your request and press Enter. Use 'exit' or 'quit' to leave.")
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("\nagent> ")
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
			core.PrintInfo("Agent session ended.")
			return nil
		}

		systemInstruction := "You are an expert terminal copilot and developer assistant. Help the user with practical, concise answers. If a shell command is useful, provide it in a clear command form. Ask for confirmation before running it."
		result, err := queryGemini(systemInstruction + "\n\nUser request: " + input)
		if err != nil {
			core.PrintError("%v", err)
			continue
		}
		cleaned := strings.TrimSpace(result)
		if cleaned == "" {
			continue
		}
		core.PrintSuccess("Agent response:")
		printSlowText(cleaned)
		maybeExecuteSuggestedCommand(cleaned)
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
	fullPrompt := systemInstruction + "\n\nInput to analyze:\n" + inputText

	if !core.IsPiped() {
		core.PrintMessage("Analyzing...")
	}

	result, err := queryGemini(fullPrompt)
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
	fullPrompt := systemInstruction + "\n\nUser Request: " + userPrompt

	if !core.IsPiped() {
		core.PrintMessage("Thinking...")
	}

	result, err := queryGemini(fullPrompt)
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
