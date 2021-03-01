package zapi_collector

import (
    "strings"
    "strconv"
    "path"

    "goharvest2/share/logger"
    "goharvest2/share/tree/node"
	"goharvest2/share/errors"
    "goharvest2/share/util"
    "goharvest2/share/matrix"
    "goharvest2/poller/collector"

    client "goharvest2/apis/zapi"
)

type Zapi struct {
    *collector.AbstractCollector
    Connection *client.Client
    System *client.System
    object string
    Query string
	TemplateFn string
	TemplateType string
    instanceKeyPrefix []string
    instanceKeys [][]string
}

func New(a *collector.AbstractCollector) collector.Collector {
    return &Zapi{AbstractCollector: a}
}

func NewZapi(a *collector.AbstractCollector) *Zapi {
    return &Zapi{AbstractCollector: a}
}

func (c *Zapi) Init() error {

    // @TODO check if cert/key files exist
    if c.Params.GetChildContentS("auth_style") == "certificate_auth" {
        if c.Params.GetChildS("ssl_cert") == nil {
			cert_path := path.Join(c.Options.ConfPath, "cert", c.Options.Poller + ".pem")
            c.Params.NewChildS("ssl_cert", cert_path)
            logger.Debug(c.Prefix, "added ssl_cert path [%s]", cert_path)
        }

        if c.Params.GetChildS("ssl_key") == nil {
			key_path := path.Join(c.Options.ConfPath, "cert", c.Options.Poller + ".key")
            c.Params.NewChildS("ssl_key", key_path)
            logger.Debug(c.Prefix, "added ssl_key path [%s]", key_path)
        }
    }

    var err error
    if c.Connection, err = client.New(c.Params); err != nil {
        return err
    }

    // @TODO handle connectivity-related errors (retry a few times)
    if c.System, err = c.Connection.GetSystem(); err != nil {
        //logger.Error(c.Prefix, "system info: %v", err)
        return err
    }
    logger.Debug(c.Prefix, "Connected to: %s", c.System.String())

    c.TemplateFn = c.Params.GetChildS("objects").GetChildContentS(c.Object) // @TODO err handling

    model := "cdot"
    if !c.System.Clustered {
        model = "7mode"
    }

    template, err := c.ImportSubTemplate(model, "default", c.TemplateFn, c.System.Version)
    if err != nil {
        logger.Error(c.Prefix, "Error importing subtemplate: %s", err)
        return err
    }
    c.Params.Union(template)

    // object name from subtemplate
    if c.object = c.Params.GetChildContentS("object"); c.object == "" {
        return errors.New(errors.MISSING_PARAM, "object")
    }

    // api query literal
    if c.Query = c.Params.GetChildContentS("query"); c.Query == "" {
        return errors.New(errors.MISSING_PARAM, "query")
    }

    // Invoke generic initializer
    // this will load Schedule, initialize Data and Metadata
    if err := collector.Init(c); err != nil {
        return err
    }

    // overwrite from abstract collector
    c.Data.Object = c.object
    
    // Add system (cluster) name 
    c.Data.SetGlobalLabel("cluster", c.System.Name)

    // Initialize counter cache
    counters := c.Params.GetChildS("counters")
    if counters == nil {
        return errors.New(errors.MISSING_PARAM, "counters")
    }

    if err = c.InitCache(); err != nil {
        return err
    }

    logger.Debug(c.Prefix, "Successfully initialized")
    return nil
}

func (c *Zapi) InitCache() error {

    //@TODO cleanup
    counters := c.Params.GetChildS("counters")

    logger.Debug(c.Prefix, "Parsing counters: %d values", len(counters.GetChildren()))
    if ! LoadCounters(c.Data, counters) {
        return errors.New(errors.ERR_NO_METRIC, "failed to parse any")
    }

    logger.Debug(c.Prefix, "Loaded %d Metrics and %d Labels", c.Data.SizeMetrics(), c.Data.SizeLabels())

    if len(c.Data.InstanceKeys) == 0 {
        return errors.New(errors.INVALID_PARAM, "no instance keys indicated")
    }

    // @TODO validate
    c.instanceKeyPrefix = ParseShortestPath(c.Data)
    logger.Debug(c.Prefix, "Parsed Instance Keys: %v", c.Data.InstanceKeys)
    logger.Debug(c.Prefix, "Parsed Instance Key Prefix: %v", c.instanceKeyPrefix)
    return nil

}

func (c *Zapi) PollInstance() (*matrix.Matrix, error) {
    var err error
    var response *node.Node
    var instances []*node.Node
    var old_count int
    var keys []string
    var keypaths [][]string
    var found bool

    logger.Debug(c.Prefix, "starting instance poll")

    //@TODO next tag
    if err = c.Connection.BuildRequestString(c.Query); err != nil {
        return nil, err
    }

    if response, err = c.Connection.Invoke(); err != nil {
        return nil, err
    }

    old_count = len(c.Data.Instances)
    c.Data.ResetInstances()

    instances = response.SearchChildren(c.instanceKeyPrefix)
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
        keys, found = instance.SearchContent(c.instanceKeyPrefix, keypaths)
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

    c.Metadata.SetValueSS("count", "instance", float64(count))
    logger.Info(c.Prefix, "added %d instances to cache (old cache had %d)", count, old_count)

    if len(c.Data.Instances) == 0 {
        return nil, errors.New(errors.ERR_NO_INSTANCE, "no instances fetched")
    }

    return nil, nil
}

func (c *Zapi) PollData() (*matrix.Matrix, error) {
    var err error
    var response *node.Node
    var fetch func(*matrix.Instance, *node.Node, []string)
    var count, skipped int

    count = 0
    skipped = 0

    fetch = func(instance *matrix.Instance, node *node.Node, path []string) {
        newpath := append(path, node.GetNameS())
        key := strings.Join(newpath, ".")
        metric := c.Data.GetMetric(key)
        content := node.GetContentS()

        if content != "" {
            if metric != nil {
                if float, err := strconv.ParseFloat(string(content), 32); err != nil {
                    logger.Warn(c.Prefix, "%sSkipping metric [%s]: failed to parse [%s] float%s", util.Red, key, content, util.End)
                    skipped += 1
                } else {
                    c.Data.SetValue(metric, instance, float64(float))
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

    if err = c.Connection.BuildRequestString(c.Query); err != nil {
        return nil, err
    }

    if response, err = c.Connection.Invoke(); err != nil {
        return nil, err
    }

    instances := response.SearchChildren(c.instanceKeyPrefix)
    if len(instances) == 0 {
        return nil, errors.New(errors.ERR_NO_INSTANCE, "")
    }

    logger.Debug(c.Prefix, "Fetched %d instance elements", len(instances))

    for _, instanceElem := range instances {
        //c.logger.Printf(c.Prefix, "Handling instance element <%v> [%s]", &instance, instance.GetName())
        keys, found := instanceElem.SearchContent(c.instanceKeyPrefix, c.Data.GetInstanceKeys())
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
