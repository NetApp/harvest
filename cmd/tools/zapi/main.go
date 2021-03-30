//
// Copyright NetApp Inc, 2021 All rights reserved
//
// Package Description:
//
// Examples:
//
package main

import (
    "fmt"
    "os"
    "path"
    "strings"

    client "goharvest2/pkg/apis/zapi"

    "goharvest2/pkg/argparse"
    "goharvest2/pkg/config"
    "goharvest2/pkg/set"
    "goharvest2/pkg/tree"
    "goharvest2/pkg/tree/node"
    "goharvest2/pkg/util"
)

var args *Args
var params *node.Node
var connection *client.Client
var system *client.System
var confpath string

var MAX_SEARCH_DEPTH = 10

var KNOWN_TYPES = set.NewFrom([]string{
    "string",
    "integer",
    "boolean",
    "node-name",
    "aggr-name",
    "vserver-name",
    "volume-name",
    "uuid", "size",
    "cache-policy",
    "junction-path",
    "volstyle",
    "repos-constituent-role",
    "language-code",
    "snaplocktype",
    "space-slo-enum",
})

type Args struct {
    Command string
    Item    string
    Poller  string
    Api     string
    Attr    string
    Object  string
    Counter string
    Export  bool
}

func (a *Args) Print() {
    fmt.Printf("command = %s\n", a.Command)
    fmt.Printf("item    = %s\n", a.Item)
    fmt.Printf("poller  = %s\n", a.Poller)
    fmt.Printf("api     = %s\n", a.Api)
    fmt.Printf("object  = %s\n", a.Object)
    fmt.Printf("attr    = %s\n", a.Attr)
    fmt.Printf("counter = %s\n", a.Counter)
}

type counter struct {
    name        string
    info        string
    scalar      bool
    base        string
    unit        string
    properties  []string
    labels1     []string
    labels2     []string
    privilege   string
    deprecated  bool
    replacement string
    key         bool
}

type attribute struct {
    Name     string
    Type     string
    Children []*attribute
}

func newAttr(Name, Type string) *attribute {
    return &attribute{Name: Name, Type: Type}
}

func (a *attribute) newChild(Name, Type string) *attribute {
    child := newAttr(Name, Type)
    a.Children = append(a.Children, child)
    return child
}

func (a *attribute) Print(depth int) {
    fmt.Printf("%s%s%s%-50s%s - %s%35s%s\n", strings.Repeat("  ", depth), util.Bold, util.Cyan, a.Name, util.End, util.Green, a.Type, util.End)
    for _, ch := range a.Children {
        ch.Print(depth + 1)
    }
}

func (c *counter) print_header() {

    fmt.Printf("\n%s%-30s %-10s %-15s %-30s %20s %15s %20s%s\n\n", util.Bold, "name", "scalar", "unit", "properties", "base counter", "deprecated", "replacement", util.End)
}
func (c *counter) print() {
    fmt.Printf("%s%s%-30s%s %-10v %-15s %-30s %20s %15v %20s\n", util.Bold, util.Cyan, c.name, util.End, c.scalar, c.unit, strings.Join(c.properties, ", "), c.base, c.deprecated, c.replacement)
    if !c.scalar {
        fmt.Printf("%sarray labels 1D%s: %s%v%s\n", util.Pink, util.End, util.Grey, c.labels1, util.End)
        if len(c.labels2) > 0 {
            fmt.Printf("%sarray labels 2D%s: %s%v%s\n", util.Pink, util.End, util.Grey, c.labels2, util.End)
        }
    }
    fmt.Printf("%s%s%s\n", util.Yellow, c.info, util.End)
}

