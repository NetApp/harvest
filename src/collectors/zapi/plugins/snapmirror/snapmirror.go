
package main

import (
	"goharvest2/poller/collector/plugin"
    "goharvest2/poller/struct/matrix"
	"goharvest2/poller/struct/options"
	"goharvest2/poller/struct/xml"
	"goharvest2/poller/struct/yaml"
	"goharvest2/poller/struct/dict"
	"goharvest2/poller/util/logger"
	//"goharvest2/poller/errors"

    client "goharvest2/poller/api/zapi"
)

var Log *logger.Logger = logger.New(1, "")

type SnapMirror struct {
	*plugin.AbstractPlugin
	connection *client.Client
	node_cache *dict.Dict
	dest_limit_cache *dict.Dict
	src_limit_cache *dict.Dict
	batch_size string
	node_cache_counter int
	limit_cache_counter int
}

func New(parent_name string, options *options.Options, params *yaml.Node, pparams *yaml.Node) plugin.Plugin {
	p := plugin.New(parent_name, options, params, pparams)
	return &SnapMirror{AbstractPlugin: p}
}


func (p *SnapMirror) Init() error {
	
	var err error
	
	if err = p.InitAbc(); err != nil {
		return err
	}

	Log = logger.New(p.Options.LogLevel, "Plugin:"+p.Name)

    if p.connection, err = client.New(p.ParentParams); err != nil {
        Log.Error("connecting: %v", err)
		return err
	}

	if p.batch_size = p.ParentParams.GetChildValue("batch_size"); p.batch_size == "" {
		p.batch_size = "500"
	}

	p.node_cache_counter = 0
	p.limit_cache_counter = 0

	p.node_cache = dict.New()
	p.dest_limit_cache = dict.New()
	p.src_limit_cache = dict.New()

	Log.Debug("plugin initialized")
	return nil
}


func (p *SnapMirror) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	// helps us to update our caches every so while
	p.node_cache_counter += 1
	p.limit_cache_counter += 1

	if p.node_cache_counter > 10 {
		p.node_cache_counter = 0
	}
	if p.limit_cache_counter > 100 {
		p.limit_cache_counter = 0
	}

	if p.node_cache_counter == 1 {
		if err := p.update_node_cache(); err != nil {
			return nil, err
		}
		Log.Debug("updated node cache")
	}

	if p.limit_cache_counter == 1 {
		if err := p.update_limit_cache(); err != nil {
			return nil, err
		}
		Log.Debug("updated limit cache")
	}	

	for _, instance := range data.GetInstances() {

		// check instances where destination node is missing
		if instance.Labels.Get("destination_node") == "" {
			
			key := instance.Labels.Get("destination_vserver") + "." + instance.Labels.Get("destination_volume")
			if node, has := p.node_cache.GetHas(key); has {
				instance.Labels.Set("destination_node", node)
			}
		}

		// check instances where source node is missing
		if instance.Labels.Get("source_node") == "" {
			
			key := instance.Labels.Get("source_vserver") + "." + instance.Labels.Get("source_volume")
			if node, has := p.node_cache.GetHas(key); has {
				instance.Labels.Set("source_node", node)
			}
		}

		// check if destination node limit is missing
		if instance.Labels.Get("destination_node_limit") == "" {
			
			if limit, has := p.dest_limit_cache.GetHas(instance.Labels.Get("destination_node")); has {
				instance.Labels.Set("destination_node_limit", limit)
			}
		}

		// check if destination node limit is missing
		if instance.Labels.Get("source_node_limit") == "" {
			
			if limit, has := p.src_limit_cache.GetHas(instance.Labels.Get("source_node")); has {
				instance.Labels.Set("source_node_limit", limit)
			}
		}
	}

	return nil, nil
}


func (p *SnapMirror) update_node_cache() error {

	count := 0	

	request := xml.New("perf-object-get-instances")
	request.CreateChild("objectname", "volume")
	request.CreateChild("max-records", p.batch_size)

	req_i := xml.New("instances")
	req_i.CreateChild("instance", "*")
	request.AddChild(req_i)

	req_c := xml.New("counters")
	req_c.CreateChild("counter", "node_name")
	req_c.CreateChild("counter", "vserver_name")
	request.AddChild(req_c)

	next_tag := "init"

	for next_tag != "" {

		if next_tag != "init" {
			request.PopChild("tag")
			request.CreateChild("tag", next_tag)
		}

		if err := p.connection.BuildRequest(request); err != nil {
			return err
		}

		resp, err := p.connection.Invoke()
		if err != nil {
			return err
		}

		next_tag_tmp := resp.GetChildContentS("next-tag")
		if next_tag_tmp == next_tag {
			Log.Warn("invalid [next-tag] (ZAPI bug)")
			break
		}
		next_tag = next_tag_tmp

		if instances, has := resp.GetChild("instances"); has {
			for _, i := range instances.GetChildren() {
				vol := i.GetChildContentS("name")
				svm := i.GetChildContentS("vserver_name")
				node := i.GetChildContentS("node_name")

				p.node_cache.Set(svm+"."+vol, node)
				count += 1
			}
		}
	}

	Log.Debug("updated node cache for %d volumes", count)
	return nil	
}


func (p *SnapMirror) update_limit_cache() error {
	request := xml.New("perf-object-get-instances")
	request.CreateChild("objectname", "smc_em")

	req_i := xml.New("instances")
	req_i.CreateChild("instance", "*")
	request.AddChild(req_i)

	req_c := xml.New("counters")
	req_c.CreateChild("counter", "node_name")
	req_c.CreateChild("counter", "dest_meter_count")
	req_c.CreateChild("counter", "src_meter_count")
	request.AddChild(req_c)

	if err := p.connection.BuildRequest(request); err != nil {
		return err
	}

	resp, err := p.connection.Invoke()
	if err != nil {
		return err
	}

	count := 0

	if instances, has := resp.GetChild("instances"); has {
		for _, i := range instances.GetChildren() {
			node := i.GetChildContentS("node_name")
			dest_limit := i.GetChildContentS("dest_meter_count")
			src_limit := i.GetChildContentS("src_meter_count")

			p.dest_limit_cache.Set(node, dest_limit)
			p.src_limit_cache.Set(node, src_limit)
			count += 1
		}
	}
	Log.Debug("updated limit cache for %d nodes", count)
	return nil

}