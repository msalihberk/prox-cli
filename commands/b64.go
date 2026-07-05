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
	"encoding/base64"
	"errors"
	"os"
)

type B64Command struct{}

func (b B64Command) Execute(args []string) error {
	parser := New(args, false)
	parser.Parse()

	operation, ok := parser.Pos(0)
	if !ok {
		return errors.New("operation required: 'encode' or 'decode'. Try 'prox b64 help' for usage information.")
	}

	if operation == "help" || parser.GetAlias("h", "help").Found {
		PrintInfo("Usage: prox b64 <encode|decode> <string> [options]")
		PrintInfo("  prox b64 encode hello")
		PrintInfo("  prox b64 decode SGVsbG8=")
		PrintInfo("Options:")
		PrintInfo("  -i, --input <file>  : Read from file")
		PrintInfo("  -o, --output <file> : Write to file")
		return nil
	}

	var input string

	if str, ok := parser.Pos(1); ok {
		input = str
	} else if val := parser.GetAlias("i", "input"); val.Found {
		// Read from file
		data, err := os.ReadFile(val.Value)
		if err != nil {
			return errors.New("failed to read input file: " + err.Error())
		}
		input = string(data)
	} else {
		return errors.New("input string or file required. Try 'prox b64 help' for usage information.")
	}

	var result string

	switch operation {
	case "encode":
		result = base64.StdEncoding.EncodeToString([]byte(input))
	case "decode":
		decoded, err := base64.StdEncoding.DecodeString(input)
		if err != nil {
			return errors.New("invalid base64 string")
		}
		result = string(decoded)
	default:
		return errors.New("unknown operation: use 'encode' or 'decode'. Try 'prox b64 help' for usage information.")
	}

	if output := parser.GetAlias("o", "output"); output.Found {
		// Write to file
		err := os.WriteFile(output.Value, []byte(result), 0644)
		if err != nil {
			return errors.New("failed to write output file: " + err.Error())
		}
		PrintSuccess("%s: result written to %s", operation, output.Value)
	} else if !isPiped() {
		PrintSuccess("%s: %s", operation, result)
	} else {
		PrintInfo("%s", result)
	}

	return nil
}
func (b B64Command) Description() string {
	return "Encode or decode strings to/from Base64 format"
}
func init() {
	register("b64", B64Command{})
}
