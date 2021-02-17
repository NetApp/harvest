package dialog

import (
	"fmt"
	"os"
	"os/exec"
	"io/ioutil"
	"strconv"
    "errors"
)

type Dialog struct {
	enabled bool
	title string
	cmd *exec.Cmd
}

func New() *Dialog {

	// use default title
	d := Dialog{title: "harvest 2.0 - config"}

	// is dialog available?
	cmd := exec.Command("dialog", "--help")
	if err := cmd.Run(); err != nil {
		d.enabled = false
		fmt.Printf("dialog not enabled: %v", err)
	} else {
		d.enabled = true
		fmt.Println("dialog enabled!")
	}
	return &d
}

func (d *Dialog) SetTitle(title string) {
	d.title = title
}

func (d *Dialog) Close() {
	d.setArgs("--clear")
	d.exec()
}

func (d *Dialog) setArgs(args... string) {
	d.cmd = exec.Command("dialog", args...)
}

func (d *Dialog) addArg(arg string) {
	d.cmd.Args = append(d.cmd.Args, arg)
}

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

func (d *Dialog) Message(msg string) {
	d.setArgs("--msgbox", msg, "0", "0")
    d.exec()
}

func (d *Dialog) Input(msg string) (string, error) {
    d.setArgs("--inputbox", msg, "0", "0")
    return d.exec()
}

func (d *Dialog) Password(msg string) (string, error) {
    d.setArgs("--passwordbox", msg, "0", "0")
    return d.exec()
}

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

func (d *Dialog) YesNo(msg string) bool {
	d.setArgs("--yesno", msg, "0", "0")
    if _, err := d.exec(); err != nil {
        return false
    }
    return true
}