func main() {

    var err error

    // set harvest config path
    if confpath = os.Getenv("HARVEST_CONF"); confpath == "" {
        confpath = "/etc/harvest"
    }

    // define arguments
    args = &Args{}
    parser := argparse.New("Zapi Utility", "harvest zapi", "Explore ZAPI of cDot or 7Mode system")

    parser.PosString(
        &args.Command,
        "command",
        "command",
        []string{"show", "export"},
    )

    parser.PosString(
        &args.Item,
        "item",
        "item to show",
        []string{"data", "apis", "attrs", "objects", "instances", "counters", "counter", "system"},
    )

    parser.String(
        &args.Poller,
        "poller",
        "p",
        "name of poller (cluster), as defined in your harvest config",
    )

    parser.String(
        &args.Api,
        "api",
        "a",
        "API (ZAPI query) to show",
    )

    parser.String(
        &args.Attr,
        "attr",
        "t",
        "ZAPI attribute to show",
    )

    parser.String(
        &args.Object,
        "object",
        "o",
        "ZapiPerf object to show",
    )

    parser.String(
        &args.Counter,
        "counter",
        "c",
        "ZapiPerf counter to show",
    )

    parser.SetHelpFlag("help")

    if !parser.Parse() {
        os.Exit(0)
    }

    if args.Command == "export" {
        args.Export = true
    }

    // validate and warn on missing arguments
    ok := true

    if args.Poller == "" {
        fmt.Println("missing required argument: --poller")
        ok = false
    }

    if args.Item == "data" && args.Api == "" && args.Object == "" {
        fmt.Println("show data: requires --api or --object")
        ok = false
    }

    if args.Item == "attrs" && args.Api == "" {
        fmt.Println("show attrs: requires --api")
        ok = false
    }

    if args.Item == "counters" && args.Object == "" {
        fmt.Println("show counters: requires --object")
        ok = false
    }

    if args.Item == "instances" && args.Object == "" {
        fmt.Println("show instances: requires --object")
        ok = false
    }

    if args.Item == "counter" && (args.Object == "" || args.Counter == "") {
        fmt.Println("show counter: requires --object and --counter")
        ok = false
    }

    if !ok {
        os.Exit(1)
    }

    // connect to cluster and retrieve system version
    if params, err = config.GetPoller(path.Join(confpath, "harvest.yml"), args.Poller); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    if connection, err = client.New(params); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    if system, err = connection.GetSystem(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    switch args.Item {

    //"data", "apis", "attrs", "objects", "counters", "counter", "system"},

    case "system":
        fmt.Println(system.String())
    case "data":
        get_data()
    case "apis":
        get_apis()
    case "attrs":
        get_attrs()
    case "objects":
        get_objects()
    case "counters":
        get_counters()
    case "instances":
        get_instances()
    case "counter":
        get_counter()
    default:
        fmt.Printf("invalid item: %s\n", args.Item)
        os.Exit(1)
    }
}

func get_apis() {

    if err := connection.BuildRequestString("system-api-list"); err != nil {
        fmt.Println(err)
        return
    }

    results, err := connection.Invoke()
    if err != nil {
        fmt.Println(err)
        return
    }

    apis := results.GetChildS("apis")
    if apis == nil {
        fmt.Println("Missing [apis] element in response")
        return
    }

    fmt.Printf("%s%s%-70s %s %20s %15s\n\n", util.Bold, util.Pink, "API", util.End, "LICENSE", "STREAM")
    for _, a := range apis.GetChildren() {
        name := a.GetChildContentS("name")
        license := a.GetChildContentS("license")
        streaming := a.GetChildContentS("is-streaming")
        fmt.Printf("%s%s%-70s %s %20s %15s\n", util.Bold, util.Pink, name, util.End, license, streaming)
    }
}

func get_objects() {

    if err := connection.BuildRequestString("perf-object-list-info"); err != nil {
        fmt.Println(err)
        return
    }

    results, err := connection.Invoke()
    if err != nil {
        fmt.Println(err)
        return
    }

    objects := results.GetChildS("objects")
    if objects == nil {
        fmt.Println("Missing [objects] element in response")
        return
    }

    fmt.Printf("%s%s%-50s %s %s %-20s %15s %15s %15s %s\n\n", util.Bold, util.Blue, "OBJECT", util.End, util.Bold, "PREFERRED KEY", "PRIV-LVL", "DEPREC", "REPLC", util.End)
    for _, o := range objects.GetChildren() {
        fmt.Printf("\n%s%s%-50s %s %s %-20s %15s %15s %15s %s\n",
            util.Bold,
            util.Blue,
            o.GetChildContentS("name"),
            util.End,
            util.Bold,
            o.GetChildContentS("get-instances-preferred-counter"),
            o.GetChildContentS("privilege-level"),
            o.GetChildContentS("is-deprecated"),
            o.GetChildContentS("replaced-by"),
            util.End,
        )
        fmt.Printf("%s     %s%s\n", util.Grey, o.GetChildContentS("description"), util.End)
    }
}

func get_instances() {

    var request, results *node.Node
    var count int
    var tag string
    var err error

    request = node.NewXmlS("perf-object-instance-list-info-iter")
    request.NewChildS("objectname", args.Object)
    request.NewChildS("max-records", "10")

    tag = "initial"

    for {

        results, tag, err = connection.InvokeBatchRequest(request, tag)

        fmt.Println("tag = ", tag)

        if err != nil {
            fmt.Println(err)
            break
        }

        if results == nil {
            break
        }

        if attrs := results.GetChildS("attributes-list"); attrs != nil {
            count += len(attrs.GetChildren())
            attrs.Print(0)
        }

        if tag == "" {
            break
        }
    }

    fmt.Printf("displayed %d [%s] instances\n", count, args.Object)

}

func get_data() {

    var request *node.Node

    request = node.NewXmlS(args.Api)

    if args.Object != "" {
        fmt.Printf("fetching raw data of zapiperf object [%s]\n", args.Object)

        if system.Clustered {
            request = node.NewXmlS("perf-object-get-instances")
            //request.NewChildS("max-records", "100")
            instances := request.NewChildS("instances", "")
            instances.NewChildS("instance", "*")
        } else {
            request = node.NewXmlS("perf-object-get-instances")
        }
        request.NewChildS("objectname", args.Object)
    } else {
        fmt.Printf("fetching raw data of zapi api [%s]\n", args.Api)
    }

    if err := connection.BuildRequest(request); err != nil {
        fmt.Println(err)
        return
    }

    results, err := connection.Invoke()
    if err != nil {
        fmt.Println(err)
    } else {
        results.Print(0)
    }
}

func get_attrs() {

    query := node.NewXmlS("system-api-get-elements")
    api_list := query.NewChildS("api-list", "")
    api_list.NewChildS("api-list-info", args.Api)

    if err := connection.BuildRequest(query); err != nil {
        fmt.Println(err)
        return
    }

    results, err := connection.Invoke()
    if err != nil {
        fmt.Println(err)
        return
    }

    output := node.NewS("output")
    input := node.NewS("input")

    if entries := results.GetChildS("api-entries"); entries != nil && len(entries.GetChildren()) > 0 {
        if elements := entries.GetChildren()[0].GetChildS("api-elements"); elements != nil {
            for _, x := range elements.GetChildren() {
                if x.GetChildContentS("is-output") == "true" {
                    x.PopChildS("is-output")
                    output.AddChild(x)
                } else {
                    input.AddChild(x)
                }
            }
        }
    }

    fmt.Println("############################        INPUT        ##########################")
    input.Print(0)
    fmt.Println()
    fmt.Println()

    fmt.Println("############################        OUPUT        ##########################")
    output.Print(0)
    fmt.Println()
    fmt.Println()

    // fetch root attribute
    attr_key := ""
    attr_name := ""

    for _, x := range output.GetChildren() {
        if t := x.GetChildContentS("type"); t == "string" || t == "integer" {
            continue
        }
        if name := x.GetChildContentS("name"); name != "num-records" || name != "next-tag" {
            attr_key = name
            attr_name = x.GetChildContentS("type")
            break
        }
    }

    if attr_name == "" {
        fmt.Println("no root attribute, stopping here.")
        return
    }

    if strings.HasSuffix(attr_name, "[]") {
        attr_name = strings.TrimSuffix(attr_name, "[]")
    }

    fmt.Printf("building tree for attribute [%s] => [%s]\n", attr_key, attr_name)

    if err = connection.BuildRequestString("system-api-list-types"); err != nil {
        fmt.Println(err)
        return
    }

    if results, err = connection.Invoke(); err != nil {
        fmt.Println(err)
        return
    }

    entries := results.GetChildS("type-entries")
    if entries == nil {
        fmt.Println("Error: missing [type-entries")
        return
    }

    attr := node.NewS(attr_name)
    search_entries(attr, entries)

    fmt.Println("############################        ATTR         ##########################")
    attr.Print(0)
    fmt.Println()
    if args.Export {
        fn := path.Join("/tmp", args.Api+".yml")
        if err = tree.Export(attr, "yaml", fn); err != nil {
            fmt.Printf("failed to export to [%s]:\n", fn)
            fmt.Println(err)
        } else {
            fmt.Printf("exported to [%s]\n", fn)
        }
    }
}

func search_entries(root, entries *node.Node) {

    cache := make(map[string]*node.Node)
    cache[root.GetNameS()] = root

    for i := 0; i < MAX_SEARCH_DEPTH; i += 1 {
        for _, entry := range entries.GetChildren() {
            name := entry.GetChildContentS("name")
            if parent, ok := cache[name]; ok {
                delete(cache, name)
                if elems := entry.GetChildS("type-elements"); elems != nil {
                    for _, elem := range elems.GetChildren() {
                        child := parent.NewChildS(elem.GetChildContentS("name"), "")
                        attr_type := strings.TrimSuffix(elem.GetChildContentS("type"), "[]")
                        if !KNOWN_TYPES.Has(attr_type) {
                            cache[attr_type] = child
                        }
                    }
                }
            }
        }
    }
}

func get_counters() {
    counters := make([]counter, 0)

    request := node.NewXmlS("perf-object-counter-list-info")

    request.NewChildS("objectname", args.Object)

    connection.BuildRequest(request)

    results, err := connection.Invoke()
    if err != nil {
        fmt.Println(err)
        return
    }

    counters_elem := results.GetChildS("counters")
    if counters_elem == nil {
        fmt.Println("no counters in response")
        return
    }

    for _, elem := range counters_elem.GetChildren() {

        name := elem.GetChildContentS("name")
        c := counter{name: name}

        if elem.GetChildContentS("is-deprecated") == "true" {
            c.deprecated = true
            c.replacement = elem.GetChildContentS("replaced-by")
        }

        if elem.GetChildContentS("is-key") == "true" {
            c.key = true
        }

        c.info = elem.GetChildContentS("desc")
        c.base = elem.GetChildContentS("base-counter")
        c.unit = elem.GetChildContentS("unit")
        c.properties = strings.Split(elem.GetChildContentS("properties"), ",")
        c.privilege = elem.GetChildContentS("privelege-level")

        if elem.GetChildContentS("type") == "" {
            c.scalar = true
        } else {
            c.scalar = false

            //elem.Print(0)

            if labels := elem.GetChildS("labels"); labels != nil {

                label_elems := labels.GetChildren()

                if len(label_elems) > 0 {
                    c.labels1 = strings.Split(node.DecodeHtml(label_elems[0].GetContentS()), ",")

                    if len(label_elems) > 1 {
                        c.labels2 = strings.Split(node.DecodeHtml(label_elems[1].GetContentS()), ",")
                    }
                }
            }
        }
        counters = append(counters, c)
    }

    if len(counters) > 0 {
        counters[0].print_header()
    }

    for _, c := range counters {
        c.print()
    }

    fmt.Println()
    fmt.Printf("showed metadata of %d counters\n", len(counters))
}

func get_counter() {

}
