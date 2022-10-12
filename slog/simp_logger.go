package slog

import (
	"fmt"

	"github.com/fatih/color"
)

func Error(format string, a ...any) {
	fmt.Println(color.RedString(fmt.Sprintf(format, a...)))
}

func Warn(format string, a ...any) {
	fmt.Println(color.YellowString(fmt.Sprintf(format, a...)))
}

func Print(format string, a ...any) {
	fmt.Println(color.GreenString(fmt.Sprintf(format, a...)))
}
