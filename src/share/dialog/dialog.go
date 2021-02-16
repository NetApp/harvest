package dialog

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

type Dialog struct {
	enabled bool
}

func New() Dialog {
	d := Dialog{}
	cmd := exec.Command("dialog", "--help")
	if err := cmd.Run(); err != nil {
		fmt.Println("dialog not enabled!")
		d.enabled = false
	} else {
		fmt.Println("dialog enabled!")
		d.enabled = true
	}
	return d
}

func (d Dialog) Menu(title, msg string, items... string) bool {
	cmd := exec.Command("dialog", "--title", escape(title), "--menu", escape(msg), "0", "0", strconv.Itoa(len(items)))
	for i, item := range items {
		cmd.Args = append(cmd.Args, strconv.Itoa(i+1))
		cmd.Args = append(cmd.Args, escape(item))
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	//fmt.Println("out:\n", out)
	fmt.Println("err:\n", err)

	//err = cmd.Wait()
	//fmt.Println("err:\n", err)

	return true
}

func (d Dialog) YesNo(title, msg string) bool {
	cmd := exec.Command("dialog", "--title", escape(title), "--yesno", escape(msg), "0", "0")


	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	fmt.Println(cmd.String())

	//stdout, _ := cmd.StdoutPipe()
	//stdin, _ := cmd.StderrPipe()

	err := cmd.Run()
	//fmt.Println("out:\n", out)
	fmt.Println("err:\n", err)

	//err = cmd.Wait()
	//fmt.Println("err:\n", err)

	return true
}

func escape(text string) string {
	return "\"" + text + "\""
}