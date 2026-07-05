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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
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
func explainCommand(parser *Parser) error {
	var inputText string

	if isPiped() {
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

	if !isPiped() {
		PrintMessage("Analyzing...")
	}

	result, err := queryGemini(fullPrompt)
	if err != nil {
		return err
	}

	cleanedResult := strings.TrimSpace(result)

	if isPiped() {
		fmt.Print(cleanedResult)
	} else {
		PrintSuccess("Analysis Result:")
		PrintMessage("%s", cleanedResult)
	}
	return nil
}
func basicQuestionCommand(systemInstruction string, userPrompt string) error {
	fullPrompt := systemInstruction + "\n\nUser Request: " + userPrompt

	if !isPiped() {
		PrintMessage("Thinking...")
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

	if isPiped() {
		fmt.Print(cleanedResult)
	} else {
		PrintSuccess("Command:")
		PrintInfo("  %s", cleanedResult)
	}
	return nil
}
func cmdCommand(parser *Parser) error {
	prompt, ok := parser.Pos(1)
	if !ok || prompt == "" {
		return errors.New("ai cmd requires a prompt string (e.g., prox ai cmd \"list all files over 50mb\")")
	}

	systemInstruction := "You are a precise CLI assistant. Convert the user request into a single one-liner terminal command. " +
		"Output ONLY the raw executable command, nothing else. No markdown formatting, no code blocks, no explanations, " +
		"no text before or after."
	basicQuestionCommand(systemInstruction, prompt)
	return nil
}
func findCommand(parser *Parser) error {
	prompt, ok := parser.Pos(1)
	if !ok || prompt == "" {
		return errors.New("ai find requires a prompt string (e.g., prox ai find \"scan for open ports\")")
	}

	systemInstruction := "You are a CLI assistant. Based on the user request, identify and suggest relevant " +
		"built-in prox modules or commands that can accomplish the task. Output ONLY the names of the modules or commands " +
		"and necessary arguments, nothing else. No explanations, no text before or after.If you don't know, say so. This is tool's usage guide: " +
		GetAllHelpTexts()
	basicQuestionCommand(systemInstruction, prompt)
	return nil
}
func (v AICommand) Execute(args []string) error {
	parser := New(args, false)
	parser.Parse()

	subCommand, ok := parser.Pos(0)
	if !ok || subCommand == "help" || parser.GetAlias("h", "help").Found {
		PrintInfo("%s", v.Help())
		return nil
	}

	switch subCommand {
	case "cmd":
		return cmdCommand(parser)
	case "explain":
		return explainCommand(parser)

	case "find":
		return findCommand(parser)
	default:
		return fmt.Errorf("unknown ai sub-command '%s'. Try 'prox ai help' for usage info", subCommand)
	}
}

func (v AICommand) Description() string {
	return "Leverage AI to generate terminal commands or analyze complex logs and payloads"
}
func (v AICommand) Help() string {
	help := "Usage: prox ai <command> [arguments]"
	help += "\n  cmd <prompt>     : Convert natural language to a one-liner terminal command"
	help += "\n  find <prompt>    : Find builtin prox modules for a specific task or command"
	help += "\n  explain [text]   : Analyze and explain logs, code, or payloads (Supports piping)"
	return help
}

func init() {
	register("ai", AICommand{})
}
