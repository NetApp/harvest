package collector

import (
    "fmt"
    "errors"
    "strings"
    "strconv"
    "regexp"
    "io/ioutil"
    "path/filepath"
    "local.host/client"
    "local.host/params"
    "local.host/matrix"
    "local.host/template"
    "local.host/xmltree"
)

type Collector struct {
    Class string
    Name string
    Client client.Client
    System SystemInfo
    Data *matrix.Matrix
    Template *template.Element
}

func New(class, name string) *Collector {
    var c Collector
    c = Collector{ Class : class, Name : name }
    //z.Connection = new(api.Connection)
    //c.System = new(SystemInfo)
    c.Data = new(matrix.Matrix)
    c.Template = new(template.Element)
    return &c
}

func (c *Collector) Log(format string, vars ...interface{}) {
    fmt.Printf("[%s:%s] ", c.Class, c.Name)
    fmt.Printf(format, vars...)
    fmt.Println()
}

func (c Collector) Init(p params.Params, t *template.Element) error {

    var err error

    c.Log("\nIntializing!")

    t.PrintTree(0)

    fmt.Printf("t = %v (%T)\n", t, t)
    fmt.Printf("&t = %v (%T)\n", &t, &t)
    fmt.Printf("*t = %v (%T)\n\n", *t, *t)

    c.Client, err = client.New(p)
    if err != nil {
        c.Log("Error connecting: %s", err)
        return err
    }

    c.System, err = c.GetSystemInfo()
    if err != nil {
        c.Log("Error fetching system info: %s", err)
        return err
    }

    c.Template, err = c.LoadSubtemplate(p.Path, p.Template, p.Subtemplate, c.Class, c.System.Version)
    if err != nil {
        c.Log("Error importing subtemplate: %s", err)
        return err
    }
    //p := c.Template
    //c.Log("[Init] Address of pointer: %v (%v). Address of value: (%v)", c.Template, p, &p)
    //subtemplate.MergeFrom(template)

    c.Data = matrix.New("volume")

    counters := c.Template.GetChild("counters")
    if counters == nil {
        c.Log("Error: subtemplate has no counters sections")
    } else {
        c.Log("Parsing subtemplate counter section: %d values, %d children", len(counters.Values()), len(counters.Children()))
        empty := make([]string, 0)
        c.ParseCounters(c.Data, counters, empty)
        c.Log("Built counter cache with %d Metrics and %d Labels", len(c.Data.Counters), len(c.Data.Instances))

        c.Log(fmt.Sprintf("Start-up success! Connected to: %s", c.System.ToString()))
    }

    query := c.Template.GetChildValue("query")
    c.Log("I got query: [%s] and template is [%v]", query, c.Template)
    return err
}

func (c Collector) PollData() error {
    var err error
    var query string
    var node *xmltree.Node

    c.Log("\n\nStarting data poll session: %s", c.System.ToString())

    if c.Template == nil {
        c.Log("template is [%v] and NIL!!", c.Template)
    } else {
        c.Log("template is [%v] and OK!", c.Template)
    }

    query = c.Template.GetChildValue("query")
    if query == "" { panic("missing query in template") }

    c.Client.BuildRequest(xmltree.New(query))

    node, err = c.Client.InvokeRequest()

    if err != nil {
        c.Log("Request for [%s] failed: %s", query, err)
    } else {
        xmltree.PrintTree(node, 0)
    }
    return err
}


func (c Collector) LoadSubtemplate(path, dir, filename, collector string, version [3]int) (*template.Element, error) {

    var err error
    var selected_version string
    var subtemplate *template.Element

    path_prefix := filepath.Join(path, "var/", strings.ToLower(collector), dir)
    c.Log("Looking for best-fitting template in [%s]", path_prefix)

    available := make(map[string]bool)
    files, _ := ioutil.ReadDir(path_prefix)
    for _, file := range files {
        if match, _ := regexp.MatchString(`\d+\.\d+\.\d+`, file.Name()); match == true && file.IsDir() {
            available[file.Name()] = true
        }
    }

    vers := version[0] * 100 + version[1] * 10 + version[2]
    if err != nil { return subtemplate, err }

    for max:=300; max>0 && vers>0; max-=1 {
        str := strings.Join(strings.Split(strconv.Itoa(vers), ""), ".")
        if _, exists := available[str]; exists == true {
            selected_version = str
            break
        }
        vers-= 1
    }

    if selected_version == "" {
        err = errors.New("No best-fitting subtemplate version found")
    } else {
        path := filepath.Join(path_prefix, selected_version, filename)
        c.Log("Selected best-fitting subtemplate [%s]", path)
        subtemplate, err = template.New(path)
    }
    return subtemplate, err
}


func (c Collector) ParseCounters(data *matrix.Matrix, elem *template.Element, path []string) {
    c.Log("Parsing [%s] with %d values and %d children", elem.Name(), len(elem.Values()), len(elem.Children()))

    if elem.Value() != "" {
        c.HandleCounter(data, path, elem.Value())
    }
    for _, value := range elem.Values() {
        c.HandleCounter(data, path, value)
    }
    new_path := append(path, elem.Name())
    for _, child := range elem.Children() {
        c.ParseCounters(data, child, new_path)
    }
}

func (c Collector) HandleCounter(data *matrix.Matrix, path []string, value string) {
    var name, display, flat_path string
    var split_value, full_path []string

    split_value = strings.Split(value, "=>")
    if len(split_value) == 1 {
        name = value
    } else {
        name = split_value[0]
        display = strings.TrimLeft(split_value[1], " ")
    }

    name = strings.TrimLeft(name, "^")
    name = strings.TrimRight(name, " ")

    full_path = append(path[1:], name)
    flat_path = strings.Join(full_path, ".")

    if display == "" {
        display = ParseDisplay(data.Object, full_path)
    }

    if value[0] == '^' {
        data.AddLabel(flat_path, display)
            c.Log("Added as Label [%s] [%s]", display, flat_path)
        if value[1] == '^' {
            data.AddInstanceKey(full_path)
            c.Log("Added as Key [%s] [%s]", display, flat_path)
        }
    } else {
        data.AddCounter(flat_path, display, true)
            c.Log("Added as Metric [%s] [%s]", display, flat_path)
    }
}

func ParseDisplay(obj string, path []string) string {
    var ignore = map[string]int{"attributes" : 0, "info" : 0, "list" : 0, "details" : 0}
    var added = map[string]int{}
    var words []string

    for _, attribute := range path {
        split := strings.Split(attribute, "-")
        for _, word := range split {
            if word == obj { continue }
            if _, exists := ignore[word]; exists { continue }
            if _, exists := added[word]; exists { continue }
            words = append(words, word)
            added[word] = 0
        }
    }
    return strings.Join(words, "_")
}
