package main

import (
    "fmt"
    "os"
    "poller/yaml"
)

func main() {
    if len(os.Args) == 1 {
        fmt.Println("Provide path to yaml file")
        os.Exit(0)
    }

    x, err := yaml.Import(os.Args[1])
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }

    x.PrintTree(0)
}
