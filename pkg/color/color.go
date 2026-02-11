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
var Yellow = "\033[93m"
var Cyan = "\033[36m"
var Grey = "\033[90m"

var withColor = false

func DetectConsole(option string) {
	switch option {
	case "never":
		withColor = false
	case "always":
		withColor = true
	default:
		fd := int(os.Stdout.Fd()) // #nosec G115
		if term.IsTerminal(fd) {
			withColor = true
		}
	}
}

func Colorize(s any, color string) string {
	if withColor {
		return fmt.Sprintf("%s%v\x1b[0m", color, s)
	}
	return fmt.Sprintf("%v", s)
}
