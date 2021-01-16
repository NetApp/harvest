package collector

import (
    "fmt"
    "log"
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
    "local.host/share"
)

type Collector struct {
    Class string
    Name string
    Client client.Client
    System SystemInfo
    Data *matrix.Matrix
    Template *template.Element
    InstanceKeyPrefix []string
    Log *log.Logger
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

func (c *Collector) Init(p params.Params, t *template.Element) error {

    var err error

    c.Log = log.New(log.Writer(), fmt.Sprintf("[%-25s]: ", c.Class + ":" + c.Name), log.Flags())
    c.Log.Printf("Opened logger, initializing collector...")

    t.PrintTree(0)

    fmt.Printf("t = %v (%T)\n", t, t)
    fmt.Printf("&t = %v (%T)\n", &t, &t)
    fmt.Printf("*t = %v (%T)\n\n", *t, *t)

    c.Client, err = client.New(p)
    if err != nil {
        c.Log.Printf("Error connecting: %s", err)
        return err
    }

    c.System, err = c.GetSystemInfo()
    if err != nil {
        c.Log.Printf("Error fetching system info: %s", err)
        return err
    }

    c.Template, err = c.LoadSubtemplate(p.Path, p.Template, p.Subtemplate, c.Class, c.System.Version)
    if err != nil {
        c.Log.Printf("Error importing subtemplate: %s", err)
        return err
    }

    c.Template.PrintTree(0)
    //p := c.Template
    //c.Log.Printf("[Init] Address of pointer: %v (%v). Address of value: (%v)", c.Template, p, &p)
    //subtemplate.MergeFrom(template)

    c.Data = matrix.New(c.Template.GetChildValue("object"))

    counters := c.Template.GetChild("counters")
    if counters == nil {
        c.Log.Printf("Error: subtemplate has no counters sections")
    } else {
        c.Log.Printf("Parsing subtemplate counter section: %d values, %d children", len(counters.Values()), len(counters.Children()))
        empty := make([]string, 0)
        c.ParseCounters(c.Data, counters, empty)
        c.Log.Printf("Built counter cache with %d Metrics and %d Labels", c.Data.MetricsIndex+1, len(c.Data.Instances))

        c.InstanceKeyPrefix = ParseKeyPrefix(c.Data.GetInstanceKeys())
        c.Log.Printf("Parsed Instance Key Prefix: %v", c.InstanceKeyPrefix)

        c.Log.Printf(fmt.Sprintf("Start-up success! Connected to: %s", c.System.ToString()))
    }

    return err
}

func ParseKeyPrefix(keys [][]string) []string {
    var prefix []string
    var i, n int
    n = share.MinLen(keys)-1
    for i=0; i<n; i+=1 {
        if share.AllSame(keys, i) {
            prefix = append(prefix, keys[0][i])
        } else {
            break
        }
    }
    return prefix
}

func (c *Collector) PollInstance() error {
    var err error
    var root *xmltree.Node
    var instances []xmltree.Node
    var old_count int
    var keys []string
    var keypaths [][]string
    var found bool

    c.Log.Printf("Collector starting InstancePoll session....")

    c.Client.BuildRequest(xmltree.New(c.Template.GetChildValue("query")))
    root, err = c.Client.InvokeRequest()

    if err != nil {
        c.Log.Printf("InstancePoll: client request failed: %s", err)
        return err
    }

    old_count = len(c.Data.GetInstances())
    c.Data.ResetInstances()

    instances = xmltree.SearchByPath(root, c.InstanceKeyPrefix)
    c.Log.Printf("Fetched %d instances!!!!", len(instances))
    keypaths = c.Data.GetInstanceKeys()

    fmt.Printf("keys=%v keypaths=%v found=%v\n", keys, keypaths, found)

    count := 0

    for _, instance := range instances {
        //c.Log.Printf("Handling instance element <%v> [%s]", &instance, instance.GetName())
        keys, found = xmltree.SearchByNames(&instance, c.InstanceKeyPrefix, keypaths)
        c.Log.Printf("Fetched instance keys (%v): %s", keypaths, strings.Join(keys, "."))

        if !found {
            c.Log.Printf("Skipping instance, keys not found")
        } else {
            _, err = c.Data.AddInstance(strings.Join(keys, "."))
            if err != nil {
                c.Log.Printf("Error adding instance: %s", err)
            } else {
                c.Log.Printf("Added new Instance to cache [%s]", strings.Join(keys, "."))
            }
        }
        //xmltree.PrintTree(instance, 0)
        //break
        count += 1
    }

    //xmltree.PrintTree(root, 0)

    c.Data.PrintInstances()
    c.Log.Printf("InstancePoll complete: added %d (or %d?) new instances (old cache had %d) (new cache: %d)", len(c.Data.GetInstances()), count, old_count, len(c.Data.Instances))
    return nil
}

func (c *Collector) PollData() (*matrix.Matrix, error) {
    var err error
    var query string
    var node *xmltree.Node
    var fetch func(*matrix.Instance, xmltree.Node, []string)
    var count, skipped int

    count = 0
    skipped = 0

    fetch = func(instance *matrix.Instance, node xmltree.Node, path []string) {
        newpath := append(path, node.GetName())
        key := strings.Join(newpath, ".")
        metric, found := c.Data.GetMetric(key)
        content, has := node.GetContent()

        if has {
            if found {
                if float, err := strconv.ParseFloat(string(content), 64); err != nil {
                    c.Log.Printf("%sSkipping metric [%s]: failed to parse [%s] float%s", share.Red, key, content, share.End)
                    skipped += 1
                } else {
                    c.Data.SetValue(metric, instance, float)
                    c.Log.Printf("%sMetric [%s] - Set Value [%f]%s", share.Green, key, float, share.End)
                    count += 1
                }
            } else if label, found := c.Data.GetLabel(key); found {
                c.Data.SetInstanceLabel(instance, label, string(content))
                c.Log.Printf("%sMetric [%s] (%s) Set Value [%s] as Instance Label%s", share.Yellow, label, key, content, share.End)
                count += 1
            } else {
                c.Log.Printf("%sSkipped [%s]: not found in metric or label cache%s", share.Blue, key, share.End)
                skipped += 1
            }
        } else {
            c.Log.Printf("Skipping metric [%s] with no value", key)
            skipped += 1
        }

        children := node.GetChildren()
        for _, child := range children {
            fetch(instance, child, newpath)
        }
    }

    fmt.Printf("\n\n")
    c.Log.Printf("Starting data poll session: %s", c.System.ToString())

    err = c.Data.InitData()
    if err != nil {
        return nil, err
    }

    query = c.Template.GetChildValue("query")
    if query == "" { panic("missing query in template") }

    c.Client.BuildRequest(xmltree.New(query))

    node, err = c.Client.InvokeRequest()

    if err != nil {
        c.Log.Printf("Request for [%s] failed: %s", query, err)
        return nil, err
    }

    instances := xmltree.SearchByPath(node, c.InstanceKeyPrefix)
    c.Log.Printf("Fetched %d instance elements", len(instances))

    for _, instance := range instances {
        //c.Log.Printf("Handling instance element <%v> [%s]", &instance, instance.GetName())
        keys, found := xmltree.SearchByNames(&instance, c.InstanceKeyPrefix, c.Data.GetInstanceKeys())
        c.Log.Printf("Fetched instance keys: %s", strings.Join(keys, "."))

        if !found {
            c.Log.Printf("Skipping instance: no keys fetched")
            continue
        }

        instanceObj, found := c.Data.GetInstance(strings.Join(keys, "."))

        if !found {
            c.Log.Printf("Skipping instance [%s]: not found in cache", strings.Join(keys, "."))
            continue
        }
        path := make([]string, 0)
        //copy(path, c.InstanceKeyPrefix)
        fetch(instanceObj, instance, path)
    }
    //xmltree.PrintTree(node, 0)

    return c.Data, nil
}

func (c *Collector) LoadSubtemplate(path, dir, filename, collector string, version [3]int) (*template.Element, error) {

    var err error
    var selected_version string
    var subtemplate *template.Element

    path_prefix := filepath.Join(path, "var/", strings.ToLower(collector), dir)
    c.Log.Printf("Looking for best-fitting template in [%s]", path_prefix)

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
        c.Log.Printf("Selected best-fitting subtemplate [%s]", path)
        subtemplate, err = template.New(path)
    }
    return subtemplate, err
}


func (c *Collector) ParseCounters(data *matrix.Matrix, elem *template.Element, path []string) {
    new_path := append(path, elem.Name())
    c.Log.Printf("%v Parsing [%s] [%s] with %d values and %d children", new_path, elem.Name(), elem.Value(), len(elem.Values()), len(elem.Children()))

    if elem.Value() != "" {
        c.HandleCounter(data, new_path, elem.Value())
    }
    for _, value := range elem.Values() {
        c.HandleCounter(data, new_path, value)
    }
    for _, child := range elem.Children() {
        c.ParseCounters(data, child, new_path)
    }
}

func (c *Collector) HandleCounter(data *matrix.Matrix, path []string, value string) {
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
            c.Log.Printf("Added as Label [%s] [%s]", display, flat_path)
        if value[1] == '^' {
            data.AddInstanceKey(full_path)
            c.Log.Printf("Added as Key [%s] [%s]", display, flat_path)
        }
    } else {
        data.AddMetric(flat_path, display, true)
            c.Log.Printf("Added as Metric [%s] [%s]", display, flat_path)
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
