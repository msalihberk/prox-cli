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
	"strconv"
)

type KeyGenCommand struct{}

func (v KeyGenCommand) Execute(args []string) error {
	if len(args) < 1 || len(args) > 5 {
		return errors.New("Usage: prox keygen <number> [A] [a] [0] [!] | [C <characters>]")
	}
	if args[0] == "help" {
		PrintInfo("Usage: prox keygen <number> [A] [a] [0] [!] | [C <characters>]")
		return nil
	}
	length, err := strconv.Atoi(args[0])
	if err != nil {
		return errors.New("Error: " + err.Error())
	}
	const charset_uppercase = "ABCDEFGHHIJKLMNOPRSTUVWXYZ"
	const charset_lowercase = "abcdefghijklmnoprstuvwxyz"
	const charset_number = "0123456789"
	const charset_special = "!@$+*%^&()-_=[]{}|;:,.<>/"
	charset := ""
	customCharSet := ""
	for j := 1; j < len(args); j++ {
		switch args[j] {
		case "A":
			charset = charset + charset_uppercase
		case "a":
			charset = charset + charset_lowercase
		case "0":
			charset = charset + charset_number
		case "!":
			charset = charset + charset_special
		default:
			customCharSet = args[j]
		}
	}
	if charset != "" {
		if customCharSet != "" {
			return errors.New("Usage: prox keygen <number> [A] [a] [0] [!] | [C <characters>]")
		}
	} else if customCharSet != "" {
		charset = customCharSet
	} else {
		charset = charset + charset_lowercase + charset_uppercase + charset_number + charset_special
	}

	password := make([]byte, length)
	charsetLength := big.NewInt(int64(len(charset)))

	for i := 0; i < length; i++ {
		randomIndex, err := rand.Int(rand.Reader, charsetLength)
		if err != nil {
			return errors.New("Error: " + err.Error())
		}
		password[i] = charset[randomIndex.Int64()]
	}
	PrintSuccess("Key Generated: %s", string(password))
	return nil
}
func (v KeyGenCommand) Description() string {
	return "Create random and secure key"
}
func init() {
	register("keygen", KeyGenCommand{})
}
