//go:build windows

package core

import (
	"errors"
	"fmt"

	"golang.org/x/sys/windows/registry"
)

const userEnvironmentKey = `Environment`

func setUserAPIKey(apiKey string) error {
	key, _, err := registry.CreateKey(registry.CURRENT_USER, userEnvironmentKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()
	if err := key.SetStringValue("PROX_API_KEY", apiKey); err != nil {
		return fmt.Errorf("write user environment variable: %w", err)
	}
	return nil
}

func removeUserAPIKey() error {
	key, err := registry.OpenKey(registry.CURRENT_USER, userEnvironmentKey, registry.SET_VALUE)
	if errors.Is(err, registry.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	defer key.Close()
	if err := key.DeleteValue("PROX_API_KEY"); err != nil && !errors.Is(err, registry.ErrNotExist) {
		return fmt.Errorf("delete user environment variable: %w", err)
	}
	return nil
}
