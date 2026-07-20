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
		PrintMessage(" \033[33m%-10s\033[0m : %s", name, cmd.Description())
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
	return "Display help information for available commands \033[32m(CORE)\033[0m"
}
func (v HelpCommand) Help() string {
	help := "Usage: prox help"
	return help
}
func (v HelpCommand) SubCommands() []string {
	return []string{""}
}
func init() {
	Register("help", HelpCommand{})
}
