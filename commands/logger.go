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
	"fmt"
	"os"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
)

func isPiped() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) == 0
}

func PrintError(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Fprintf(os.Stderr, "%sError: %s%s\n", colorRed, msg, colorReset)
}

func PrintInfo(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	if isPiped() {
		fmt.Print(msg)
	} else {
		fmt.Printf("%s%s%s\n", colorCyan, msg, colorReset)
	}
}

func PrintSuccess(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	if isPiped() {
		fmt.Print(msg)
	} else {
		fmt.Printf("%s%s%s\n", colorGreen, msg, colorReset)
	}
}

func PrintMessage(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	if isPiped() {
		fmt.Print(msg)
	} else {
		fmt.Printf("%s%s%s\n", colorReset, msg, colorReset)
	}
}

func PrintWarning(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Fprintf(os.Stderr, "%sWarning: %s%s\n", colorYellow, msg, colorReset)
}

func PrintNewLine() {
	if !isPiped() {
		fmt.Println()
	}
}
