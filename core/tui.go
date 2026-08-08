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

package core

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

// Exported style/color constants for modernized TUI
const (
	ColorReset        = "\033[0m"
	ColorBold         = "\033[1m"
	ColorDim          = "\033[2m"
	ColorRed          = "\033[31m"
	ColorGreen        = "\033[32m"
	ColorYellow       = "\033[33m"
	ColorBlue         = "\033[34m"
	ColorCyan         = "\033[36m"
	ColorMagenta      = "\033[35m"
	ColorWhite        = "\033[37m"
	ColorLightCyan    = "\033[96m"
	ColorLightGreen   = "\033[92m"
	ColorLightYellow  = "\033[93m"
	ColorLightMagenta = "\033[95m"
	ColorLightRed     = "\033[91m"
	ColorGray         = "\033[90m"
)

// StyleText exports styleText for other packages
func StyleText(text, color string) string {
	return styleText(text, color)
}

// DrawBox returns a formatted string containing content wrapped in a beautiful unicode box.
func DrawBox(title string, content string, borderStyle string) string {
	lines := strings.Split(content, "\n")
	maxWidth := len(title) + 6
	for _, line := range lines {
		trimmedLine := strings.ReplaceAll(line, "\t", "    ")
		// Calculate display length (ignoring ANSI color codes)
		displayLen := len(stripAnsi(trimmedLine))
		if displayLen > maxWidth {
			maxWidth = displayLen
		}
	}

	// Add padding
	width := maxWidth + 4

	var sb strings.Builder
	// Top border
	titleBar := ""
	if title != "" {
		titleBar = " " + StyleText(title, borderStyle+ColorBold) + " "
		rawTitleLen := len(title) + 2
		dashCount := width - rawTitleLen - 2
		if dashCount < 2 {
			dashCount = 2
		}
		sb.WriteString(StyleText("┌", borderStyle))
		sb.WriteString(titleBar)
		sb.WriteString(StyleText(strings.Repeat("─", dashCount), borderStyle))
		sb.WriteString(StyleText("┐\n", borderStyle))
	} else {
		sb.WriteString(StyleText("┌"+strings.Repeat("─", width)+"┐\n", borderStyle))
	}

	// Content lines
	for _, line := range lines {
		trimmedLine := strings.ReplaceAll(line, "\t", "    ")
		displayLen := len(stripAnsi(trimmedLine))
		padding := width - displayLen
		sb.WriteString(StyleText("│ ", borderStyle))
		sb.WriteString(trimmedLine)
		sb.WriteString(strings.Repeat(" ", padding-2))
		sb.WriteString(StyleText(" │\n", borderStyle))
	}

	// Bottom border
	sb.WriteString(StyleText("└"+strings.Repeat("─", width)+"┘", borderStyle))
	return sb.String()
}

// stripAnsi removes ANSI escape sequences to compute length of visible characters
func stripAnsi(str string) string {
	var sb strings.Builder
	inSeq := false
	for i := 0; i < len(str); i++ {
		if str[i] == '\033' {
			inSeq = true
			continue
		}
		if inSeq {
			if (str[i] >= 'a' && str[i] <= 'z') || (str[i] >= 'A' && str[i] <= 'Z') {
				inSeq = false
			}
			continue
		}
		sb.WriteByte(str[i])
	}
	return sb.String()
}

// ShowSpinner runs an animated terminal spinner in the background.
// Stop it by sending true/false to the returned channel.
func ShowSpinner(message string) chan bool {
	stopChan := make(chan bool)
	if IsPiped() {
		fmt.Print(message + "...")
		return stopChan
	}

	go func() {
		frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		i := 0
		for {
			select {
			case <-stopChan:
				// Clear the spinner line
				fmt.Print("\r\033[K")
				return
			default:
				fmt.Printf("\r%s %s", StyleText(frames[i], ColorLightCyan+ColorBold), message)
				i = (i + 1) % len(frames)
				time.Sleep(80 * time.Millisecond)
			}
		}
	}()

	return stopChan
}

// PromptConfirm asks the user for confirmation with a modernized UI
func PromptConfirm(promptText string, defaultVal bool) bool {
	if IsPiped() {
		return defaultVal
	}

	suffix := " [y/N]"
	if defaultVal {
		suffix = " [Y/n]"
	}

	fmt.Printf("\n%s %s %s: ", StyleText("?", ColorYellow+ColorBold), StyleText(promptText, ColorWhite+ColorBold), StyleText(suffix, ColorGray))
	reader := bufio.NewReader(os.Stdin)
	answer, err := reader.ReadString('\n')
	if err != nil {
		return defaultVal
	}
	answer = strings.TrimSpace(strings.ToLower(answer))
	if answer == "" {
		return defaultVal
	}
	if answer == "y" || answer == "yes" {
		return true
	}
	if answer == "n" || answer == "no" {
		return false
	}
	return defaultVal
}
