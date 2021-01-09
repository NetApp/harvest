package main

import (
    "fmt"
    "os"
    "local.host/lib"
)

func main() {

    var api, hostname string
    var auth_style bool
    var auth [2]string
    var connection lib.Connection
    var response []byte
    var err error
    var root *lib.Node

    if len(os.Args) != 5 {
        fmt.Println("Expected 4 args:")
        fmt.Println(" hostname \"cert\" cert_file key_file")
        fmt.Println("OR:")
        fmt.Println(" hostname \"basic\" username password")
        os.Exit(1)
    }

    hostname = os.Args[1]
    auth[0] = os.Args[3]
    auth[1] = os.Args[4]

    if os.Args[2] == "cert" {
        auth_style = true
    } else if os.Args[2] == "basic" {
        auth_style = false
    } else {
        fmt.Println("Invalid authentication style")
    }

    connection, err = lib.NewConnection(hostname, auth_style, auth, 10)

    if err != nil { panic(err) }

    for {
        fmt.Print("Enter \"exit\" or API: ")
        fmt.Scanf("%s", &api)

        if api == "exit" { break }

        response, err = connection.InvokeAPI(api)

        if err != nil { panic(err) }

        root, err = lib.Parse(response)
        if err != nil { panic(err) }

        //fmt.Printf("\n%s\n", string(response))

        lib.PrintTree(root, 0)
    }

    fmt.Println("bye!")

    fmt.Println("Time to build our own XML!!!!")

    node := lib.NewNode("perf-object-counter-list-info")
    node.CreateChild("objectname", "volume")

    xml, err := node.Build()
    if err != nil { panic(err) }
    fmt.Println(string(xml))

    response, err = connection.InvokeAPI(string(xml))
    if err != nil { panic(err) }

    root, err = lib.Parse(response)
    if err != nil {
        fmt.Println("Failed to parse XML response")
        panic(err)
    }

    lib.PrintTree(root, 0)

    fmt.Printf("\n\nRoot has name: %s\n", root.GetName())

    child, found := root.GetChild("results")
    if found == true {
        fmt.Printf("Root has child results\n")
    } else {
        fmt.Printf("Root has NO child results\n")
    }
    fmt.Println(child)

    content, found := child.GetChildContent("counters")
    if found == true {
        fmt.Printf("Child has counters\n")
        fmt.Println(content)
    } else {
        fmt.Printf("Child has NO counters\n")
    }
}
