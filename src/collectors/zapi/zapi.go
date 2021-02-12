package main

import (
    "strings"
    "strconv"

    "goharvest2/share/logger"
    "goharvest2/poller/struct/matrix"
    "goharvest2/poller/struct/xml"
    "goharvest2/poller/util"
    "goharvest2/poller/collector"
	"goharvest2/poller/errors"

    client "goharvest2/poller/api/zapi"
)

type Zapi struct {
    *collector.AbstractCollector
    connection *client.Client
    system *client.System
    object string
	template_fn string
	template_type string
    instanceKeyPrefix []string
}

func New(a *collector.AbstractCollector) collector.Collector {
    return &Zapi{AbstractCollector: a}
}

func (c *Zapi) Init() error {

    var err error
    if c.connection, err = client.New(c.Params); err != nil {
        return err
    }

    // @TODO handle connectivity-related errors (retry a few times)
    if c.system, err = c.connection.GetSystem(); err != nil {
        //logger.Error(c.Prefix, "system info: %v", err)
        return err
    }
    logger.Debug(c.Prefix, "Connected to: %s", c.system.String())

    template_fn := c.Params.GetChild("objects").GetChildValue(c.Object) // @TODO err handling

    template, err := collector.ImportSubTemplate(c.Options.Path, "default", template_fn, c.Name, c.system.Version)
    if err != nil {
        logger.Error(c.Prefix, "Error importing subtemplate: %s", err)
        return err
    }
    c.Params.Union(template, false)

    // object name from subtemplate
    if c.object = c.Params.GetChildValue("object"); c.object == "" {
        return errors.New(errors.MISSING_PARAM, "object")
    }
 
    // Invoke generic initializer
    // this will load Schedule, initialize Data and Metadata
    if err := collector.Init(c); err != nil {
        return err
    }

    // Add system (cluster) name 
    c.Data.SetGlobalLabel("system", c.system.Name)

    // Initialize counter cache
    counters := c.Params.GetChild("counters")
    if counters == nil {
        return errors.New(errors.MISSING_PARAM, "counters")
    }

    //@TODO cleanup
    logger.Debug(c.Prefix, "Parsing counters: %d values, %d children", len(counters.Values), len(counters.Children))
    ParseCounters(c.Data, counters, make([]string, 0))
    logger.Debug(c.Prefix, "Built counter cache with %d Metrics and %d Labels", c.Data.MetricsIndex+1, len(c.Data.Instances))

    if len(c.Data.InstanceKeys) == 0 {
        return errors.New(errors.INVALID_PARAM, "no instance keys indicated")
    }

    c.instanceKeyPrefix = ParseShortestPath(c.Data)
    logger.Debug(c.Prefix, "Parsed Instance Keys: %v", c.Data.InstanceKeys)
    logger.Debug(c.Prefix, "Parsed Instance Key Prefix: %v", c.instanceKeyPrefix)

    return nil

}

func (c *Zapi) PollInstance() (*matrix.Matrix, error) {
    var err error
    var response *xml.Node
    var instances []*xml.Node
    var old_count int
    var keys []string
    var keypaths [][]string
    var found bool

    logger.Debug(c.Prefix, "starting instance poll")

    //@TODO next tag
    if err = c.connection.BuildRequestString(c.Params.GetChildValue("query")); err != nil {
        return nil, err
    }

    if response, err = c.connection.Invoke(); err != nil {
        return nil, err
    }

    old_count = len(c.Data.Instances)
    c.Data.ResetInstances()

    instances = xml.SearchByPath(response, c.instanceKeyPrefix)
    if len(instances) == 0 {
        return nil, errors.New(errors.ERR_NO_INSTANCE, "no instances in server response")
    }

    logger.Debug(c.Prefix, "fetching %d instances", len(instances))
    // @Cleanup
    keypaths = c.Data.GetInstanceKeys()
    logger.Debug(c.Prefix, "keys=%v keypaths=%v found=%v", keys, keypaths, found)

    count := 0

    for _, instance := range instances {
        //c.logger.Printf(c.Prefix, "Handling instance element <%v> [%s]", &instance, instance.GetName())
        keys, found = xml.SearchByNames(instance, c.instanceKeyPrefix, keypaths)
        logger.Debug(c.Prefix, "fetched instance keys (%v): %s", keypaths, strings.Join(keys, "."))

        if !found {
            logger.Debug(c.Prefix, "skipping element, no instance keys not found")
        } else {
            if _, err = c.Data.AddInstance(strings.Join(keys, ".")); err != nil {
                logger.Error(c.Prefix, err.Error())
            } else {
                logger.Debug(c.Prefix, "Added new Instance to cache [%s]", strings.Join(keys, "."))
                count += 1
            }
        }
    }

    c.Metadata.SetValueSS("count", "instance", float32(count))
    logger.Info(c.Prefix, "added %d instances to cache (old cache had %d)", count, old_count)

    if len(c.Data.Instances) == 0 {
        return nil, errors.New(errors.ERR_NO_INSTANCE, "no instances fetched")
    }

    return nil, nil
}

