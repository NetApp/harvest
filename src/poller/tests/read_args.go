package main

import (
    "fmt"
    "goharvest2/poller/structs/options"
)

func main() {
    args, _, err := options.GetOpts()
    if err == nil {
        args.Print()
    } else {
        fmt.Println(err)
    }
}