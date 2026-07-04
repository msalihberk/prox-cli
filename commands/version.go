package commands

import (
	"errors"
	"fmt"
)

type VersionCommand struct{}

func (v VersionCommand) Execute(args []string) error {
	if len(args) > 0 {
		return errors.New("Version command does not accept any arguments")
	}
	fmt.Printf("\n")
	fmt.Println("Prox CLI - Version 0.1.0")
	fmt.Println("Developed by Mustafa Salih Berk")
	fmt.Println("GitHub: https://github.com/msalihberk/prox-cli")
	fmt.Printf("\n")
	return nil
}
func (v VersionCommand) Description() string {
	return "Display the current version of the CLI tool"
}
