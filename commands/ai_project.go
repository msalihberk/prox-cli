package commands

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	maxProjectFileBytes    = 12 << 10
	maxProjectCommandBytes = 4 << 10
)

var projectSecretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?im)^(\s*(?:api[_-]?key|access[_-]?token|password|secret|private[_-]?key)\s*[:=]\s*).+$`),
	regexp.MustCompile(`(?i)\bAIza[0-9A-Za-z_-]{20,}\b`),
	regexp.MustCompile(`(?i)\b(?:ghp|github_pat)_[A-Za-z0-9_]{20,}\b`),
	regexp.MustCompile(`(?s)-----BEGIN (?:RSA |EC |OPENSSH )?PRIVATE KEY-----.*?-----END (?:RSA |EC |OPENSSH )?PRIVATE KEY-----`),
}

type boundedBuffer struct {
	bytes.Buffer
	limit     int
	truncated bool
}

func (b *boundedBuffer) Write(data []byte) (int, error) {
	originalLen := len(data)
	remaining := b.limit - b.Len()
	if remaining > 0 {
		if len(data) > remaining {
			data = data[:remaining]
			b.truncated = true
		}
		_, _ = b.Buffer.Write(data)
	} else if len(data) > 0 {
		b.truncated = true
	}
	return originalLen, nil
}

func collectProjectContext() (string, error) {
	rootBytes, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", errors.New("project context requires a Git repository")
	}
	root := strings.TrimSpace(string(rootBytes))

	status, err := runGitSummary(root, "status", "--short", "--branch")
	if err != nil {
		return "", err
	}
	diff, err := runGitSummary(root, "diff", "--stat")
	if err != nil {
		return "", err
	}

	var context strings.Builder
	context.WriteString("Repository summary (automatically redacted; review before sharing):\n")
	context.WriteString("Git status:\n")
	context.WriteString(status)
	if status == "" {
		context.WriteString("(empty)\n")
	}
	context.WriteString("\nGit diff summary:\n")
	context.WriteString(diff)
	if diff == "" {
		context.WriteString("(no unstaged diff)\n")
	}
	for _, fileName := range []string{"README.md", "go.mod"} {
		content, truncated, err := readProjectFile(filepath.Join(root, fileName))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return "", fmt.Errorf("read project summary file %s: %w", fileName, err)
		}
		context.WriteString("\n")
		context.WriteString(fileName)
		if truncated {
			context.WriteString(" (truncated)")
		}
		context.WriteString(":\n")
		context.WriteString(content)
	}
	return redactProjectContext(context.String()), nil
}

func runGitSummary(root string, args ...string) (string, error) {
	commandArgs := append([]string{"-C", root}, args...)
	command := exec.Command("git", commandArgs...)
	output := &boundedBuffer{limit: maxProjectCommandBytes}
	command.Stdout = output
	if err := command.Run(); err != nil {
		return "", fmt.Errorf("collect project summary with git %s: %w", args[0], err)
	}
	result := output.String()
	if output.truncated {
		result += "\n[Git output truncated]\n"
	}
	return result, nil
}

func readProjectFile(path string) (string, bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", false, err
	}
	defer file.Close()

	output := &boundedBuffer{limit: maxProjectFileBytes}
	if _, err := io.Copy(output, file); err != nil {
		return "", false, err
	}
	return output.String(), output.truncated, nil
}

func redactProjectContext(context string) string {
	for _, pattern := range projectSecretPatterns {
		context = pattern.ReplaceAllString(context, "${1}[REDACTED]")
	}
	return context
}

func projectExplainInput(userPrompt, projectContext string) string {
	return "User request:\n" + userPrompt + "\n\nUser-reviewed project context:\n" + projectContext
}

func reviewProjectContext(output io.Writer, projectContext string, confirm func() bool) (bool, error) {
	if _, err := fmt.Fprintln(output, "Project context preview (only Git status/diff summary, README.md, and go.mod; likely secrets redacted). Review all content before sharing:"); err != nil {
		return false, err
	}
	if _, err := fmt.Fprintln(output, projectContext); err != nil {
		return false, err
	}
	return confirm(), nil
}
