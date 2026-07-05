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
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"hash"
	"io"
	"os"
	"strings"
)

type HashCommand struct{}

func calculateHash(reader io.Reader, parser *Parser) (string, string, error) {
	var hasher hash.Hash

	if parser.GetAlias("b", "sha512").Found {
		hasher = sha512.New()
		return "SHA512", runHash(hasher, reader), nil
	} else if parser.GetAlias("c", "md5").Found {
		hasher = md5.New()
		return "MD5", runHash(hasher, reader), nil
	} else if parser.GetAlias("d", "sha1").Found {
		hasher = sha1.New()
		return "SHA1", runHash(hasher, reader), nil
	} else {
		hasher = sha256.New()
		return "SHA256", runHash(hasher, reader), nil
	}
}

func runHash(h hash.Hash, reader io.Reader) string {
	if _, err := io.Copy(h, reader); err != nil {
		return ""
	}
	return hex.EncodeToString(h.Sum(nil))
}

func (v HashCommand) Execute(args []string) error {
	parser := New(args, false)
	parser.Parse()

	if len(args) == 0 {
		return errors.New("hash command requires an input string or file path. Try 'prox hash help' for usage information.")
	}

	operation, _ := parser.Pos(0)
	if operation == "help" || parser.GetAlias("h", "help").Found {
		PrintInfo("%s", v.Help())
		return nil
	}

	var reader io.Reader
	var algoName, hashResult string
	var err error

	if parser.GetAlias("s", "string").Found {
		inputStr := parser.GetAlias("s", "string").Value
		reader = strings.NewReader(inputStr)
		algoName, hashResult, err = calculateHash(reader, parser)
	} else if parser.GetAlias("f", "file").Found {
		filePath := parser.GetAlias("f", "file").Value
		file, err := os.Open(filePath)
		if err != nil {
			return errors.New("failed to open input file: " + err.Error())
		}
		defer file.Close()

		reader = file
		algoName, hashResult, err = calculateHash(reader, parser)
	} else {
		return errors.New("please specify an input using -s <string> or -f <file>")
	}

	if err != nil || hashResult == "" {
		return errors.New("failed to compute hash")
	}

	if isPiped() {
		PrintInfo("%s", hashResult)
	} else {
		PrintInfo("%s: %s", algoName, hashResult)
	}
	return nil
}

func (v HashCommand) Description() string {
	return "Compute the hash of a given input string or file securely"
}
func (v HashCommand) Help() string {
	help := "Usage: prox hash -s <string> | -f <file> [-a] [-b] [-c] [-d]"
	help += "\n  -a, --sha256      : Compute SHA256 hash (Default)"
	help += "\n  -b, --sha512      : Compute SHA512 hash"
	help += "\n  -c, --md5         : Compute MD5 hash"
	help += "\n  -d, --sha1        : Compute SHA1 hash"
	return help
}
func init() {
	register("hash", HashCommand{})
}
