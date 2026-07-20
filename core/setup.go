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

type SetupCommand struct{}

func (c SetupCommand) Execute(args []string) error {
	if len(args) > 1 {
		PrintError("This command expects at most one argument. For more information, run the 'prox setup help' command.")
		return nil
	}
	if len(args) == 0 {
		setupAll()
		return nil
	}
	switch args[0] {
	case "help":
		PrintInfo("%s", c.Help())
		return nil
	case "install":
		if err := InstallBinary(); err != nil {
			return err
		}
		return nil
	case "setup-env":
		return SetupEnvironment()
	case "setup-completion":
		return SetupShellCompletion()
	default:
		return nil
	}
}
func setupAll() {
	SetupEnvironment()
	SetupShellCompletion()
	if err := InstallBinary(); err != nil {
		PrintError("%s", err.Error())
	}
}
func (c SetupCommand) Description() string {
	return "Set up the proxy system automatically or using specific manual commands \033[32m(CORE)\033[0m"
}

func (c SetupCommand) SubCommands() []string {
	return []string{"install", "setup-env", "setup-completion", "help"}
}

func (c SetupCommand) Help() string {
	return "Usage:\n  prox setup          - Automatically configure the proxy system" +
		"\n Usage:\n  prox setup install          - Install binary and register auto-completion" +
		"\n Usage:\n  prox setup setup-env        - Set up or update your PROX_API_KEY securely" +
		"\n Usage:\n  prox setup setup-completion - Automate context-aware tab completion adjustments"
}

func init() {
	Register("setup", SetupCommand{})
}
