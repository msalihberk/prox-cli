package commands

import (
	"errors"
)

type VersionCommand struct{}

func (v VersionCommand) Execute(args []string) error {
	if len(args) > 0 {
		return errors.New("Version command does not accept any arguments")
	}
	PrintNewLine()
	PrintInfo("Prox CLI - Version 0.1.0")
	PrintInfo("Developed by Mustafa Salih Berk")
	PrintInfo("GitHub: https://github.com/msalihberk/prox-cli")
	PrintNewLine()
	return nil
}
func (v VersionCommand) Description() string {
	return "Display the current version of the CLI tool"
}
