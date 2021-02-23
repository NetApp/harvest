package main

import (
	"fmt"
	"goharvest2/share/options"
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

	opts := options.New("Harvest Options", "Quickly parse positional and named flags")

	opts.PosString(&t.Command, "command", "command, first positional argument", []string{"start", "stop", "status"})
	opts.String(&t.Poller, "poller", "p", "name of poller")
	opts.Bool(&t.Debug, "debug", "d", "run in debug mode")
	opts.Bool(&t.Verbose, "verbose", "v", "equal to loglevel=1")
	opts.Int(&t.Loglevel, "loglevel", "l", "logging level")
	opts.Slice(&t.Collectors, "collectors", "c", "names of collectors to start")

	ok := opts.Parse()

	fmt.Println("\n\n~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~\n\n")
	// Parse returns true if commands were parsed successfully
	if ok {
		fmt.Println("parsed successfully")
	// return false, if options were incorrect or user asked for help
	// this indicates we shouldn't continue the program
	} else {
		fmt.Println("parsed with errors or asked for help")
	}
	opts.PrintValues()
}