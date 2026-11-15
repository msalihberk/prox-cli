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
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	DefaultAIModel                 = "gemini-flash-lite-latest"
	DefaultAIRequestTimeoutSeconds = 30
)

type Config struct {
	AutoApprove             bool   `json:"auto_approve"`
	AIModel                 string `json:"ai_model,omitempty"`
	AIRequestTimeoutSeconds int    `json:"ai_request_timeout_seconds,omitempty"`
}

func GetConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".prox_config.json")
}

func DefaultConfig() Config {
	return Config{
		AIModel:                 DefaultAIModel,
		AIRequestTimeoutSeconds: DefaultAIRequestTimeoutSeconds,
	}
}

func LoadConfig() (Config, error) {
	path := GetConfigPath()
	if path == "" {
		return Config{}, errors.New("locate user home directory")
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return DefaultConfig(), nil
	}
	if err != nil {
		return Config{}, fmt.Errorf("read configuration: %w", err)
	}

	cfg, err := decodeConfig(data)
	if err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func decodeConfig(data []byte) (Config, error) {
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("decode configuration: %w", err)
	}
	if cfg.AIModel == "" {
		cfg.AIModel = DefaultAIModel
	}
	if cfg.AIRequestTimeoutSeconds == 0 {
		cfg.AIRequestTimeoutSeconds = DefaultAIRequestTimeoutSeconds
	}
	return cfg, nil
}

func SaveConfig(cfg Config) error {
	path := GetConfigPath()
	if path == "" {
		return errors.New("locate user home directory")
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := atomicWriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write configuration: %w", err)
	}
	return nil
}
