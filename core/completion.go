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
	script, err := GenerateCompletionScript(shell)
	if err != nil {
		return err
	}

	fmt.Println(script)
	return nil
}

func (c CompletionCommand) Description() string {
	return "Generate context-aware tab-completion scripts for bash, zsh, or powershell"
}

func (c CompletionCommand) SubCommands() []string {
	return []string{"bash", "zsh", "powershell"}
}

func (c CompletionCommand) Help() string {
	return "Usage: prox completion <bash|zsh|powershell>\n  Generate tab-completion scripts for the specified shell."
}

func GenerateCompletionScript(shell string) (string, error) {
	var rootSubCmds []string
	for name := range CommandRegistry {
		rootSubCmds = append(rootSubCmds, name)
	}
	sort.Strings(rootSubCmds)
	spaceSeparatedRoot := strings.Join(rootSubCmds, " ")

	switch shell {
	case "bash":
		var bashSwitchCases strings.Builder
		for name, cmd := range CommandRegistry {
			subs := cmd.SubCommands()
			if len(subs) > 0 {
				bashSwitchCases.WriteString(fmt.Sprintf(
					"        %s)\n            opts=\"%s\"\n            ;;\n",
					name, strings.Join(subs, " "),
				))
			}
		}

		return fmt.Sprintf(`_prox_completion() {
    local cur prev opts cmd
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    opts="%s"

    cmd=""
    local i=1
    while [[ $i -lt $COMP_CWORD ]]; do
        local s="${COMP_WORDS[i]}"
        if [[ "$s" != -* ]]; then
            cmd="${s##*/}"
            break
        fi
        i=$((i+1))
    done

    if [[ -z "$cmd" ]] || [[ "$cmd" == "$cur" ]]; then
        COMPREPLY=( $(compgen -W "${opts}" -- "${cur}") )
        return 0
    fi

    case "${cmd}" in
%s
        *)
            opts=""
            ;;
    esac

    COMPREPLY=( $(compgen -W "${opts}" -- "${cur}") )
    return 0
}
complete -F _prox_completion prox ._prox ./prox prox.exe`, spaceSeparatedRoot, bashSwitchCases.String()), nil

	case "zsh":
		var zshSwitchCases strings.Builder
		for name, cmd := range CommandRegistry {
			subs := cmd.SubCommands()
			if len(subs) > 0 {
				zshSwitchCases.WriteString(fmt.Sprintf(
					"        %s)\n            local -a subcmds\n            subcmds=(\n%s            )\n            _describe -t subcmds \"prox %s commands\" subcmds\n            ;;\n",
					name, formatZshSubCommands(subs), name,
				))
			}
		}

		return fmt.Sprintf(`#compdef prox ./prox prox.exe
_prox_completion() {
    local -a main_cmds
    
    if (( CURRENT == 2 )); then
        main_cmds=(
%s
        )
        _describe -t main_cmds 'prox commands' main_cmds
        return 0
    fi

    if (( CURRENT > 2 )); then
        local cmd_word="${words[2]##*/}"
        case $cmd_word in
%s
            *)
                _files
                ;;
        esac
    fi
}
_prox_completion "$@"`, formatZshMainCommands(rootSubCmds), zshSwitchCases.String()), nil

	case "powershell", "pwsh":
		var psCases strings.Builder
		for name, cmd := range CommandRegistry {
			subs := cmd.SubCommands()
			if len(subs) > 0 {
				var quotedSubs []string
				for _, s := range subs {
					quotedSubs = append(quotedSubs, fmt.Sprintf("'%s'", s))
				}
				psCases.WriteString(fmt.Sprintf("        '%s' { $sub = @(%s) }\n", name, strings.Join(quotedSubs, ", ")))
			}
		}

		var rootQuoted []string
		for _, r := range rootSubCmds {
			rootQuoted = append(rootQuoted, fmt.Sprintf("'%s'", r))
		}

		return fmt.Sprintf(`$proxCompleter = {
    param($wordToComplete, $commandAst, $cursorPosition)
    $elements = $commandAst.CommandElements | Where-Object { $_.Value -ne $null }
    $rawText = $commandAst.Extent.Text
    $isTrailingSpace = $rawText.EndsWith(" ")

    if ($elements.Count -le 0) {
        return
    }

    $mainCmd = $elements[1].Value
    if ($mainCmd -like "*\*") { $mainCmd = Split-Path $mainCmd -Leaf }
    if ($mainCmd -like "*/*") { $mainCmd = $mainCmd.Split('/')[-1] }

    if ($elements.Count -eq 1 -and -not $isTrailingSpace) {
        $choices = @(%s)
        $choices | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
            [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
        }
        return
    }

    $sub = @()
    switch ($mainCmd) {
%s
    }

    $sub | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
        [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
    }
}
Register-ArgumentCompleter -Native -CommandName 'prox' -ScriptBlock $proxCompleter
Register-ArgumentCompleter -Native -CommandName 'prox.exe' -ScriptBlock $proxCompleter
Register-ArgumentCompleter -Native -CommandName './prox' -ScriptBlock $proxCompleter
Register-ArgumentCompleter -Native -CommandName './prox.exe' -ScriptBlock $proxCompleter`, strings.Join(rootQuoted, ", "), psCases.String()), nil

	default:
		return "", fmt.Errorf("unsupported shell '%s'", shell)
	}
}

func formatZshMainCommands(cmds []string) string {
	var sb strings.Builder
	for _, cmd := range cmds {
		desc := CommandRegistry[cmd].Description()
		desc = strings.ReplaceAll(desc, "'", "")
		sb.WriteString(fmt.Sprintf("        '%s:%s'\n", cmd, desc))
	}
	return strings.TrimSuffix(sb.String(), "\n")
}

func formatZshSubCommands(subs []string) string {
	var sb strings.Builder
	for _, sub := range subs {
		sb.WriteString(fmt.Sprintf("                '%s'\n", sub))
	}
	return strings.TrimSuffix(sb.String(), "\n")
}
