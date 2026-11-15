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
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	configBlockStart = "## PROX-CLI CONFIG BEGIN"
	configBlockEnd   = "## PROX-CLI CONFIG END"
)

func InstallBinary() error {
	fmt.Println("⏳ Starting system installation...")

	currentExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current executable path: %v", err)
	}

	var targetDir string
	if runtime.GOOS == "windows" {
		targetDir = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local", "Microsoft", "WindowsApps")
	} else {
		targetDir = "/usr/local/bin"
	}

	binaryName := "prox"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	targetPath := filepath.Join(targetDir, binaryName)

	src, err := os.Open(currentExe)
	if err != nil {
		return fmt.Errorf("failed to open source binary: %v", err)
	}
	defer src.Close()

	dst, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		if runtime.GOOS != "windows" && os.IsPermission(err) {
			return fmt.Errorf("permission denied. Please run this command using sudo:\n-> sudo ./prox self install")
		}
		return fmt.Errorf("failed to create target binary at %s: %v", targetPath, err)
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return fmt.Errorf("failed to copy binary data: %v", err)
	}

	fmt.Printf("✓ Binary successfully installed to: %s\n", targetPath)
	return nil
}

func SetupEnvironment() error {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("🤖 prox-cli AI Configuration Wizard")
	fmt.Print("Enter your PROX_API_KEY (For skip this step enter 'skip'): ")

	apiKey, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read API key: %v", err)
	}
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "skip" {
		return nil
	}
	if apiKey == "" {
		return errors.New("API key cannot be empty")
	}
	if strings.ContainsAny(apiKey, "\r\n\x00") {
		return errors.New("API key must not contain newline or null characters")
	}

	if err := setUserAPIKey(apiKey); err != nil {
		return fmt.Errorf("save PROX_API_KEY: %w", err)
	}
	if runtime.GOOS == "windows" {
		fmt.Println("✓ PROX_API_KEY saved for your Windows user account. Restart terminals to apply it.")
		return nil
	}

	rcPath, err := getUnixShellRcPath()
	if err != nil {
		return err
	}
	fmt.Printf("✓ PROX_API_KEY saved in %s\n", rcPath)
	fmt.Printf("✨ Run 'source %s' to apply changes.\n", rcPath)
	return nil
}

func RemoveAPIKey() error {
	if err := removeUserAPIKey(); err != nil {
		return fmt.Errorf("remove saved PROX_API_KEY: %w", err)
	}
	if err := os.Unsetenv("PROX_API_KEY"); err != nil {
		return fmt.Errorf("clear current PROX_API_KEY: %w", err)
	}
	return nil
}

func SetupShellCompletion() error {
	var targetPath string
	var shellName string
	var err error

	if runtime.GOOS == "windows" {
		shellName = "powershell"
		targetPath, err = getWindowsPowerShellProfilePath()
		if err != nil {
			return fmt.Errorf("failed to resolve PowerShell profile path: %v", err)
		}
		dir := filepath.Dir(targetPath)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create profile directory: %v", dir)
		}
	} else {
		shellPath := os.Getenv("SHELL")
		if strings.Contains(shellPath, "zsh") {
			shellName = "zsh"
		} else {
			shellName = "bash"
		}
		targetPath, err = getUnixShellRcPath()
		if err != nil {
			return err
		}
	}

	fmt.Printf("⏳ Generating tab-completion script for %s...\n", shellName)
	script, err := GenerateCompletionScript(shellName)
	if err != nil {
		return fmt.Errorf("failed to generate completion script: %v", err)
	}

	updated, err := updateTargetFileBlock(targetPath, "COMPLETION_CONFIG", script)
	if err != nil {
		return err
	}

	if updated {
		fmt.Printf("✓ Tab-completion injected/updated successfully in: %s\n", targetPath)
		if runtime.GOOS == "windows" {
			fmt.Println("✨ Restart your PowerShell window to activate autocomplete!")
		} else {
			fmt.Printf("✨ Run 'source %s' to activate autocomplete!\n", targetPath)
		}
	} else {
		fmt.Printf("✓ Tab-completion is already up to date in %s. No updates required.\n", targetPath)
	}

	return nil
}

func updateTargetFileBlock(filePath, blockIdentifier, payload string) (bool, error) {
	content, err := os.ReadFile(filePath)
	if err != nil && !os.IsNotExist(err) {
		return false, err
	}

	startTag := fmt.Sprintf("%s_%s", configBlockStart, blockIdentifier)
	endTag := fmt.Sprintf("%s_%s", configBlockEnd, blockIdentifier)
	newBlockContent := fmt.Sprintf("%s\n%s\n%s", startTag, payload, endTag)

	fileStr := string(content)

	startIndex := strings.Index(fileStr, startTag)
	endTagIndex := strings.Index(fileStr, endTag)
	if startIndex >= 0 || endTagIndex >= 0 {
		if startIndex < 0 || endTagIndex < startIndex {
			return false, fmt.Errorf("incomplete or misordered Prox configuration block in %s", filePath)
		}
		endIndex := endTagIndex + len(endTag)
		existingBlock := fileStr[startIndex:endIndex]

		if existingBlock == newBlockContent {
			return false, nil
		}

		fileStr = fileStr[:startIndex] + newBlockContent + fileStr[endIndex:]
	} else {
		if len(fileStr) > 0 && !strings.HasSuffix(fileStr, "\n") {
			fileStr += "\n"
		}
		fileStr += newBlockContent + "\n"
	}

	mode := os.FileMode(0600)
	if info, err := os.Stat(filePath); err == nil {
		mode = info.Mode().Perm()
	} else if !errors.Is(err, os.ErrNotExist) {
		return false, err
	}
	err = atomicWriteFile(filePath, []byte(fileStr), mode)
	return true, err
}

func removeTargetFileBlock(filePath, blockIdentifier string) (bool, error) {
	content, err := os.ReadFile(filePath)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	startTag := fmt.Sprintf("%s_%s", configBlockStart, blockIdentifier)
	endTag := fmt.Sprintf("%s_%s", configBlockEnd, blockIdentifier)
	start := strings.Index(string(content), startTag)
	if start < 0 {
		if strings.Contains(string(content), endTag) {
			return false, fmt.Errorf("misordered Prox configuration block in %s", filePath)
		}
		return false, nil
	}
	end := strings.Index(string(content)[start+len(startTag):], endTag)
	if end < 0 {
		return false, fmt.Errorf("incomplete Prox configuration block in %s", filePath)
	}
	end += start + len(startTag) + len(endTag)
	updated := append([]byte{}, content[:start]...)
	updated = append(updated, content[end:]...)
	info, err := os.Stat(filePath)
	if err != nil {
		return false, err
	}
	if err := atomicWriteFile(filePath, updated, info.Mode().Perm()); err != nil {
		return false, err
	}
	return true, nil
}

func getUnixShellRcPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to locate home directory: %v", err)
	}

	shell := os.Getenv("SHELL")
	if strings.Contains(shell, "zsh") {
		return filepath.Join(homeDir, ".zshrc"), nil
	}
	return filepath.Join(homeDir, ".bashrc"), nil
}

func getWindowsPowerShellProfilePath() (string, error) {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1"), nil
}
