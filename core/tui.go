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

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type AgentMessage struct {
	Role    string
	Content string
}

type AgentResponder func(string, []AgentMessage) (string, error)

type CommandExecutor func(string) (string, error)

type RiskAssessment struct {
	Level  string
	Reason string
}

type CommandRiskAssessor func(string) (RiskAssessment, error)

var (
	tuiAccentColor  = lipgloss.Color("14")
	tuiFocusColor   = lipgloss.Color("205")
	tuiMutedColor   = lipgloss.Color("241")
	tuiSuccessColor = lipgloss.Color("42")
	tuiWarningColor = lipgloss.Color("196")

	tuiHeaderStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Background(tuiAccentColor).Padding(0, 2)
	tuiChatStyle    = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("36")).Padding(0, 1)
	tuiInputStyle   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(tuiFocusColor).Padding(0, 1)
	tuiConfirmStyle = lipgloss.NewStyle().Border(lipgloss.DoubleBorder()).BorderForeground(tuiWarningColor).Padding(1, 2).Align(lipgloss.Center)
	tuiYesStyle     = lipgloss.NewStyle().Background(tuiSuccessColor).Foreground(lipgloss.Color("0")).Bold(true).Padding(0, 1)
	tuiNoStyle      = lipgloss.NewStyle().Background(tuiWarningColor).Foreground(lipgloss.Color("15")).Bold(true).Padding(0, 1)
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

type agentTUIModel struct {
	viewport    viewport.Model
	textarea    textarea.Model
	width       int
	height      int
	ready       bool
	showConfirm bool
	pendingCmd  string
	pendingRisk RiskAssessment
	messages    []string
	history     []AgentMessage
	status      string
	busy        bool
	animation   int
	respond     AgentResponder
	execute     CommandExecutor
	assessRisk  CommandRiskAssessor
}

type agentResponseMsg struct {
	response string
	err      error
}

type commandResultMsg struct {
	command string
	output  string
	err     error
}

type riskAssessmentMsg struct {
	command    string
	assessment RiskAssessment
	err        error
}

type agentTickMsg struct{}

const maxAgentContextChars = 6000

func newAgentTUIModel(respond AgentResponder, execute CommandExecutor, assessRisk CommandRiskAssessor) agentTUIModel {
	input := textarea.New()
	input.Focus()
	input.Prompt = "│ "
	input.CharLimit = 500
	input.SetHeight(2)
	input.ShowLineNumbers = false

	return agentTUIModel{
		textarea:   input,
		messages:   []string{"Prox CLI AI Agent Core Running..."},
		status:     "IDLE",
		respond:    respond,
		execute:    execute,
		assessRisk: assessRisk,
	}
}

func (m agentTUIModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m *agentTUIModel) refreshAgentMessages() {
	if m.ready {
		m.viewport.SetContent(wrapAgentMessages(m.messages, maxInt(10, m.viewport.Width-2)))
		m.viewport.GotoBottom()
	}
}

func (m agentTUIModel) nextAgentTick() tea.Cmd {
	return tea.Tick(350*time.Millisecond, func(time.Time) tea.Msg {
		return agentTickMsg{}
	})
}

func (m agentTUIModel) requestAgent(prompt string) tea.Cmd {
	return func() tea.Msg {
		response, err := m.respond(prompt, m.contextSnapshot())
		return agentResponseMsg{response: response, err: err}
	}
}

func (m agentTUIModel) assessAgentCommand(command string) tea.Cmd {
	return func() tea.Msg {
		assessment, err := m.assessRisk(command)
		return riskAssessmentMsg{command: command, assessment: assessment, err: err}
	}
}

func (m *agentTUIModel) addAgentHistory(role, content string) {
	if strings.TrimSpace(content) == "" {
		return
	}
	m.history = append(m.history, AgentMessage{Role: role, Content: content})
	for contextCharacterCount(m.history) > maxAgentContextChars && len(m.history) > 1 {
		m.history = m.history[1:]
	}
}

func (m agentTUIModel) contextSnapshot() []AgentMessage {
	context := make([]AgentMessage, len(m.history))
	copy(context, m.history)
	return context
}

func contextCharacterCount(messages []AgentMessage) int {
	count := 0
	for _, message := range messages {
		count += len([]rune(message.Role)) + len([]rune(message.Content))
	}
	return count
}

func (m agentTUIModel) executeAgentCommand(command string) tea.Cmd {
	return func() tea.Msg {
		output, err := m.execute(command)
		return commandResultMsg{command: command, output: output, err: err}
	}
}

