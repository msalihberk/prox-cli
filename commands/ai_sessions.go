package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"unicode/utf8"

	"prox-cli/core"
)

type AISessionsCommand struct{}

func (c AISessionsCommand) Execute(args []string) error {
	parser := core.New(args, false)
	if err := parser.Parse(); err != nil {
		return err
	}
	return agentSessionsCommand(parser.AllPos())
}

func (c AISessionsCommand) ExecuteAgentTool(_ context.Context, args []string, approved bool) (string, error) {
	args = append([]string(nil), args...)
	if len(args) == 0 {
		args = []string{"list"}
	}
	if sessionActionRequiresApproval(args) && !approved {
		return "", errors.New("this session action requires user approval")
	}
	store, err := core.NewHistoryStore()
	if err != nil {
		return "", err
	}
	var output bytes.Buffer
	if err := runAgentSessionAction(args, store, &output, nil, func(string) bool { return approved }); err != nil {
		return "", err
	}
	return strings.TrimSpace(output.String()), nil
}

func sessionActionRequiresApproval(args []string) bool {
	if len(args) == 0 {
		return false
	}
	switch args[0] {
	case "rename", "export", "delete", "clear":
		return true
	default:
		return false
	}
}

func (c AISessionsCommand) Description() string {
	return "List, resume, rename, export, and delete saved AI agent sessions"
}

func (c AISessionsCommand) Help() string {
	return "Usage: prox ai sessions [list|resume <id>|rename <id> <name>|export <id> <file>|delete <id>|clear]"
}

func (c AISessionsCommand) SubCommands() []string {
	return []string{"list", "resume", "rename", "export", "delete", "clear", "help"}
}

func agentSessionsCommand(args []string) error {
	store, err := core.NewHistoryStore()
	if err != nil {
		return err
	}
	return runAgentSessionAction(args, store, os.Stdout, startInteractiveAgentWithSession, func(sessionID string) bool {
		if sessionID == "all" {
			return core.PromptConfirm("Permanently delete all saved agent sessions?", false)
		}
		return core.PromptConfirm("Permanently delete agent session "+sessionID+"?", false)
	})
}

func runAgentSessionAction(args []string, store core.HistoryStore, output io.Writer, resume func(core.ChatSession) error, confirmDelete func(string) bool) error {
	if len(args) == 0 || args[0] == "list" {
		return listAgentSessions(store, output)
	}

	switch args[0] {
	case "resume":
		if len(args) != 2 {
			return errors.New("usage: prox ai sessions resume <id>")
		}
		if resume == nil {
			return errors.New("agent session resume handler is unavailable")
		}
		session, err := store.Find(args[1])
		if err != nil {
			return err
		}
		return resume(session)
	case "delete":
		if len(args) != 2 {
			return errors.New("usage: prox ai sessions delete <id>")
		}
		if _, err := store.Find(args[1]); err != nil {
			return err
		}
		if confirmDelete == nil || !confirmDelete(args[1]) {
			_, err := fmt.Fprintln(output, "Deletion cancelled")
			return err
		}
		if err := store.Delete(args[1]); err != nil {
			return err
		}
		_, err := fmt.Fprintf(output, "Deleted agent session %s\n", args[1])
		return err
	case "clear":
		if len(args) != 1 {
			return errors.New("usage: prox ai sessions clear")
		}
		if confirmDelete == nil || !confirmDelete("all") {
			_, err := fmt.Fprintln(output, "Deletion cancelled")
			return err
		}
		if err := store.Clear(); err != nil {
			return err
		}
		_, err := fmt.Fprintln(output, "Deleted all agent sessions")
		return err
	case "rename":
		if len(args) < 3 {
			return errors.New("usage: prox ai sessions rename <id> <name>")
		}
		name := strings.TrimSpace(strings.Join(args[2:], " "))
		if name == "" || utf8.RuneCountInString(name) > 100 {
			return errors.New("session name must contain 1 to 100 characters")
		}
		session, err := store.Find(args[1])
		if err != nil {
			return err
		}
		session.Name = name
		if err := store.Save(session); err != nil {
			return err
		}
		_, err = fmt.Fprintf(output, "Renamed agent session %s\n", session.ID)
		return err
	case "export":
		if len(args) < 3 {
			return errors.New("usage: prox ai sessions export <id> <file>")
		}
		session, err := store.Find(args[1])
		if err != nil {
			return err
		}
		data, err := json.MarshalIndent(session, "", "  ")
		if err != nil {
			return fmt.Errorf("encode agent session: %w", err)
		}
		path := strings.Join(args[2:], " ")
		if err := os.WriteFile(path, data, 0600); err != nil {
			return fmt.Errorf("export agent session: %w", err)
		}
		if err := os.Chmod(path, 0600); err != nil {
			return fmt.Errorf("secure exported agent session: %w", err)
		}
		_, err = fmt.Fprintf(output, "Exported agent session to %s\n", path)
		return err
	case "help":
		_, err := fmt.Fprintln(output, (AISessionsCommand{}).Help())
		return err
	default:
		return fmt.Errorf("unknown session action %q; run 'prox ai sessions help' for usage", args[0])
	}
}

func listAgentSessions(store core.HistoryStore, output io.Writer) error {
	sessions, err := store.Load()
	if err != nil {
		return err
	}
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].Timestamp.After(sessions[j].Timestamp)
	})
	if _, err := fmt.Fprintln(output, "ID\tUPDATED\tMESSAGES\tNAME"); err != nil {
		return err
	}
	for _, session := range sessions {
		if _, err := fmt.Fprintf(output, "%s\t%s\t%d\t%s\n",
			session.ID,
			session.Timestamp.Format("2006-01-02 15:04"),
			len(session.Messages),
			session.Name,
		); err != nil {
			return err
		}
	}
	return nil
}

func startInteractiveAgentWithSession(session core.ChatSession) error {
	return core.RunAgentTUIWithSessionAndTools(requestAgentResponse, executeShellCommand, assessCommandRisk, session, proxAgentTools())
}

func init() {
	core.Register("ai-sessions", AISessionsCommand{})
}
