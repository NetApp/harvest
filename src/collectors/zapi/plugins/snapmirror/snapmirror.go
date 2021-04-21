package main

import (
	"strings"
	"goharvest2/poller/plugin"
	"goharvest2/share/dict"
	"goharvest2/share/logger"
	"goharvest2/share/matrix"
	"goharvest2/share/tree/node"
	"goharvest2/api/ontapi/zapi"
)

type SnapMirror struct {
	*plugin.AbstractPlugin
	connection        *zapi.Client
	node_cache        *dict.Dict
	dest_limit_cache  *dict.Dict
	src_limit_cache   *dict.Dict
	node_upd_counter  int
	limit_upd_counter int
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &SnapMirror{AbstractPlugin: p}
}

func (my *SnapMirror) Init() error {

	var err error

	if err = my.InitAbc(); err != nil {
		return err
	}

	if my.connection, err = zapi.New(my.ParentParams); err != nil {
		logger.Error(my.Prefix, "connecting: %v", err)
		return err
	}

	if _, err = my.connection.GetSystem(); err != nil {
		return err
	}

	my.node_upd_counter = 0
	my.limit_upd_counter = 0

	my.node_cache = dict.New()
	my.dest_limit_cache = dict.New()
	my.src_limit_cache = dict.New()

	logger.Debug(my.Prefix, "plugin initialized")
	return nil
}

func (my *SnapMirror) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	// update caches every so while
	if my.node_upd_counter == 0 && my.connection.IsClustered() {
		if err := my.update_node_cache(); err != nil {
			return nil, err
		}
		logger.Debug(my.Prefix, "updated node cache")
	} else if my.node_upd_counter > 10 {
		my.node_upd_counter = 0
	} else {
		my.node_upd_counter += 1
	}

	if my.limit_upd_counter == 0 {
		if err := my.update_limit_cache(); err != nil {
			return nil, err
		}
		logger.Debug(my.Prefix, "updated limit cache")
	} else if my.limit_upd_counter > 100 {
		my.limit_upd_counter = 0
	} else {
		my.limit_upd_counter += 1
	}

	dest_upd_count := 0
	src_upd_count := 0
	limit_upd_count := 0

	for _, instance := range data.GetInstances() {

		if my.connection.IsClustered() {
			// check instances where destination node is missing
			if instance.GetLabel("destination_node") == "" {

				key := instance.GetLabel("destination_vserver") + "." + instance.GetLabel("destination_volume")
				if node, has := my.node_cache.GetHas(key); has {
					instance.SetLabel("destination_node", node)
					dest_upd_count += 1
				}
			}

			// check instances where source node is missing
			if instance.GetLabel("source_node") == "" {

				key := instance.GetLabel("source_vserver") + "." + instance.GetLabel("source_volume")
				if node, has := my.node_cache.GetHas(key); has {
					instance.SetLabel("source_node", node)
					src_upd_count += 1
				}
			}
		} else {
			// 7 Mode
			// source / destination nodes can be something like:
			//		tobago-1:vol_4kb_neu
			//      tobago-1:D
			if src := instance.GetLabel("source_node"); src != "" {
				if x := strings.Split(src, ":"); len(x) == 2 {
					instance.SetLabel("source_node", x[0])
					if len(x[1]) != 1 {
						instance.SetLabel("source_volume", x[1])
						src_upd_count += 1
					}
				} else {
					break
				}
			}

			if dest := instance.GetLabel("destination_node"); dest != "" {
				if x := strings.Split(dest, ":"); len(x) == 2 {
					instance.SetLabel("destination_node", x[0])
					if len(x[1]) != 1 {
						instance.SetLabel("destination_volume", x[1])
						dest_upd_count += 1
					}
				} else {
					break
				}
			}
		}

		// check if destination node limit is missing
		if instance.GetLabel("destination_node_limit") == "" {

			if limit, has := my.dest_limit_cache.GetHas(instance.GetLabel("destination_node")); has {
				instance.SetLabel("destination_node_limit", limit)
				limit_upd_count += 1
			}
		}

		// check if destination node limit is missing
		if instance.GetLabel("source_node_limit") == "" {

			if limit, has := my.src_limit_cache.GetHas(instance.GetLabel("source_node")); has {
				instance.SetLabel("source_node_limit", limit)
			}
		}
	}

	logger.Debug(my.Prefix, "updated %d destination and %d source nodes, %d node limits", dest_upd_count, src_upd_count, limit_upd_count)

	return nil, nil
}

func (my *SnapMirror) update_node_cache() error {

	var (
		request, resp *node.Node
		err           error
	)

	count := 0

	request = node.NewXmlS("perf-object-get-instances")
	request.NewChildS("objectname", "volume")
	//request.CreateChild("max-records", my.batch_size)

	request_instances := request.NewChildS("instances", "")
	request_instances.NewChildS("instance", "*")

	request_counters := request.NewChildS("counters", "")
	request_counters.NewChildS("counter", "node_name")
	request_counters.NewChildS("counter", "vserver_name")

	if resp, err = my.connection.InvokeRequest(request); err != nil {
		return err
	}

	if instances := resp.GetChildS("instances"); instances != nil {
		for _, i := range instances.GetChildren() {
			vol := i.GetChildContentS("name")
			svm := i.GetChildContentS("vserver_name")
			node := i.GetChildContentS("node_name")

			my.node_cache.Set(svm+"."+vol, node)
			count += 1
		}
	}

	logger.Debug(my.Prefix, "updated node cache for %d volumes", count)
	return nil
}

func (my *SnapMirror) update_limit_cache() error {
	request := node.NewXmlS("perf-object-get-instances")
	request.NewChildS("objectname", "smc_em")

	req_i := request.NewChildS("instances", "")
	req_i.NewChildS("instance", "*")

	req_c := request.NewChildS("counters", "")
	req_c.NewChildS("counter", "node_name")
	req_c.NewChildS("counter", "dest_meter_count")
	req_c.NewChildS("counter", "src_meter_count")

	if err := my.connection.BuildRequest(request); err != nil {
		return err
	}

	resp, err := my.connection.Invoke()
	if err != nil {
		return err
	}

	count := 0

	if instances := resp.GetChildS("instances"); instances != nil {
		for _, i := range instances.GetChildren() {
			node := i.GetChildContentS("node_name")
			dest_limit := i.GetChildContentS("dest_meter_count")
			src_limit := i.GetChildContentS("src_meter_count")

			my.dest_limit_cache.Set(node, dest_limit)
			my.src_limit_cache.Set(node, src_limit)
			count += 1
		}
	}
	logger.Debug(my.Prefix, "updated limit cache for %d nodes", count)
	return nil

}
