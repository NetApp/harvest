package main

import (
	"fmt"
	"goharvest2/share/argparse"
)

type TestOptions struct {
	Command string
	Poller string
	Debug bool
	Verbose bool
	Loglevel int
	Collectors []string
} 

func (t TestOptions) Print() {
	fmt.Printf("command     = %s", t.Command)
	fmt.Printf("poller      = %s", t.Poller)
	fmt.Printf("debug       = %v", t.Debug)
	fmt.Printf("verbose     = %v", t.Verbose)
	fmt.Printf("loglevel    = %d", t.Loglevel)
	fmt.Printf("collectors  = %v", t.Collectors)
}

func main() {

	t := TestOptions{}

	p := argparse.New("Harvest Options Tester", "options.go", "Quickly parse positional and named flags")

	p.PosString(&t.Command, "command", "command, first positional argument", []string{"start", "stop", "status"})
	p.String(&t.Poller, "poller", "p", "name of poller")
	p.Bool(&t.Debug, "debug", "d", "run in debug mode")
	p.Bool(&t.Verbose, "verbose", "v", "equal to loglevel=1")
	p.Int(&t.Loglevel, "loglevel", "l", "logging level")
	p.Slice(&t.Collectors, "collectors", "c", "names of collectors to start")

	ok := p.Parse()

	fmt.Println("\n\n~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~\n\n")
	// Parse returns true if commands were parsed successfully
	if ok {
		fmt.Println("parsed successfully")
	// return false, if options were incorrect or user asked for help
	// this indicates we shouldn't continue the program
	} else {
		fmt.Println("parsed with errors or asked for help")
	}
	p.PrintValues()
}