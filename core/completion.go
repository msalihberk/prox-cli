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
	"fmt"
	"sort"
	"strings"
)

type CompletionCommand struct{}

func (c CompletionCommand) Execute(args []string) error {
	if len(args) < 1 {
		return errors.New("usage: prox completion [bash|zsh|powershell]")
	}

	shell := strings.ToLower(args[0])

	var cmdList []string
	for cmdName := range CommandRegistry {
		cmdList = append(cmdList, cmdName)
	}
	sort.Strings(cmdList)
	spaceSeparatedCmds := strings.Join(cmdList, " ")

	switch shell {
	case "bash":
		script := fmt.Sprintf(`_prox_completion() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    opts="%s"

    if [[ ${COMP_CWORD} -eq 1 ]] ; then
        COMPREPLY=( $(compgen -W "${opts}" -- "${cur}") )
        return 0
    fi
	}
	complete -F _prox_completion prox`, spaceSeparatedCmds)
		fmt.Println(script)

	case "zsh":
		script := fmt.Sprintf(`#compdef prox
	_prox_completion() {
    local -a commands
    commands=(
	%s
    )
    if (( CURRENT == 2 )); then
        _describe -t commands 'prox commands' commands
    fi
	}
	_prox_completion "$@"`, formatZshCommands(cmdList))
		fmt.Println(script)

	case "powershell", "pwsh":
		// PowerShell için dinamik dize dizisi formatı: 'ai', 'b64', 'hash'
		var psCmds []string
		for _, cmd := range cmdList {
			psCmds = append(psCmds, fmt.Sprintf("'%s'", cmd))
		}
		commaSeparatedCmds := strings.Join(psCmds, ", ")

		script := fmt.Sprintf(`$proxCompleter = {
    param($wordToComplete, $commandAst, $cursorPosition)
    $commands = @(%s)
    $commands | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
        [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
    }
	}
	Register-ArgumentCompleter -Native -CommandName 'prox' -ScriptBlock $proxCompleter
	Register-ArgumentCompleter -Native -CommandName 'prox.exe' -ScriptBlock $proxCompleter`, commaSeparatedCmds)
		fmt.Println(script)

	default:
		return fmt.Errorf("unsupported shell '%s'. Supported shells: bash, zsh, powershell, pwsh", shell)
	}
	return nil
}

func (c CompletionCommand) Description() string {
	return "Generate tab-completion scripts for bash, zsh, or PowerShell"
}
func (c CompletionCommand) Help() string {
	return ""
}

func formatZshCommands(cmds []string) string {
	var sb strings.Builder
	for _, cmd := range cmds {
		desc := CommandRegistry[cmd].Description()
		desc = strings.ReplaceAll(desc, "'", "")
		sb.WriteString(fmt.Sprintf("        '%s:%s'\n", cmd, desc))
	}
	return strings.TrimSuffix(sb.String(), "\n")
}
func (c CompletionCommand) SubCommands() []string {
	return []string{"bash", "zsh", "powershell", "pwsh"}
}
func init() {
	Register("completion", CompletionCommand{})
}
