/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package color

import (
	"fmt"
	"golang.org/x/term"
	"os"
)

var Bold string = "\033[1m"
var End string = "\033[0m"
var Italic string = "\033[3m"
var Green string = "\033[92m"
var Red string = "\033[31m"
var Pink string = "\033[35m"
var Yellow string = "\033[93m"
var Cyan string = "\033[36m"
var Blue string = "\033[96m"
var Grey string = "\033[90m"
var BlueBG string = "\033[46m"
var GreenBG string = "\033[42m"
var RedBG string = "\033[41m"
var PinkBG string = "\033[45m"

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
