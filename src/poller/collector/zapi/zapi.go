package zapi

import (
    "fmt"
    "errors"
    "strings"
    "strconv"
    "sync"
    client "poller/apis/zapi"
    "poller/structs/matrix"
    "poller/structs/opts"
    "poller/yaml"
    "poller/xml"
    "poller/share"
    "poller/share/logger"
    "poller/exporter"
    "poller/schedule"
)

var Log *logger.Logger = logger.New(1, "")

type Zapi struct {
    Class string
    Name string
    Params *yaml.Node
    Args *opts.Opts
    Exporters []exporter.Exporter
    Schedule *schedule.Schedule
    Client client.Client
    System client.SystemInfo
    TemplateFn string
    Data *matrix.Matrix
    InstanceKeyPrefix []string
}

func (c *Zapi) GetName() string {
    return c.Name
}

func (c *Zapi) GetClass() string {
    return c.Class
}

func (c *Zapi) GetExporterNames() []string {
    var names []string
    e := c.Params.GetChild("exporters")
    if e != nil {
        names = e.Values
    }
    return names
}

func (c *Zapi) AddExporter(e exporter.Exporter) {
    c.Exporters = append(c.Exporters, e)
}

func New(class string, params *yaml.Node, options *opts.Opts) []*Zapi {
    var subcollectors []*Zapi
    Log = logger.New(options.LogLevel, class)
    
    connection, err := client.New(params)
    if err != nil {
        Log.Error("connecting: %v", err)
        return subcollectors
    }

    system_info, err := connection.GetSystemInfo()
    if err != nil {
        Log.Error("system info: %v", err)
        return subcollectors
    }
    Log.Info("Connected to: %s", system_info.String())

    template, err := ImportDefaultTemplate(class, options.Path)
    if err != nil {
        Log.Error("load template: %v", err)
        return subcollectors
    }

    objects := template.PopChild("objects")
    if objects == nil {
        Log.Error("no objects in template")
        return subcollectors
    }

    params.Union(template, false)

    for _, object := range objects.GetChildren() {
        c := Zapi{ 
            Class: class, 
            Name: object.Name, 
            Params: params.Copy(), 
            System: system_info,
            TemplateFn: object.Value,
        }
        Log.Debug("Initialized subcollector [%s:%s]", c.Class, c.Name)
        subcollectors = append(subcollectors, &c)
    }
    return subcollectors
}

func (c *Zapi) Init() error {

    var err error

    c.Client, err = client.New(c.Params)
    if err != nil {
        Log.Error("Error connecting: %s", err)
        return err
    }

    template, err := LoadSubTemplate(c.Args.Path, "default", c.TemplateFn, c.Class, c.System.Version)
    if err != nil {
        Log.Error("Error importing subtemplate: %s", err)
        return err
    }
 
    c.Params.Union(template, false)

    object := c.Params.GetChildValue("object")
    if object == "" {
        Log.Warn("Missing object in template")
    }
    counters := c.Params.GetChild("counters")
    if counters == nil {
        Log.Warn("Missing counters in template")
    }
    if object == "" || counters == nil {
        return errors.New("missing parameters")
    }

    c.Data = matrix.New(object, c.Class, "", c.Params.GetChild("export_options"))

    Log.Debug("Parsing counters: %d values, %d children", len(counters.Values), len(counters.Children))
    empty := make([]string, 0)
    ParseCounters(c.Data, counters, empty)
    Log.Debug("Built counter cache with %d Metrics and %d Labels", c.Data.MetricsIndex+1, len(c.Data.Instances))

    c.InstanceKeyPrefix = ParseKeyPrefix(c.Data.GetInstanceKeys())
    Log.Debug("Parsed Instance Key Prefix: %v", c.InstanceKeyPrefix)

    return nil

}

func (c *Zapi) Start(wg *sync.WaitGroup) {

    var err error

    defer wg.Done()

    if err = c.PollInstance(); err != nil {
        return
    }

    for {
        c.Schedule.Start()
        Log.Debug("Starting DataPoll")

        data, err := c.PollData()

        if err != nil {
            Log.Warn("DataPoll failed: %v", err)
        } else {
            for _, e := range c.Exporters {
                
                err = e.Export(data)

                if err != nil {
                    Log.Warn("Failed to export to [%s]", e.GetName())
                } else {
                    Log.Debug("Exported successfully to [%s]", e.GetName())
                }
            }
        }

        d := c.Schedule.Pause()
        if d < 0 {
            Log.Warn("Lagging behind schedule [%s]", d.String())
        }
    }
}

func (c *Zapi) Poll() error {
    _, err := c.PollData()
    return err
}

