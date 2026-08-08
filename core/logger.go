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
	"fmt"
	"os"
	"strings"
)

const (
	colorReset   = "\033[0m"
	colorBold    = "\033[1m"
	colorDim     = "\033[2m"
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorBlue    = "\033[34m"
	colorCyan    = "\033[36m"
	colorMagenta = "\033[35m"
	colorWhite   = "\033[37m"
)

func IsPiped() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) == 0
}

func styleText(text, color string) string {
	if IsPiped() {
		return text
	}
	return color + text + colorReset
}

func PrintError(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	if IsPiped() {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", msg)
	} else {
		fmt.Fprintf(os.Stderr, "%s %s\n", styleText("✖", colorRed+colorBold), msg)
	}
}

func PrintInfo(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	if IsPiped() {
		fmt.Print(msg)
	} else {
		fmt.Printf("%s %s\n", styleText("ℹ", colorCyan+colorBold), msg)
	}
}

func PrintSuccess(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	if IsPiped() {
		fmt.Print(msg)
	} else {
		fmt.Printf("%s %s\n", styleText("✔", colorGreen+colorBold), msg)
	}
}

func PrintMessage(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	if IsPiped() {
		fmt.Print(msg)
	} else {
		fmt.Print(msg)
		if !strings.HasSuffix(msg, "\n") {
			fmt.Println()
		}
	}
}

func PrintWarning(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	if IsPiped() {
		fmt.Fprintf(os.Stderr, "WARN: %s\n", msg)
	} else {
		fmt.Fprintf(os.Stderr, "%s %s\n", styleText("⚠", colorYellow+colorBold), msg)
	}
}

func PrintNewLine() {
	if !IsPiped() {
		fmt.Println()
	}
}