func (c *Zapi) PollData() (*matrix.Matrix, error) {
    var err error
    var query string
    var response *xml.Node
    var fetch func(*matrix.Instance, *xml.Node, []string)
    var count, skipped int

    count = 0
    skipped = 0

    fetch = func(instance *matrix.Instance, node *xml.Node, path []string) {
        newpath := append(path, node.GetName())
        key := strings.Join(newpath, ".")
        metric := c.Data.GetMetric(key)
        content, has := node.GetContent()

        if has {
            if metric != nil {
                if float, err := strconv.ParseFloat(string(content), 32); err != nil {
                    logger.Warn(c.Prefix, "%sSkipping metric [%s]: failed to parse [%s] float%s", util.Red, key, content, util.End)
                    skipped += 1
                } else {
                    c.Data.SetValue(metric, instance, float32(float))
                    logger.Trace(c.Prefix, "%sMetric [%s] - Set Value [%f]%s", util.Green, key, float, util.End)
                    count += 1
                }
            } else if label, found := c.Data.GetLabel(key); found {
                //c.Data.SetInstanceLabel(instance, label, string(content))
                instance.Labels.Set(label, string(content))
                logger.Trace(c.Prefix, "%sMetric [%s] (%s) Set Value [%s] as Instance Label%s", util.Yellow, label, key, content, util.End)
                count += 1
            } else {
                logger.Trace(c.Prefix, "%sSkipped [%s]: not found in metric or label cache%s", util.Blue, key, util.End)
                skipped += 1
            }
        } else {
            logger.Trace(c.Prefix, "Skipping metric [%s] with no value", key)
            skipped += 1
        }

        for _, child := range node.GetChildren() {
            fetch(instance, child, newpath)
        }
    }

    logger.Debug(c.Prefix, "starting data poll")

    if err = c.Data.InitData(); err != nil {
        return nil, err
    }

    // @todo just verify once in init
    if query = c.Params.GetChildValue("query"); query == "" {
        return nil, errors.New(errors.MISSING_PARAM, "query")
    }

    if err = c.connection.BuildRequestString(query); err != nil {
        return nil, err
    }

    if response, err = c.connection.Invoke(); err != nil {
        return nil, err
    }

    instances := xml.SearchByPath(response, c.instanceKeyPrefix)
    if len(instances) == 0 {
        return nil, errors.New(errors.ERR_NO_INSTANCE, "")
    }

    logger.Debug(c.Prefix, "Fetched %d instance elements", len(instances))

    for _, instanceElem := range instances {
        //c.logger.Printf(c.Prefix, "Handling instance element <%v> [%s]", &instance, instance.GetName())
        keys, found := xml.SearchByNames(instanceElem, c.instanceKeyPrefix, c.Data.GetInstanceKeys())
        logger.Debug(c.Prefix, "Fetched instance keys: %s", strings.Join(keys, "."))

        if !found {
            logger.Debug(c.Prefix, "Skipping instance: no keys fetched")
            continue
        }

        instance := c.Data.GetInstance(strings.Join(keys, "."))

        if instance == nil {
            logger.Debug(c.Prefix, "Skipping instance [%s]: not found in cache", strings.Join(keys, "."))
            continue
        }
        fetch(instance, instanceElem, make([]string, 0))
    }
    return c.Data, nil
}
