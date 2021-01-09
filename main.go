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

        root, err = lib.NewTree(response)
        if err != nil { panic(err) }

        //fmt.Printf("\n%s\n", string(response))

        lib.PrintTree(root, 0)
    }

    fmt.Println("bye!")
}