func (m agentTUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var inputCmd tea.Cmd
	var viewportCmd tea.Cmd

	switch message := msg.(type) {
	case agentTickMsg:
		if !m.busy {
			return m, nil
		}
		m.animation = (m.animation + 1) % 4
		return m, m.nextAgentTick()

	case agentResponseMsg:
		m.busy = false
		if message.err != nil {
			m.messages = append(m.messages, fmt.Sprintf("%s %s", lipgloss.NewStyle().Bold(true).Foreground(tuiWarningColor).Render("ERROR"), message.err.Error()))
			m.status = "ERROR"
		} else {
			explanation, command, hasCommand := splitTUIAgentResponse(message.response)
			m.addAgentHistory("assistant", message.response)
			if explanation != "" {
				m.messages = append(m.messages, fmt.Sprintf("%s %s", lipgloss.NewStyle().Bold(true).Foreground(tuiAccentColor).Render("AGENT"), explanation))
			}
			if hasCommand {
				m.pendingCmd = command
				m.busy = true
				m.status = "ASSESSING RISK"
				m.refreshAgentMessages()
				return m, tea.Batch(m.assessAgentCommand(command), m.nextAgentTick())
			} else {
				m.status = "READY"
			}
		}
		m.refreshAgentMessages()
		return m, nil

	case riskAssessmentMsg:
		m.busy = false
		if message.err != nil {
			m.messages = append(m.messages, fmt.Sprintf("%s Risk assessment failed: %s", lipgloss.NewStyle().Bold(true).Foreground(tuiWarningColor).Render("BLOCKED"), message.err.Error()))
			m.status = "BLOCKED"
			m.pendingCmd = ""
		} else {
			m.pendingRisk = message.assessment
			m.showConfirm = true
			m.status = "REVIEW REQUIRED"
		}
		m.refreshAgentMessages()
		return m, nil

	case commandResultMsg:
		m.busy = false
		if message.err != nil {
			m.messages = append(m.messages, fmt.Sprintf("%s `%s`: %s", lipgloss.NewStyle().Bold(true).Foreground(tuiWarningColor).Render("FAILED"), message.command, message.err.Error()))
			m.status = "ERROR"
		} else {
			result := fmt.Sprintf("%s `%s`", lipgloss.NewStyle().Bold(true).Foreground(tuiSuccessColor).Render("EXECUTED"), message.command)
			if message.output != "" {
				result += "\n" + message.output
			}
			m.messages = append(m.messages, result)
			m.addAgentHistory("tool", "Command output: "+message.output)
			m.status = "READY"
		}
		m.refreshAgentMessages()
		return m, nil

	case tea.WindowSizeMsg:
		m.width = message.Width
		m.height = message.Height
		chatWidth := maxInt(20, m.width-32)
		contentHeight := maxInt(5, m.height-11)
		if !m.ready {
			m.viewport = viewport.New(chatWidth, contentHeight)
			m.viewport.SetContent(wrapAgentMessages(m.messages, maxInt(10, m.viewport.Width-2)))
			m.ready = true
		} else {
			m.viewport.Width = chatWidth
			m.viewport.Height = contentHeight
			m.refreshAgentMessages()
		}
		m.textarea.SetWidth(maxInt(10, m.width-4))

	case tea.KeyMsg:
		if m.showConfirm {
			switch strings.ToLower(message.String()) {
			case "y":
				command := m.pendingCmd
				m.showConfirm = false
				m.pendingCmd = ""
				m.busy = true
				m.status = "EXECUTING"
				m.messages = append(m.messages, fmt.Sprintf("Running `%s`...", command))
				m.addAgentHistory("tool", "Command approved and executed: "+command)
				m.refreshAgentMessages()
				return m, tea.Batch(m.executeAgentCommand(command), m.nextAgentTick())
			case "n", "esc":
				m.showConfirm = false
				m.messages = append(m.messages, fmt.Sprintf("%s `%s`", lipgloss.NewStyle().Bold(true).Foreground(tuiAccentColor).Render("CANCELLED"), m.pendingCmd))
				m.pendingCmd = ""
				m.pendingRisk = RiskAssessment{}
				m.status = "CANCELLED"
				m.addAgentHistory("tool", "Command rejected by user.")
				m.refreshAgentMessages()
			}
			return m, nil
		}

		switch message.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			if m.busy {
				return m, nil
			}
			prompt := strings.TrimSpace(m.textarea.Value())
			if prompt == "" {
				return m, nil
			}
			m.messages = append(m.messages, fmt.Sprintf("%s %s", lipgloss.NewStyle().Bold(true).Foreground(tuiAccentColor).Render("USER"), prompt))
			m.addAgentHistory("user", prompt)
			m.textarea.Reset()
			m.busy = true
			m.status = "PROCESSING"
			m.refreshAgentMessages()
			return m, tea.Batch(m.requestAgent(prompt), m.nextAgentTick())
		}
	}

	m.textarea, inputCmd = m.textarea.Update(msg)
	m.viewport, viewportCmd = m.viewport.Update(msg)
	return m, tea.Batch(inputCmd, viewportCmd)
}

