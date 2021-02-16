package main

import (
	"fmt"
	"goharvest2/share/options"
)

func main() {

	opts := options.New("Harvest Config", "Quickly configure Pollers and Exporters")

	opts.AddBool("debug", "d", "run in debug mode", false)
	opts.AddInt("loglevel", "l", "logging level", 2)
	opts.AddString("poller", "p", "name of poller", "")
	opts.AddSlice("collectors", "c", "names of collectors to start")
	opts.AddBool("help", "h", "print this message", false)

	if opts.Parse() {
		fmt.Println("parsed successfully")
	} else {
		fmt.Println("parsed with errors")
		opts.PrintErrors()
	}

	if help, _ := opts.GetBool("help"); help {
		opts.PrintHelp()
	} else {
		opts.PrintValues()
	}
}