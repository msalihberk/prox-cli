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
)

type B64Command struct{}

func (b B64Command) Execute(args []string) error {
	if len(args) < 2 {
		return errors.New("Base64 command requires 2 arguments. Usage: b64 <encode|decode> <string>")
	}
	if args[0] == "help" {
		PrintInfo("Base64 command requires 2 arguments. Usage: b64 <encode|decode> <string>")
		return nil
	}
	switch args[0] {
	case "encode":
		// Perform base64 encoding
		encoded := base64.StdEncoding.EncodeToString([]byte(args[1]))
		PrintSuccess("Encoded: %s", encoded)
	case "decode":
		// Perform base64 decoding
		decoded, err := base64.StdEncoding.DecodeString(args[1])
		if err != nil {
			return errors.New("Invalid base64 string")
		}
		PrintSuccess("Decoded: %s", string(decoded))
	default:
		return errors.New("Invalid operation. Use 'encode' or 'decode'")
	}
	return nil
}
func (b B64Command) Description() string {
	return "Encode or decode strings to/from Base64 format"
}
func init() {
	register("b64", B64Command{})
}
