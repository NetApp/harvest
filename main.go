package main

import (
    "fmt"
    "path/filepath"
    "local.host/params"
    "local.host/template"
	"local.host/collector"
	"local.host/exporter"
    "local.host/matrix"
)


func main() {

    var err error
    var data *matrix.Matrix
    var p params.Params
    var t *template.Element
    var c *collector.Collector
    var e *exporter.Exporter

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


    err = c.PollInstance()
    if err != nil { panic(err) }

    data, err = c.PollData()
    if err != nil { panic(err) }

    e = exporter.New("Prometheus", "P_Andromeda")
    err = e.Init()
    if err != nil { panic(err) }

    err = e.Export(data, c.Template.GetChild("export_options"))
    if err != nil { panic(err) }
    fmt.Println("SUCCESS")

    data.Print()
}
