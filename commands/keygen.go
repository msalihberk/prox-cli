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
	"crypto/rand"
	"errors"
	"math/big"
	"prox-cli/core"
	"strconv"
)

type KeyGenCommand struct{}

func (v KeyGenCommand) Execute(args []string) error {
	parser := core.New(args, false)
	parser.Parse()

	lengthStr, ok := parser.Pos(0)

	if lengthStr == "help" || parser.GetAlias("h", "help").Found {
		core.PrintInfo("Usage: prox keygen <length> [-A] [-a] [-N] [-S] [-c custom_chars]")
		core.PrintInfo("  -A, --uppercase   : Include uppercase letters")
		core.PrintInfo("  -a, --lowercase   : Include lowercase letters")
		core.PrintInfo("  -N, --numbers     : Include numbers")
		core.PrintInfo("  -S, --special     : Include special characters")
		core.PrintInfo("  -c, --chars <str> : Use custom character set")
		return nil
	}
	if !ok {
		return errors.New("length argument required. Try 'prox keygen help' for usage information.")
	}

	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return errors.New("invalid length: " + err.Error())
	}

	const charset_uppercase = "ABCDEFGHHIJKLMNOPRSTUVWXYZ"
	const charset_lowercase = "abcdefghijklmnoprstuvwxyz"
	const charset_number = "0123456789"
	const charset_special = "!@$+*%^&()-_=[]{}|;:,.<>/"

	charset := ""

	if parser.GetAlias("c", "chars").Found {
		charset = parser.GetAlias("c", "chars").Value
	} else {
		if parser.GetAlias("A", "uppercase").Found {
			charset += charset_uppercase
		}
		if parser.GetAlias("a", "lowercase").Found {
			charset += charset_lowercase
		}
		if parser.GetAlias("N", "numbers").Found {
			charset += charset_number
		}
		if parser.GetAlias("S", "special").Found {
			charset += charset_special
		}
		if charset == "" {
			charset = charset_lowercase + charset_uppercase + charset_number + charset_special
		}
	}

	password := make([]byte, length)
	charsetLength := big.NewInt(int64(len(charset)))

	for i := 0; i < length; i++ {
		randomIndex, err := rand.Int(rand.Reader, charsetLength)
		if err != nil {
			return errors.New("random generation error: " + err.Error())
		}
		password[i] = charset[randomIndex.Int64()]
	}

	if core.IsPiped() {
		core.PrintInfo("%s", string(password))
	} else {
		core.PrintSuccess("Key Generated: %s", string(password))
	}
	return nil
}
func (v KeyGenCommand) Description() string {
	return "Create random and secure key"
}
func (v KeyGenCommand) Help() string {
	help := "Usage: prox ai <command> [arguments]"
	help += "\n  cmd <prompt>     : Convert natural language to a one-liner terminal command"
	help += "\n  explain [text]   : Analyze and explain logs, code, or payloads (Supports piping)"
	return help
}
func init() {
	core.Register("keygen", KeyGenCommand{})
}
