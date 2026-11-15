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
	"strings"
	"time"
	"unicode/utf8"
)

type HistoryMessage struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

type ChatSession struct {
	ID        string           `json:"id"`
	Name      string           `json:"name,omitempty"`
	Messages  []HistoryMessage `json:"messages"`
	Timestamp time.Time        `json:"timestamp"`
}

type HistoryStore struct {
	Path string
}

func NewHistoryStore() (HistoryStore, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return HistoryStore{}, fmt.Errorf("locate user home directory: %w", err)
	}
	return HistoryStore{Path: filepath.Join(home, ".prox_agent_history.json")}, nil
}

func (s HistoryStore) Load() ([]ChatSession, error) {
	data, err := os.ReadFile(s.Path)
	if errors.Is(err, os.ErrNotExist) {
		return []ChatSession{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read agent history: %w", err)
	}

	var sessions []ChatSession
	if err := json.Unmarshal(data, &sessions); err != nil {
		return nil, fmt.Errorf("decode agent history: %w", err)
	}
	for i := range sessions {
		if sessions[i].Name == "" {
			sessions[i].Name = defaultSessionName(sessions[i].Messages, sessions[i].Timestamp)
		}
	}
	return sessions, nil
}

func (s HistoryStore) Find(id string) (ChatSession, error) {
	sessions, err := s.Load()
	if err != nil {
		return ChatSession{}, err
	}
	for _, session := range sessions {
		if session.ID == id {
			return session, nil
		}
	}
	return ChatSession{}, fmt.Errorf("agent session %q not found", id)
}

func (s HistoryStore) Save(session ChatSession) error {
	if session.ID == "" {
		session.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	if session.Timestamp.IsZero() {
		session.Timestamp = time.Now()
	}
	if session.Name == "" {
		session.Name = defaultSessionName(session.Messages, session.Timestamp)
	}
	for i := range session.Messages {
		if session.Messages[i].Timestamp.IsZero() {
			session.Messages[i].Timestamp = session.Timestamp
		}
	}

	sessions, err := s.Load()
	if err != nil {
		return err
	}
	found := false
	for i := range sessions {
		if sessions[i].ID == session.ID {
			sessions[i] = session
			found = true
			break
		}
	}
	if !found {
		sessions = append(sessions, session)
	}
	return s.write(sessions)
}

func (s HistoryStore) Delete(id string) error {
	sessions, err := s.Load()
	if err != nil {
		return err
	}
	for i := range sessions {
		if sessions[i].ID == id {
			sessions = append(sessions[:i], sessions[i+1:]...)
			return s.write(sessions)
		}
	}
	return fmt.Errorf("agent session %q not found", id)
}

func (s HistoryStore) Clear() error {
	return s.write([]ChatSession{})
}

func (s HistoryStore) write(sessions []ChatSession) error {
	data, err := json.MarshalIndent(sessions, "", "  ")
	if err != nil {
		return fmt.Errorf("encode agent history: %w", err)
	}
	if err := atomicWriteFile(s.Path, data, 0600); err != nil {
		return fmt.Errorf("write agent history: %w", err)
	}
	return nil
}

func defaultSessionName(messages []HistoryMessage, timestamp time.Time) string {
	for _, message := range messages {
		if message.Role != "user" {
			continue
		}
		name := strings.Join(strings.Fields(message.Content), " ")
		if name == "" {
			continue
		}
		if utf8.RuneCountInString(name) > 60 {
			name = string([]rune(name)[:57]) + "..."
		}
		return name
	}
	if timestamp.IsZero() {
		timestamp = time.Now()
	}
	return "Agent session " + timestamp.Format("2006-01-02 15:04")
}

func SaveHistory(messages []HistoryMessage) error {
	store, err := NewHistoryStore()
	if err != nil {
		return err
	}
	return store.Save(ChatSession{Messages: messages})
}
