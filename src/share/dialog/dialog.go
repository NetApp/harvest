package dialog

import (
	"fmt"
	"os"
	"os/exec"
	"io/ioutil"
	"strconv"
    "errors"
    "strings"
)

/*
    Text-based UI for interacting with user. Uses whiptail or dialog if
    available, otherwise will fall back to STDIN/STDOUT (not implemented yet)

    Note that whiptail is usually avaible on RHEL/CentOs and similar systems,
    while dialog is more typical for Debian and related systems. Both use
    the same set of flags, so we can easily switch between the two commands
    (tested at least for the flags used in our methods).
*/

type Dialog struct {
	enabled bool
	title string
    bin string
	cmd *exec.Cmd
}

func New() *Dialog {

	// use default title
	d := Dialog{title: "harvest 2.0 - config"}

	// whiptail or dialog available?

    d.enabled = true

    if out, err := exec.Command("which", "whiptail").Output(); err == nil {
        d.bin = strings.TrimSpace(string(out))
    } else if out, err := exec.Command("which", "dialog").Output(); err == nil {
        d.bin = strings.TrimSpace(string(out))
    } else {
        d.enabled = false
    }

    return &d
}

// init new process with given args
func (d *Dialog) setArgs(args... string) {
	d.cmd = exec.Command(d.bin, args...)
}

// add arg to new process, only use after setArgs()
// and before exec(), otherwise program will panic
func (d *Dialog) addArg(arg string) {
	d.cmd.Args = append(d.cmd.Args, arg)
}

// execute process and return user response
// from whiptail / dialog.
// @TODO handle situation when d.enabled == false
func (d *Dialog) exec() (string, error) {
    os.Stdout.Sync()
    d.cmd.Stdout = os.Stdout

    stderr, err := d.cmd.StderrPipe()
    if err != nil {
        return "", err
    }

    if err = d.cmd.Start(); err != nil {
        return "", err
    }

    out, err := ioutil.ReadAll(stderr)
    if err != nil {
        return "", err
    }

    if err = d.cmd.Wait(); err != nil {
        return "", err
    }

    return string(out), nil
}

// Info about the dialog struct, for debugging
func (d *Dialog) Info() string {
    if d.enabled {
        return fmt.Sprintf("enabled, using binary [%s]", d.bin)
    } else {
        return "disabled, using StdIn/StdOut"
    }
}

func (d *Dialog) Enabled() bool {
    return d.enabled
}

// clear screen, good to call this function after last message
func (d *Dialog) Close() {
	d.setArgs("--clear")
	d.exec()
}

// change default title that's display in whiptail/display
func (d *Dialog) SetTitle(title string) {
	d.title = title
}

// show message to user
func (d *Dialog) Message(msg string) {
	d.setArgs("--msgbox", msg, "0", "0")
    d.exec()
}

// get input from user
func (d *Dialog) Input(msg string) (string, error) {
    d.setArgs("--inputbox", msg, "0", "0")
    return d.exec()
}

// get password as input
func (d *Dialog) Password(msg string) (string, error) {
    d.setArgs("--passwordbox", msg, "8", "30")
    return d.exec()
}

// get user choice from menu items
func (d *Dialog) Menu(msg string, items... string) (string, error) {
	d.setArgs("--menu", msg, "0", "0", strconv.Itoa(len(items)))
	for i, item := range items {
		d.addArg(strconv.Itoa(i))
		d.addArg(item)
	}

    out, err := d.exec()
    if err != nil {
        return "", err
    }

    index, err := strconv.Atoi(out)
    if err != nil {
        return "", err
    }

    if index < 0 || index >= len(items) {
        return "", errors.New("invalid choice")
    }
    return items[index], nil
}

// get consent from user
func (d *Dialog) YesNo(msg string) bool {
	d.setArgs("--yesno", msg, "0", "0")
    if _, err := d.exec(); err != nil {
        return false
    }
    return true
}

