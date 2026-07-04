package commands

import (
	"errors"
	"fmt"
	"sort"
)

type HelpCommand struct{}

func (h HelpCommand) Execute(args []string) error {
	if len(args) > 0 {
		return errors.New("Help command does not accept any arguments")
	}
	fmt.Printf("\n")
	fmt.Println("Prox CLI - Available Commands:")
	fmt.Println("--------------------------------")

	var commandNames []string
	for name := range CommandRegistry {
		commandNames = append(commandNames, name)
	}
	sort.Strings(commandNames)

	for _, name := range commandNames {
		cmd := CommandRegistry[name]
		fmt.Printf("  %-10s : %s\n", name, cmd.Description())
	}
	fmt.Printf("\n")
	cmd, exists := CommandRegistry["version"]
	if exists {
		cmd.Execute(nil)
	}
	fmt.Printf("\n")

	return nil
}
func (h HelpCommand) Description() string {
	return "Display help information for available commands"
}
