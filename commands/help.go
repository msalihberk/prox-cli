package commands

import (
	"errors"
	"sort"
)

type HelpCommand struct{}

func (h HelpCommand) Execute(args []string) error {
	if len(args) > 0 {
		return errors.New("Help command does not accept any arguments")
	}
	PrintNewLine()
	PrintInfo("Prox CLI - Available Commands:")
	PrintInfo("--------------------------------")

	var commandNames []string
	for name := range CommandRegistry {
		commandNames = append(commandNames, name)
	}
	sort.Strings(commandNames)

	for _, name := range commandNames {
		cmd := CommandRegistry[name]
		PrintMessage("  %-10s : %s", name, cmd.Description())
	}
	PrintNewLine()
	cmd, exists := CommandRegistry["version"]
	if exists {
		cmd.Execute(nil)
	}
	PrintNewLine()

	return nil
}
func (h HelpCommand) Description() string {
	return "Display help information for available commands"
}
