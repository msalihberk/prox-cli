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
	PrintInfo("Prox CLI - Version 0.7.0")
	PrintInfo("Copyright 2026 Mustafa Salih Berk")
	PrintInfo("GitHub: https://github.com/msalihberk/prox-cli")
	PrintInfo("Licensed under the Apache License 2.0.")
	PrintNewLine()
	return nil
}
func (v VersionCommand) Description() string {
	return "Display the current version of the CLI tool"
}
func init() {
	register("version", VersionCommand{})
}
