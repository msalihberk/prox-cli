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

type HelpCommand struct{}

func RenderWelcomeScreen() string {
	var builder strings.Builder
	builder.WriteString("\n")
	builder.WriteString(styleText("PROX CLI", colorCyan+colorBold))
	builder.WriteString("\n")
	builder.WriteString(styleText("Modern command toolkit for developers and security workflows", colorWhite+colorBold))
	builder.WriteString("\n\n")
	builder.WriteString("Quick start:\n")
	builder.WriteString("  • ")
	builder.WriteString(styleText("prox help", colorGreen+colorBold))
	builder.WriteString("     shows the complete command catalog\n")
	builder.WriteString("  • ")
	builder.WriteString(styleText("prox version", colorYellow+colorBold))
	builder.WriteString("  prints the current build metadata\n")
	builder.WriteString("  • ")
	builder.WriteString(styleText("prox setup", colorMagenta+colorBold))
	builder.WriteString("   configures the environment for first-time use\n\n")
	builder.WriteString("Tip: use ")
	builder.WriteString(styleText("prox help", colorGreen+colorBold))
	builder.WriteString(" for a full overview of available utilities.\n")
	return builder.String()
}

func GetColorizedBanner() string {
	banner := `
       ______________      
      /              \
     /                \
    /    __            \        ██████╗ ██████╗  ██████╗ ██╗  ██╗      ██████╗██╗     ██╗
   |    \  \            |       ██╔══██╗██╔══██╗██╔═══██╗╚██╗██╔╝     ██╔════╝██║     ██║
   |     \  \    ___    |       ██████╔╝██████╔╝██║   ██║ ╚███╔╝      ██║     ██║     ██║
   |     /  /   |___|   |       ██╔═══╝ ██╔══██╗██║   ██║ ██╔██╗      ██║     ██║     ██║ 
   |    /__/            |       ██║     ██║  ██║╚██████╔╝██╔╝ ██╗     ╚██████╗███████╗██║
    \                  /        ╚═╝     ╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═╝      ╚═════╝╚══════╝╚═╝ 
     \                /         
      \______________/          
                                
══════════════════════════════════════════════⬢══════════════════════════════════════════════

                        The Swiss Army Knife for Developers & Sysadmins

══════════════════════════════════════════════⬢══════════════════════════════════════════════`
	return colorizeBanner(banner)
}

func RenderHelpScreen() string {
	var commandNames []string
	for name := range CommandRegistry {
		commandNames = append(commandNames, name)
	}
	sort.Strings(commandNames)

	var builder strings.Builder
	builder.WriteString("\n")
	builder.WriteString(GetColorizedBanner())
	builder.WriteString("\n\n")

	for _, name := range commandNames {
		cmd := CommandRegistry[name]
		builder.WriteString("  ")
		builder.WriteString(styleText(fmt.Sprintf("%-12s", name), colorYellow+colorBold))
		builder.WriteString("  ")
		builder.WriteString(strings.TrimSpace(cmd.Description()))
		builder.WriteString("\n")
	}

	builder.WriteString("\n")
	builder.WriteString(styleText("Quick examples:", colorWhite+colorBold))
	builder.WriteString("\n")
	builder.WriteString("  • ")
	builder.WriteString(styleText("prox help", colorGreen+colorBold))
	builder.WriteString("\n")
	builder.WriteString("  • ")
	builder.WriteString(styleText("prox version", colorGreen+colorBold))
	builder.WriteString("\n")
	builder.WriteString("  • ")
	builder.WriteString(styleText("prox setup", colorGreen+colorBold))
	builder.WriteString("\n")
	return builder.String()
}

func colorizeBanner(banner string) string {
	lines := strings.Split(banner, "\n")
	var colorized []string
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			colorized = append(colorized, line)
			continue
		}
		// If divider line (contains ═══)
		if strings.Contains(line, "═══") {
			colorized = append(colorized, styleText(line, ColorLightCyan))
			continue
		}
		// If subtitle
		if strings.Contains(line, "Swiss Army Knife") {
			colorized = append(colorized, styleText(line, ColorLightMagenta+colorBold))
			continue
		}
		// Shield and text parts. Split at column index 32
		runes := []rune(line)
		if len(runes) > 32 {
			left := string(runes[:32])
			right := string(runes[32:])
			colorized = append(colorized, styleText(left, ColorLightCyan)+styleText(right, ColorLightMagenta+colorBold))
		} else {
			colorized = append(colorized, styleText(line, ColorLightCyan))
		}
	}
	return strings.Join(colorized, "\n")
}

func (h HelpCommand) Execute(args []string) error {
	if len(args) > 0 {
		return errors.New("Help command does not accept any arguments")
	}
	PrintMessage("%s", RenderHelpScreen())
	return nil
}
func (h HelpCommand) Description() string {
	return "Display help information for available commands (CORE)"
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

