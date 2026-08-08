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
	"os"
	"path/filepath"
)

type Config struct {
	AutoApprove bool `json:"auto_approve"`
}

func GetConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".prox_config.json")
}

func LoadConfig() Config {
	path := GetConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{AutoApprove: false}
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{AutoApprove: false}
	}
	return cfg
}

func SaveConfig(cfg Config) error {
	path := GetConfigPath()
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