func (c *Zapi) PollInstance() error {
    var err error
    var root *xml.Node
    var instances []xml.Node
    var old_count int
    var keys []string
    var keypaths [][]string
    var found bool

    Log.Debug("Collector starting InstancePoll session....")

    c.Client.BuildRequest(xml.New(c.Params.GetChildValue("query")))
    root, err = c.Client.InvokeRequest()

    if err != nil {
        Log.Error("InstancePoll: client request failed: %s", err)
        return err
    }

    old_count = len(c.Data.Instances)
    c.Data.ResetInstances()

    instances = xml.SearchByPath(root, c.InstanceKeyPrefix)
    Log.Debug("Fetched %d instances!!!!", len(instances))
    keypaths = c.Data.GetInstanceKeys()

    fmt.Printf("keys=%v keypaths=%v found=%v\n", keys, keypaths, found)

    count := 0

    for _, instance := range instances {
        //c.Log.Printf("Handling instance element <%v> [%s]", &instance, instance.GetName())
        keys, found = xml.SearchByNames(&instance, c.InstanceKeyPrefix, keypaths)
        Log.Debug("Fetched instance keys (%v): %s", keypaths, strings.Join(keys, "."))

        if !found {
            Log.Debug("Skipping instance, keys not found")
        } else {
            _, err = c.Data.AddInstance(strings.Join(keys, "."))
            if err != nil {
                Log.Error("Error adding instance: %s", err)
            } else {
                Log.Debug("Added new Instance to cache [%s]", strings.Join(keys, "."))
            }
        }
        //xmltree.PrintTree(instance, 0)
        //break
        count += 1
    }

    //xmltree.PrintTree(root, 0)

    c.Data.PrintInstances()
    Log.Info("InstancePoll complete: added %d (or %d?) new instances (old cache had %d) (new cache: %d)", len(c.Data.Instances), count, old_count, len(c.Data.Instances))
    return nil
}

func (c *Zapi) PollData() (*matrix.Matrix, error) {
    var err error
    var query string
    var node *xml.Node
    var fetch func(*matrix.Instance, xml.Node, []string)
    var count, skipped int

    count = 0
    skipped = 0

    fetch = func(instance *matrix.Instance, node xml.Node, path []string) {
        newpath := append(path, node.GetName())
        key := strings.Join(newpath, ".")
        metric, found := c.Data.GetMetric(key)
        content, has := node.GetContent()

        if has {
            if found {
                if float, err := strconv.ParseFloat(string(content), 64); err != nil {
                    Log.Warn("%sSkipping metric [%s]: failed to parse [%s] float%s", share.Red, key, content, share.End)
                    skipped += 1
                } else {
                    c.Data.SetValue(metric, instance, float)
                    Log.Trace("%sMetric [%s] - Set Value [%f]%s", share.Green, key, float, share.End)
                    count += 1
                }
            } else if label, found := c.Data.GetLabel(key); found {
                c.Data.SetInstanceLabel(instance, label, string(content))
                Log.Trace("%sMetric [%s] (%s) Set Value [%s] as Instance Label%s", share.Yellow, label, key, content, share.End)
                count += 1
            } else {
                Log.Trace("%sSkipped [%s]: not found in metric or label cache%s", share.Blue, key, share.End)
                skipped += 1
            }
        } else {
            Log.Trace("Skipping metric [%s] with no value", key)
            skipped += 1
        }

        children := node.GetChildren()
        for _, child := range children {
            fetch(instance, child, newpath)
        }
    }

    Log.Debug("Starting data poll session: %s", c.System.String())

    err = c.Data.InitData()
    if err != nil {
        return nil, err
    }

    query = c.Params.GetChildValue("query")
    if query == "" { panic("missing query in template") }

    c.Client.BuildRequest(xml.New(query))

    node, err = c.Client.InvokeRequest()

    if err != nil {
        Log.Debug("Request for [%s] failed: %s", query, err)
        return nil, err
    }

    instances := xml.SearchByPath(node, c.InstanceKeyPrefix)
    Log.Debug("Fetched %d instance elements", len(instances))

    for _, instance := range instances {
        //c.Log.Printf("Handling instance element <%v> [%s]", &instance, instance.GetName())
        keys, found := xml.SearchByNames(&instance, c.InstanceKeyPrefix, c.Data.GetInstanceKeys())
        Log.Debug("Fetched instance keys: %s", strings.Join(keys, "."))

        if !found {
            Log.Debug("Skipping instance: no keys fetched")
            continue
        }

        instanceObj, found := c.Data.GetInstance(strings.Join(keys, "."))

        if !found {
            Log.Debug("Skipping instance [%s]: not found in cache", strings.Join(keys, "."))
            continue
        }
        path := make([]string, 0)
        //copy(path, c.InstanceKeyPrefix)
        fetch(instanceObj, instance, path)
    }
    //xmltree.PrintTree(node, 0)

    return c.Data, nil
}
