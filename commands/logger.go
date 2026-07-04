package commands

import (
	"fmt"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
)

func PrintError(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("%sError: %s%s\n", colorRed, msg, colorReset)
}

func PrintInfo(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("%s%s%s\n", colorCyan, msg, colorReset)
}

func PrintSuccess(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("%s%s%s\n", colorGreen, msg, colorReset)
}

func PrintMessage(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("%s%s%s\n", colorReset, msg, colorReset)
}
func PrintWarning(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("%sWarning: %s%s\n", colorYellow, msg, colorReset)
}
func PrintNewLine() {
	fmt.Println()
}
