// Copyright 2026 Mustafa Salih Berk
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package commands

import (
	"fmt"
	"strings"
)

// ArgValue encapsulates a parsed argument's state
type ArgValue struct {
	Value string
	Found bool
}

// Parser implements a clean, extensible argument parser
// Supports: positional args, short flags (-v), long flags (--verbose), and flag values
type Parser struct {
	args  []string
	flags map[string]*ArgValue
	pos   []string
	debug bool
}

// New creates a parser instance
func New(args []string, debug bool) *Parser {
	return &Parser{
		args:  args,
		flags: make(map[string]*ArgValue),
		pos:   []string{},
		debug: debug,
	}
}

// Parse processes raw arguments into flags and positional values
func (p *Parser) Parse() error {
	for i := 0; i < len(p.args); i++ {
		arg := p.args[i]

		// Handle long flags (--flag or --flag=value)
		if strings.HasPrefix(arg, "--") {
			key, val, needsNext := p.parseLongFlag(arg)
			if needsNext && i+1 < len(p.args) && !isFlag(p.args[i+1]) {
				p.flags[key] = &ArgValue{Value: p.args[i+1], Found: true}
				i++
			} else {
				p.flags[key] = &ArgValue{Value: val, Found: true}
			}
			continue
		}

		// Handle short flags (-f or -fvalue)
		if strings.HasPrefix(arg, "-") && arg != "-" && !isNumeric(arg) {
			key, val, needsNext := p.parseShortFlag(arg)
			if needsNext && i+1 < len(p.args) && !isFlag(p.args[i+1]) {
				p.flags[key] = &ArgValue{Value: p.args[i+1], Found: true}
				i++
			} else {
				p.flags[key] = &ArgValue{Value: val, Found: true}
			}
			continue
		}

		// Positional argument
		p.pos = append(p.pos, arg)
	}

	p.logDebug()
	return nil
}

// parseLongFlag extracts key, value, and whether next arg is needed
func (p *Parser) parseLongFlag(arg string) (string, string, bool) {
	flag := strings.TrimPrefix(arg, "--")
	if idx := strings.IndexByte(flag, '='); idx != -1 {
		return flag[:idx], flag[idx+1:], false
	}
	return flag, "", true
}

// parseShortFlag handles combined flags like -abc or -av value
func (p *Parser) parseShortFlag(arg string) (string, string, bool) {
	flag := strings.TrimPrefix(arg, "-")
	if len(flag) > 1 && flag[0] != '-' {
		return string(flag[0]), flag[1:], len(flag[1:]) == 0
	}
	return flag, "", true
}

// Get retrieves a flag value by name (works with both short and long forms)
func (p *Parser) Get(flag string) *ArgValue {
	if val, ok := p.flags[flag]; ok {
		return val
	}
	return &ArgValue{Value: "", Found: false}
}

// GetAlias retrieves a flag value by checking multiple aliases (e.g. "v" and "verbose")
// Returns the first found match, allowing -v and --verbose to work interchangeably
func (p *Parser) GetAlias(aliases ...string) *ArgValue {
	for _, alias := range aliases {
		if val := p.Get(alias); val.Found {
			return val
		}
	}
	return &ArgValue{Value: "", Found: false}
}

// Pos retrieves a positional argument by index
func (p *Parser) Pos(idx int) (string, bool) {
	if idx < len(p.pos) {
		return p.pos[idx], true
	}
	return "", false
}

// AllPos returns all positional arguments
func (p *Parser) AllPos() []string {
	return p.pos
}

// PosCount returns count of positional arguments
func (p *Parser) PosCount() int {
	return len(p.pos)
}

// isFlag checks if a string is a flag
func isFlag(s string) bool {
	return strings.HasPrefix(s, "-") && len(s) > 1 && !isNumeric(s)
}

// isNumeric checks if a string is a number (to avoid treating -123 as a flag)
func isNumeric(s string) bool {
	if len(s) < 2 {
		return false
	}
	num := strings.TrimPrefix(s, "-")
	_, err := fmt.Sscanf(num, "%d", new(int))
	return err == nil
}

// logDebug outputs parsed arguments for debugging
func (p *Parser) logDebug() {
	if !p.debug {
		return
	}
	fmt.Println("\n[DEBUG] Argument Parser Results:")
	fmt.Println("Flags:")
	for k, v := range p.flags {
		fmt.Printf("  -%s: \"%s\" (found: %v)\n", k, v.Value, v.Found)
	}
	fmt.Println("Positional Arguments:")
	for i, v := range p.pos {
		fmt.Printf("  [%d]: %s\n", i, v)
	}
	fmt.Println()
}

// Legacy wrapper for backward compatibility
func getArgs(args []string, value string, flag bool) (string, bool) {
	for i := 0; i < len(args); i++ {
		if args[i] == value && !flag {
			if i+1 < len(args) {
				return args[i+1], true
			}
		} else if args[i] == value && flag {
			return "", true
		}
	}
	return "", false
}
