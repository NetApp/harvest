
package params

import (
    "os"
    "strings"
    "fmt"
)

type Params struct {
    Hostname string
    UseCert bool
    Authorization [2]string
    Timeout int
    Path string
    Object string
    Template string
    Subtemplate string
}


func NewFromArgs() Params {

	var params Params
    var cwd string

    if len(os.Args) < 6 {
        fmt.Println("Expected 4 args:")
        fmt.Println(" hostname \"cert\" cert_file key_file object [subtemplate]")
        fmt.Println("OR:")
        fmt.Println(" hostname \"basic\" username password object [subtemplate]")
        os.Exit(1)
    }

    cwd, _ = os.Getwd()

	params = Params{}
    params.Hostname = os.Args[1]
    params.Authorization[0] = os.Args[3]
    params.Authorization[1] = os.Args[4]
    params.Timeout = 5
    params.Object = strings.Title(os.Args[5])
    params.Path = cwd
    params.Template = "default"

    if len(os.Args) == 7 {
        params.Subtemplate = os.Args[6]
    } else {
        params.Subtemplate = strings.ToLower(os.Args[5]) + ".yaml"
    }

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