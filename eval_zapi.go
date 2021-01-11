package main

import (
    "fmt"
    "os"
    "path/filepath"
	"local.host/api"
	"local.host/collector"
    "local.host/share"
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
    params.Timeout = 4

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

    cwd, _ := os.Getwd()

    var params = map[string]string {
        "harvest_path"     : cwd,
        "subtemplate"      : "disk.yaml",
        "subtemplate_dir"  : "default",
    }


	connection_params := get_params()

    template_path := filepath.Join(cwd, "var/zapi/default.yaml")
    template, err := share.ImportTemplate(template_path)
    if err != nil { panic(err) }

	zapi := collector.Zapi{ Class: "Zapi", Name: "Volume" }
    err = zapi.Init(params, template, connection_params)
    if err != nil { panic(err) }

}
