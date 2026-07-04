package main

import (
	"fmt"
	"os"
	"prox-cli/commands"
)

func main() {
	controlArguments()
}
func controlArguments() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: prox [command] <arguments>")
		os.Exit(1)
	}

	subCommand := os.Args[1]
	subArgs := os.Args[2:]

	cmd, exists := commands.CommandRegistry[subCommand]
	if !exists {
		fmt.Printf("Unknown command: %s\n", subCommand)
		os.Exit(1)
	}

	err := cmd.Execute(subArgs)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
