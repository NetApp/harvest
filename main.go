package main

import (
    "fmt"
    "path/filepath"
    "local.host/params"
	"local.host/collector"
    "local.host/template"
)


func main() {

    var err error
    var p params.Params
    var t *template.Element
    var c *collector.Collector

	p = params.NewFromArgs()

    t, err = template.New(filepath.Join(p.Path, "var/zapi/", p.Template+".yaml" ))
    if err != nil { panic(err) }
    
    t.PrintTree(0)

    fmt.Printf("t = %v (%T)\n", t, t)
    fmt.Printf("&t = %v (%T)\n", &t, &t)
    fmt.Printf("*t = %v (%T)\n", *t, *t)

    c = collector.New("Zapi", p.Object)
    fmt.Printf("c = %v (%T)\n", c, c)
    fmt.Printf("&c = %v (%T)\n", &c, &c)
    fmt.Printf("*c = %v (%T)\n", *c, *c)    
    
    err = c.Init(p, t)
    if err != nil { panic(err) }

    /*
    err = c.PollData()
    if err != nil { panic(err) }
    */

    fmt.Println("SUCCESS")
}
