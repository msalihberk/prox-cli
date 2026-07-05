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
	"errors"
	"fmt"
	"math/rand"
	"prox-cli/core"
	"strconv"
	"strings"
	"time"
)

type LoremCommand struct{}

var loremWords = []string{
	"lorem", "ipsum", "dolor", "sit", "amet", "consectetur", "adipiscing", "elit",
	"sed", "do", "eiusmod", "tempor", "incididunt", "ut", "labore", "et", "dolore",
	"magna", "aliqua", "ut", "enim", "ad", "minim", "veniam", "quis", "nostrud",
	"exercitation", "ullamco", "laboris", "nisi", "ut", "aliquip", "ex", "ea",
	"commodo", "consequat", "duis", "aute", "irure", "dolor", "in", "reprehenderit",
	"in", "voluptate", "velit", "esse", "cillum", "dolore", "eu", "fugiat", "nulla",
	"pariatur", "excepteur", "sint", "occaecat", "cupidatat", "non", "proident",
	"sunt", "in", "culpa", "qui", "officia", "deserunt", "mollit", "anim", "id", "est", "laborum",
}

func generateParagraph(wordCount int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var words []string
	for i := 0; i < wordCount; i++ {
		words = append(words, loremWords[r.Intn(len(loremWords))])
	}
	sentence := strings.Join(words, " ")
	if len(sentence) > 0 {
		sentence = strings.ToUpper(string(sentence[0])) + sentence[1:] + "."
	}
	return sentence
}

func (v LoremCommand) Execute(args []string) error {
	parser := core.New(args, false)
	parser.Parse()

	operation, _ := parser.Pos(0)
	if operation == "help" || parser.GetAlias("h", "help").Found {
		core.PrintInfo("Usage: prox lorem [-w <words>] [-p <paragraphs>] [-c]")
		core.PrintInfo("  -w, --words       : Number of words per paragraph. Default: 50")
		core.PrintInfo("  -p, --paragraphs  : Number of paragraphs to generate. Default: 1")
		core.PrintInfo("  -c, --copy        : Copy generated text to clipboard automatically")
		return nil
	}

	words := 50
	if parser.GetAlias("w", "words").Found {
		valStr := parser.GetAlias("w", "words").Value
		val, err := strconv.Atoi(valStr)
		if err != nil {
			return fmt.Errorf("invalid value '%s' for --words: must be an integer", valStr)
		}
		if val <= 0 {
			return errors.New("number of words must be greater than 0")
		}
		words = val
	}

	paragraphs := 1
	if parser.GetAlias("p", "paragraphs").Found {
		valStr := parser.GetAlias("p", "paragraphs").Value
		val, err := strconv.Atoi(valStr)
		if err != nil {
			return fmt.Errorf("invalid value '%s' for --paragraphs: must be an integer", valStr)
		}
		if val <= 0 {
			return errors.New("number of paragraphs must be greater than 0")
		}
		paragraphs = val
	}

	var result strings.Builder
	for i := 0; i < paragraphs; i++ {
		result.WriteString(generateParagraph(words))
		if i < paragraphs-1 {
			result.WriteString("\n\n")
		}
	}

	output := result.String()

	if parser.GetAlias("c", "copy").Found {
		if err := core.CopyToClipboard(output); err != nil {
			if !core.IsPiped() {
				core.PrintMessage("%s", output)
				core.PrintError("Clipboard error: %s", err.Error())
			}
			return nil
		}
	}

	if core.IsPiped() {
		core.PrintInfo("%s", output)
	} else {
		core.PrintMessage("%s", output)
		if parser.GetAlias("c", "copy").Found {
			core.PrintSuccess("✓ Copied to clipboard!")
		}
	}

	return nil
}

func (v LoremCommand) Description() string {
	return "Generate dummy Lorem Ipsum text for testing and placeholders"
}
func (v LoremCommand) Help() string {
	help := "Usage: prox ai <command> [arguments]"
	help += "\n  cmd <prompt>     : Convert natural language to a one-liner terminal command"
	help += "\n  explain [text]   : Analyze and explain logs, code, or payloads (Supports piping)"
	return help
}
func init() {
	core.Register("lorem", LoremCommand{})
}
