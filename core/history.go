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
	"time"
)

type HistoryMessage struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

type ChatSession struct {
	ID        string           `json:"id"`
	Messages  []HistoryMessage `json:"messages"`
	Timestamp time.Time        `json:"timestamp"`
}

func SaveHistory(messages []HistoryMessage) error {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".prox_agent_history.json")

	var sessions []ChatSession
	data, err := os.ReadFile(path)
	if err == nil {
		json.Unmarshal(data, &sessions)
	}

	newSession := ChatSession{
		ID:        time.Now().Format("20060102-150405"),
		Messages:  messages,
		Timestamp: time.Now(),
	}
	sessions = append(sessions, newSession)

	newData, err := json.MarshalIndent(sessions, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, newData, 0644)
}
