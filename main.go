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

package main

import (
	"bufio"
	"os"
	"prox-cli/commands"
	"strings"
)

func main() {
	controlArguments()
}
func readPipeInputs() ([]string, error) {
	var lines []string

	stat, err := os.Stdin.Stat()

	if err != nil {
		return nil, err
	}

	if (stat.Mode() & os.ModeCharDevice) == 0 {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			return nil, err
		}
	} else {
		return nil, nil
	}
	return lines, nil
}
func controlArguments() {
	if len(os.Args) < 2 {
		commands.PrintMessage("Usage: prox [command] <arguments>")
		commands.PrintMessage("Run 'prox help' for a list of available commands.")
		os.Exit(1)
	}

	subCommand := strings.TrimLeft(os.Args[1], "-")
	subArgs := os.Args[2:]
	input, errp := readPipeInputs()
	if errp == nil && input != nil {
		subArgs = append(subArgs, input...)
	}

	if subCommand == "v" {
		subCommand = "version"
	}
	if subCommand == "h" {
		subCommand = "help"
	}

	cmd, exists := commands.CommandRegistry[subCommand]
	if !exists {
		commands.PrintError("Unknown command: %s", os.Args[1])

		if suggestion := getClosestCommand(subCommand); suggestion != "" {
			commands.PrintInfo("Did you mean: '%s'?\n\n", suggestion)
		}

		commands.PrintMessage("Run 'prox help' to see all available commands.")
		os.Exit(1)
	}

	err := cmd.Execute(subArgs)
	if err != nil {
		commands.PrintError("%v", err)
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
