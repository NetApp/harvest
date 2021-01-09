package main

import (
    "fmt"
    "os"
	"local.host/api"
	"local.host/collector"
)

func get_params() api.ConnectionParams {

	var params api.ConnectionParams

	params = api.ConnectionParams{}

    if len(os.Args) != 5 {
        fmt.Println("Expected 4 args:")
        fmt.Println(" hostname \"cert\" cert_file key_file")
        fmt.Println("OR:")
        fmt.Println(" hostname \"basic\" username password")
        os.Exit(1)
    }

    params.Hostname = os.Args[1]
    params.Authorization[0] = os.Args[3]
    params.Authorization[1] = os.Args[4]

    if os.Args[2] == "cert" {
        params.UseCert = true
    } else if os.Args[2] == "basic" {
        params.UseCert = false
    } else {
		fmt.Println("Invalid authentication style")
		os.Exit(1)
	}
	return params
}


func main() {
	params := get_params()

	zapi := collector.Zapi{ Class: "Zapi", Name: "Volume" }

	zapi.Init(params)
}
