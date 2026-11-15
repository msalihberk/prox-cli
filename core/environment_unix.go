//go:build !windows

package core

import (
	"fmt"
	"os"
	"strings"
)

func setUserAPIKey(apiKey string) error {
	path, err := getUnixShellRcPath()
	if err != nil {
		return err
	}
	value := "export PROX_API_KEY=" + shellSingleQuote(apiKey)
	_, err = updateTargetFileBlock(path, "ENV_CONFIG", value)
	if err != nil {
		return fmt.Errorf("update %s: %w", path, err)
	}
	if err := os.Chmod(path, 0600); err != nil {
		return fmt.Errorf("restrict permissions on %s: %w", path, err)
	}
	return nil
}

func removeUserAPIKey() error {
	path, err := getUnixShellRcPath()
	if err != nil {
		return err
	}
	if _, err := removeTargetFileBlock(path, "ENV_CONFIG"); err != nil {
		return fmt.Errorf("update %s: %w", path, err)
	}
	return nil
}

func shellSingleQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}
