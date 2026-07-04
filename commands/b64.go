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
