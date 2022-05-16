/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package color

import (
	"fmt"
	"golang.org/x/term"
	"os"
)

var Bold = "\033[1m"
var End = "\033[0m"
var Green = "\033[92m"
var Red = "\033[31m"
var Pink = "\033[35m"
var Yellow = "\033[93m"
var Cyan = "\033[36m"
var Blue = "\033[96m"
var Grey = "\033[90m"

var withColor = false

func DetectConsole(option string) {
	switch option {
	case "never":
		withColor = false
	case "always":
		withColor = true
	default:
		if term.IsTerminal(int(os.Stdout.Fd())) {
			withColor = true
		}
	}
}

func Colorize(s interface{}, color string) string {
	if withColor {
		return fmt.Sprintf("%s%v\x1b[0m", color, s)
	}
	return fmt.Sprintf("%s", s)
}
