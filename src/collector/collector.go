package collector

import (
    "fmt"
    "errors"
    "strings"
    "strconv"
    "sync"
    "local.host/client"
    "local.host/matrix"
    "local.host/yaml"
    "local.host/xml"
    "local.host/share"
    "local.host/logger"
    "local.host/exporter"
    "local.host/schedule"
    "local.host/args"
)

var Log *logger.Logger = logger.New(1, "")

type Collector struct {
    Class string
    Name string
    Master bool
    Params *yaml.Node
    Args *args.Args
    SubCollectors []*Collector
    Exporters []*exporter.Exporter
    Schedule *schedule.Schedule
    Client client.Client
    System SystemInfo
    Data *matrix.Matrix
    InstanceKeyPrefix []string
}

func New(class string, params *yaml.Node, options *args.Args) *Collector {
    var c Collector
    c = Collector{ Class : class, Master : true, Params : params, Args : options }
    //z.Connection = new(api.Connection)
    //c.System = new(SystemInfo)
    //c.Data = new(matrix.Matrix)
    //c.Template = new(template.Element)

	Log = logger.New(options.LogLevel, class)
    return &c
}


func newChild(class, name string, params *yaml.Node, options *args.Args) *Collector {
    var c Collector
    c = Collector{Class: class, Name: name, Master: false, Params: params, Args: options}
    c.Schedule = schedule.New(20)

    Log = logger.New(options.LogLevel, class + ":" + name)
    return &c
}

func (c *Collector) Init() error {

    var err error

    c.Client, err = client.New(c.Params)
    if err != nil {
        Log.Error("Error connecting: %s", err)
        return err
    }

    c.System, err = c.GetSystemInfo()
    if err != nil {
        Log.Error("Error fetching system info: %s", err)
        return err
    }
    
    template, err := ImportDefaultTemplate(c.Class, c.Args.Path)
    subtemplates := template.PopChild("subtemplates")
    if subtemplates == nil {
        return errors.New("No subtemplates defined in template")
    }
    c.Params.Union(template, false)
    
    for _, subt := range subtemplates.GetChildren() {
        child := newChild(c.Class, subt.Name, c.Params.Copy(), c.Args)
        err := child.initChild(subt.Value)
        if err == nil {
            Log.Debug("Iniitialized subcollector [%s:%s]", child.Class, child.Name)
            c.SubCollectors = append(c.SubCollectors, child)
        } else {
            Log.Error("Failed initializing [%s:%s]: %v", child.Class, child.Name, err)
        }
    }

    if len(c.SubCollectors) == 0 {
        Log.Warn("Failed to initialize any subcollectors")
        return errors.New("No subcollectors")
    }

    Log.Info("Connected to: %s", c.System.ToString())
    Log.Info("Start-up success! Initialized %d subcollectors", len(c.SubCollectors))
    return nil
}


func (c *Collector) initChild(template_name string) error {

    var err error

    c.Client, err = client.New(c.Params)
    if err != nil {
        Log.Error("Error connecting: %s", err)
        return err
    }

    template, err := LoadSubTemplate(c.Args.Path, "default", template_name, c.Class, c.System.Version)
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

    c.Data = matrix.New(object)

    Log.Debug("Parsing counters: %d values, %d children", len(counters.Values), len(counters.Children))
    empty := make([]string, 0)
    ParseCounters(c.Data, counters, empty)
    Log.Debug("Built counter cache with %d Metrics and %d Labels", c.Data.MetricsIndex+1, len(c.Data.Instances))

    c.InstanceKeyPrefix = ParseKeyPrefix(c.Data.GetInstanceKeys())
    Log.Debug("Parsed Instance Key Prefix: %v", c.InstanceKeyPrefix)

    return nil

}

func (c *Collector) Start(wg *sync.WaitGroup) {

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
                
                err = e.Export(data, c.Params.GetChild("export_options"))

                if err != nil {
                    Log.Warn("Failed to export to [%s]", e.Name)
                } else {
                    Log.Debug("Exported successfully to [%s]", e.Name)
                }
            }
        }

        d := c.Schedule.Pause()
        if d < 0 {
            Log.Warn("Lagging behind schedule [%s]", d.String())
        }
    }
}

func (c *Collector) PollInstance() error {
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

    old_count = len(c.Data.GetInstances())
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
    Log.Info("InstancePoll complete: added %d (or %d?) new instances (old cache had %d) (new cache: %d)", len(c.Data.GetInstances()), count, old_count, len(c.Data.Instances))
    return nil
}

func (c *Collector) PollData() (*matrix.Matrix, error) {
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

    Log.Debug("Starting data poll session: %s", c.System.ToString())

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
