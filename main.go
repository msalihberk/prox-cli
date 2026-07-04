package main

import (
	"fmt"
	"os"
	"prox-cli/commands"
	"strings"
)

func main() {
	controlArguments()
}
func controlArguments() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: prox [command] <arguments>")
		fmt.Println("Run 'prox help' for a list of available commands.")
		os.Exit(1)
	}

	subCommand := strings.TrimLeft(os.Args[1], "-")
	subArgs := os.Args[2:]

	if subCommand == "v" {
		subCommand = "version"
	}
	if subCommand == "h" {
		subCommand = "help"
	}

	cmd, exists := commands.CommandRegistry[subCommand]
	if !exists {
		fmt.Printf("Unknown command: %s\n", os.Args[1])

		if suggestion := getClosestCommand(subCommand); suggestion != "" {
			fmt.Printf("Did you mean: '%s'?\n\n", suggestion)
		}

		fmt.Println("Run 'prox help' to see all available commands.")
		os.Exit(1)
	}

	err := cmd.Execute(subArgs)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
func getClosestCommand(input string) string {
	closest := ""
	minDistance := 3

	for registeredName := range commands.CommandRegistry {
		dist := distance(input, registeredName)
		if dist < minDistance {
			minDistance = dist
			closest = registeredName
		}
	}
	return closest
}

func distance(s, t string) int {
	d := make([][]int, len(s)+1)
	for i := range d {
		d[i] = make([]int, len(t)+1)
		d[i][0] = i
	}
	for j := range d[0] {
		d[0][j] = j
	}
	for i := 1; i <= len(s); i++ {
		for j := 1; j <= len(t); j++ {
			if s[i-1] == t[j-1] {
				d[i][j] = d[i-1][j-1]
			} else {
				min := d[i-1][j] + 1
				if d[i][j-1]+1 < min {
					min = d[i][j-1] + 1
				}
				if d[i-1][j-1]+1 < min {
					min = d[i-1][j-1] + 1
				}
				d[i][j] = min
			}
		}
	}
	return d[len(s)][len(t)]
}
