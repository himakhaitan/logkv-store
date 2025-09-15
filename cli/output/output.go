package output

import (
	"fmt"
	"os"
)

// ANSI color codes
const (
	reset = "\033[0m"
	bold  = "\033[1m"

	red    = "\033[31m"
	yellow = "\033[33m"
	blue   = "\033[34m"
	green  = "\033[32m"
	cyan   = "\033[36m"
	grey   = "\033[90m"
)

// core printer
func printMessage(title, color, message string) {
	fmt.Fprintf(os.Stdout, "%s%s[%s]%s %s%s%s\n",
		color, bold, title, reset, color, message, reset,
	)
}

// Public functions

func Info(msg string) {
	printMessage("INFO", blue, msg)
}

func Warn(msg string) {
	printMessage("WARN", yellow, msg)
}

func Error(msg string) {
	printMessage("ERROR", red, msg)
}

func Success(msg string) {
	printMessage("SUCCESS", green, msg)
}

func Debug(msg string) {
	printMessage("DEBUG", cyan, msg)
}

func Dim(msg string) {
	fmt.Fprintf(os.Stdout, "%s%s%s\n", grey, msg, reset)
}