func (m agentTUIModel) View() string {
	if !m.ready {
		return "\n  Initializing Prox Core Engine..."
	}

	header := tuiHeaderStyle.Width(maxInt(10, m.width)).Render("PROX AGENT")
	content := tuiChatStyle.Width(maxInt(10, m.width-4)).Height(m.viewport.Height).Render(m.viewport.View())

	if m.showConfirm {
		command := wrapAgentMessage(m.pendingCmd, maxInt(10, m.width-16))
		level := strings.ToUpper(m.pendingRisk.Level)
		if level == "" {
			level = "UNKNOWN"
		}
		reason := m.pendingRisk.Reason
		if reason == "" {
			reason = "No additional risk details were provided."
		}
		modal := tuiConfirmStyle.Width(maxInt(10, m.width-10)).Render(fmt.Sprintf("COMMAND REVIEW [%s]\n\nCommand:\n%s\n\nRisk: %s\n%s\n\n%s    %s", level, command, level, wrapAgentMessage(reason, maxInt(10, m.width-16)), tuiYesStyle.Render("[Y] Confirm"), tuiNoStyle.Render("[N] Abort")))
		return lipgloss.JoinVertical(lipgloss.Left, header, content, modal)
	}

	status := m.status
	if m.busy {
		status = "THINKING" + strings.Repeat(".", m.animation)
	}
	help := lipgloss.NewStyle().Foreground(tuiMutedColor).Render(fmt.Sprintf(" %s | Enter send | Esc quit", status))
	return lipgloss.JoinVertical(lipgloss.Left, header, content, tuiInputStyle.Render(m.textarea.View()), help)
}

func splitTUIAgentResponse(response string) (string, string, bool) {
	trimmed := strings.TrimSpace(strings.ReplaceAll(response, "\r", ""))
	markers := []string{"PROX_EXECUTE:", "PROX_RUN:", "EXECUTE:"}
	command := ""
	markerLine := ""
	for _, marker := range markers {
		if index := strings.Index(strings.ToUpper(trimmed), marker); index >= 0 {
			markerLine = marker
			command = strings.TrimSpace(trimmed[index+len(marker):])
			break
		}
	}
	if command == "" {
		return trimmed, "", false
	}
	command = strings.Trim(strings.TrimSpace(command), "`\n\r")
	command = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(command, "```bash"), "```"))
	command = strings.TrimSuffix(command, "```")
	if command == "" {
		return trimmed, "", false
	}

	var explanation []string
	for _, line := range strings.Split(trimmed, "\n") {
		if !strings.Contains(strings.ToUpper(line), markerLine) {
			explanation = append(explanation, line)
		}
	}
	return strings.TrimSpace(strings.Join(explanation, "\n")), command, true
}

func wrapAgentMessages(messages []string, width int) string {
	wrapped := make([]string, 0, len(messages))
	for _, message := range messages {
		wrapped = append(wrapped, wrapAgentMessage(message, width))
	}
	return strings.Join(wrapped, "\n\n")
}

func wrapAgentMessage(message string, width int) string {
	var lines []string
	for _, paragraph := range strings.Split(message, "\n") {
		words := strings.Fields(paragraph)
		if len(words) == 0 {
			lines = append(lines, "")
			continue
		}

		current := ""
		for _, word := range words {
			if lipgloss.Width(word) > width {
				if current != "" {
					lines = append(lines, current)
					current = ""
				}
				for lipgloss.Width(word) > width {
					part, rest := splitAgentWord(word, width)
					lines = append(lines, part)
					word = rest
				}
				if word != "" {
					current = word
				}
				continue
			}

			candidate := word
			if current != "" {
				candidate = current + " " + word
			}
			if lipgloss.Width(candidate) > width && current != "" {
				lines = append(lines, current)
				current = word
			} else {
				current = candidate
			}
		}
		if current != "" {
			lines = append(lines, current)
		}
	}
	return strings.Join(lines, "\n")
}

func splitAgentWord(word string, width int) (string, string) {
	runes := []rune(word)
	for index := 1; index <= len(runes); index++ {
		if lipgloss.Width(string(runes[:index])) > width {
			return string(runes[:index-1]), string(runes[index-1:])
		}
	}
	return word, ""
}

func maxInt(left, right int) int {
	if left > right {
		return left
	}
	return right
}

func RunAgentTUI(respond AgentResponder, execute CommandExecutor, assessRisk CommandRiskAssessor) error {
	if respond == nil {
		return fmt.Errorf("agent responder is required")
	}
	if execute == nil {
		return fmt.Errorf("command executor is required")
	}
	if assessRisk == nil {
		return fmt.Errorf("command risk assessor is required")
	}
	program := tea.NewProgram(newAgentTUIModel(respond, execute, assessRisk), tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		return fmt.Errorf("failed to start TUI: %w", err)
	}
	return nil
}
