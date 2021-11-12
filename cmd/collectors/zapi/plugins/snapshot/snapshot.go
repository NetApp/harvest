package snapshot

import (
	"goharvest2/cmd/poller/collector"
	"goharvest2/cmd/poller/plugin"
	"goharvest2/pkg/api/ontapi/zapi"
	"goharvest2/pkg/conf"
	"goharvest2/pkg/dict"
	"goharvest2/pkg/errors"
	"goharvest2/pkg/matrix"
	"goharvest2/pkg/tree/node"
	"strconv"
	"strings"
)

const BatchSize = "500"

type Snapshot struct {
	*plugin.AbstractPlugin
	data           *matrix.Matrix
	instanceKeys   map[string]string
	instanceLabels map[string]*dict.Dict
	batchSize      string
	client         *zapi.Client
	query          string
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Snapshot{AbstractPlugin: p}
}

func (my *Snapshot) Init() error {

	var err error

	if err = my.InitAbc(); err != nil {
		return err
	}

	if my.client, err = zapi.New(conf.ZapiPoller(my.ParentParams)); err != nil {
		my.Logger.Error().Stack().Err(err).Msg("connecting")
		return err
	}

	if err = my.client.Init(5); err != nil {
		return err
	}

	my.query = "snapshot-policy-get-iter"
	my.Logger.Debug().Msg("plugin connected!")

	my.data = matrix.New(my.Parent+".Snapshot", "snapshot", "snapshot")
	my.instanceKeys = make(map[string]string)
	my.instanceLabels = make(map[string]*dict.Dict)

	exportOptions := node.NewS("export_options")
	instanceKeys := exportOptions.NewChildS("instance_keys", "")

	instanceKeys.NewChildS("", "vserver")
	instanceKeys.NewChildS("", "snapshot_policy")
	instanceKeys.NewChildS("", "comment")

	objects := my.Params.GetChildS("objects")
	if objects == nil {
		return errors.New(errors.MISSING_PARAM, "objects")
	}

	for _, obj := range objects.GetAllChildContentS() {
		metricName, display := collector.ParseMetricName(obj)

		metric, err := my.data.NewMetricFloat64(metricName)
		if err != nil {
			my.Logger.Error().Stack().Err(err).Msg("add metric")
			return err
		}

		metric.SetName(display)
		my.Logger.Debug().Msgf("added metric: (%s) [%s] %s", metricName, display, metric)
	}

	my.Logger.Debug().Msgf("added data with %d metrics", len(my.data.GetMetrics()))
	my.data.SetExportOptions(exportOptions)

	// batching the request
	if my.client.IsClustered() {
		if b := my.Params.GetChildContentS("batch_size"); b != "" {
			if _, err := strconv.Atoi(b); err == nil {
				my.batchSize = b
				my.Logger.Info().Msgf("using batch-size [%s]", my.batchSize)
			}
		} else {
			my.batchSize = BatchSize
			my.Logger.Trace().Str("BatchSize", BatchSize).Msg("Using default batch-size")
		}
	}

	return nil
}

func (my *Snapshot) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	var (
		request, result *node.Node
		policyInfos     []*node.Node
		tag             string
		err             error
	)

	var output []*matrix.Matrix

	// Purge and reset data
	my.data.PurgeInstances()
	my.data.Reset()

	// Set all global labels from zapi.go if already not exist
	my.data.SetGlobalLabels(data.GetGlobalLabels())

	request = node.NewXmlS(my.query)
	if my.client.IsClustered() && my.batchSize != "" {
		request.NewChildS("max-records", my.batchSize)
	}

	tag = "initial"

	for {
		result, tag, err = my.client.InvokeBatchRequest(request, tag)

		if err != nil {
			return nil, err
		}

		if result == nil {
			break
		}

		if x := result.GetChildS("attributes-list"); x != nil {
			policyInfos = x.GetChildren()
		}

		if len(policyInfos) == 0 {
			return nil, errors.New(errors.ERR_NO_INSTANCE, "no snapshot policy info instances found")
		}

		my.Logger.Debug().Msgf("fetching %d snapshot policy counters", len(policyInfos))

		for policyInfoIndex, policyInfo := range policyInfos {

			var comment string
			vserver := policyInfo.GetChildContentS("vserver-name")
			policyName := policyInfo.GetChildContentS("policy")
			if comment = policyInfo.GetChildContentS("comment"); comment == "" {
				comment = "No Comment"
			}

			for attribute, m := range my.data.GetMetrics() {

				objectElem := policyInfo.GetChildS(attribute)
				if objectElem == nil {
					my.Logger.Warn().Msgf("no [%s] instances on this %s.%s.%s", attribute, vserver, policyName, comment)
					continue
				}

				if attrValue := policyInfo.GetChildContentS(attribute); attrValue != "" {
					// Ex. InstanceKey: fas8040-206-21.default.Default policy with hourly, daily.12.total-schedules
					instanceKey := vserver + "." + policyName + "." + comment + "." + strconv.Itoa(policyInfoIndex) + "." + attribute
					//my.Logger.Info().Msgf(instanceKey)
					instance, err := my.data.NewInstance(instanceKey)

					if err != nil {
						my.Logger.Debug().Msgf("add (%s) instance: %v", attribute, err)
						return nil, err
					}

					my.Logger.Debug().Msgf("add (%s) instance: %s.%s.%s", attribute, vserver, policyName, comment)

					instance.SetLabel("vserver", vserver)
					instance.SetLabel("snapshot_policy", policyName)
					instance.SetLabel("comment", comment)

					// populate numeric data
					if value := strings.Split(attrValue, " ")[0]; value != "" {
						if err := m.SetValueString(instance, value); err != nil {
							my.Logger.Debug().Msgf("(%s) failed to parse value (%s): %v", attribute, value, err)
						} else {
							my.Logger.Debug().Msgf("(%s) added value (%s)", attribute, value)
						}
					}

				} else {
					my.Logger.Debug().Msgf("instance without [%s], skipping", attribute)
				}

				output = append(output, my.data)
			}
		}

	}
	return output, nil
}
