package main

import (
	"goharvest2/poller/collector/plugin"
	"goharvest2/share/dict"
	"goharvest2/share/logger"
	"goharvest2/share/matrix"
	"goharvest2/share/tree/node"
	"strings"

	client "goharvest2/apis/zapi"
)

type SnapMirror struct {
	*plugin.AbstractPlugin
	connection        *client.Client
	node_cache        *dict.Dict
	dest_limit_cache  *dict.Dict
	src_limit_cache   *dict.Dict
	batch_size        string
	node_upd_counter  int
	limit_upd_counter int
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &SnapMirror{AbstractPlugin: p}
}

func (p *SnapMirror) Init() error {

	var err error

	if err = p.InitAbc(); err != nil {
		return err
	}

	if p.connection, err = client.New(p.ParentParams); err != nil {
		logger.Error(p.Prefix, "connecting: %v", err)
		return err
	}

	if _, err = p.connection.GetSystem(); err != nil {
		return err
	}

	if p.batch_size = p.ParentParams.GetChildContentS("batch_size"); p.batch_size == "" {
		p.batch_size = "500"
	}

	p.node_upd_counter = 0
	p.limit_upd_counter = 0

	p.node_cache = dict.New()
	p.dest_limit_cache = dict.New()
	p.src_limit_cache = dict.New()

	logger.Debug(p.Prefix, "plugin initialized")
	return nil
}

func (p *SnapMirror) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	// update caches every so while
	if p.node_upd_counter == 0 && p.connection.IsClustered() {
		if err := p.update_node_cache(); err != nil {
			return nil, err
		}
		logger.Debug(p.Prefix, "updated node cache")
	} else if p.node_upd_counter > 10 {
		p.node_upd_counter = 0
	} else {
		p.node_upd_counter += 1
	}

	if p.limit_upd_counter == 0 {
		if err := p.update_limit_cache(); err != nil {
			return nil, err
		}
		logger.Debug(p.Prefix, "updated limit cache")
	} else if p.limit_upd_counter > 100 {
		p.limit_upd_counter = 0
	} else {
		p.limit_upd_counter += 1
	}

	dest_upd_count := 0
	src_upd_count := 0
	limit_upd_count := 0

	for _, instance := range data.GetInstances() {

		if p.connection.IsClustered() {
			// check instances where destination node is missing
			if instance.Labels.Get("destination_node") == "" {

				key := instance.Labels.Get("destination_vserver") + "." + instance.Labels.Get("destination_volume")
				if node, has := p.node_cache.GetHas(key); has {
					instance.Labels.Set("destination_node", node)
					dest_upd_count += 1
				}
			}

			// check instances where source node is missing
			if instance.Labels.Get("source_node") == "" {

				key := instance.Labels.Get("source_vserver") + "." + instance.Labels.Get("source_volume")
				if node, has := p.node_cache.GetHas(key); has {
					instance.Labels.Set("source_node", node)
					src_upd_count += 1
				}
			}
		} else {
			// 7 Mode
			// source / destination nodes can be something like:
			//		tobago-1:vol_4kb_neu
			//      tobago-1:D
			if src := instance.Labels.Get("source_node"); src != "" {
				if x := strings.Split(src, ":"); len(x) == 2 {
					instance.Labels.Set("source_node", x[0])
					if len(x[1]) != 1 {
						instance.Labels.Set("source_volume", x[1])
						src_upd_count += 1
					}
				} else {
					break
				}
			}

			if dest := instance.Labels.Get("destination_node"); dest != "" {
				if x := strings.Split(dest, ":"); len(x) == 2 {
					instance.Labels.Set("destination_node", x[0])
					if len(x[1]) != 1 {
						instance.Labels.Set("destination_volume", x[1])
						dest_upd_count += 1
					}
				} else {
					break
				}
			}
		}

		// check if destination node limit is missing
		if instance.Labels.Get("destination_node_limit") == "" {

			if limit, has := p.dest_limit_cache.GetHas(instance.Labels.Get("destination_node")); has {
				instance.Labels.Set("destination_node_limit", limit)
				limit_upd_count += 1
			}
		}

		// check if destination node limit is missing
		if instance.Labels.Get("source_node_limit") == "" {

			if limit, has := p.src_limit_cache.GetHas(instance.Labels.Get("source_node")); has {
				instance.Labels.Set("source_node_limit", limit)
			}
		}
	}

	logger.Debug(p.Prefix, "updated %d destination and %d source nodes, %d node limits", dest_upd_count, src_upd_count, limit_upd_count)

	return nil, nil
}

func (p *SnapMirror) update_node_cache() error {

	var (
		request, resp *node.Node
		next_tag      string
		err           error
	)

	count := 0

	request = node.NewXmlS("perf-object-get-instances")
	request.NewChildS("objectname", "volume")
	//request.CreateChild("max-records", p.batch_size)

	request_instances := request.NewChildS("instances", "")
	request_instances.NewChildS("instance", "*")

	request_counters := request.NewChildS("counters", "")
	request_counters.NewChildS("counter", "node_name")
	request_counters.NewChildS("counter", "vserver_name")

	if p.connection.IsClustered() {
		request.NewChildS("max-records", p.batch_size)
	}

	next_tag = "initial"

	for {

		resp, next_tag, err = p.connection.InvokeBatchRequest(request, next_tag)

		if err != nil {
			return err
		}

		if resp == nil {
			break
		}

		if instances := resp.GetChildS("instances"); instances != nil {
			for _, i := range instances.GetChildren() {
				vol := i.GetChildContentS("name")
				svm := i.GetChildContentS("vserver_name")
				node := i.GetChildContentS("node_name")

				p.node_cache.Set(svm+"."+vol, node)
				count += 1
			}
		}
	}

	logger.Debug(p.Prefix, "updated node cache for %d volumes", count)
	return nil
}

func (p *SnapMirror) update_limit_cache() error {
	request := node.NewXmlS("perf-object-get-instances")
	request.NewChildS("objectname", "smc_em")

	req_i := request.NewChildS("instances", "")
	req_i.NewChildS("instance", "*")

	req_c := request.NewChildS("counters", "")
	req_c.NewChildS("counter", "node_name")
	req_c.NewChildS("counter", "dest_meter_count")
	req_c.NewChildS("counter", "src_meter_count")

	if err := p.connection.BuildRequest(request); err != nil {
		return err
	}

	resp, err := p.connection.Invoke()
	if err != nil {
		return err
	}

	count := 0

	if instances := resp.GetChildS("instances"); instances != nil {
		for _, i := range instances.GetChildren() {
			node := i.GetChildContentS("node_name")
			dest_limit := i.GetChildContentS("dest_meter_count")
			src_limit := i.GetChildContentS("src_meter_count")

			p.dest_limit_cache.Set(node, dest_limit)
			p.src_limit_cache.Set(node, src_limit)
			count += 1
		}
	}
	logger.Debug(p.Prefix, "updated limit cache for %d nodes", count)
	return nil

}
