// Copyright NetApp Inc, 2021 All rights reserved

package qtree

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/dict"
	"github.com/netapp/harvest/v2/pkg/errors"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
	"strconv"
	"strings"
	"time"
)

const BatchSize = "500"

// Qtree plugin is needed to match qtrees with quotas.
type Qtree struct {
	*plugin.AbstractPlugin
	data           *matrix.Matrix
	instanceKeys   map[string]string
	instanceLabels map[string]*dict.Dict
	batchSize      string
	client         *zapi.Client
	query          string
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Qtree{AbstractPlugin: p}
}

func (my *Qtree) Init() error {

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

	if my.client.IsClustered() {
		my.query = "quota-report-iter"
	} else {
		my.query = "quota-report"
	}
	my.Logger.Debug().Msg("plugin connected!")

	my.data = matrix.New(my.Parent+".Qtree", "qtree", "qtree")
	my.instanceKeys = make(map[string]string)
	my.instanceLabels = make(map[string]*dict.Dict)

	exportOptions := node.NewS("export_options")
	instanceKeys := exportOptions.NewChildS("instance_keys", "")

	// apply all instance keys, instance labels from parent (qtree.yaml) to all quota metrics
	//parent instancekeys would be added in plugin metrics
	for _, parentKeys := range my.ParentParams.GetChildS("export_options").GetChildS("instance_keys").GetAllChildContentS() {
		instanceKeys.NewChildS("", parentKeys)
	}
	// parent instacelabels would be added in plugin metrics
	for _, parentLabels := range my.ParentParams.GetChildS("export_options").GetChildS("instance_labels").GetAllChildContentS() {
		instanceKeys.NewChildS("", parentLabels)
	}

	objects := my.Params.GetChildS("objects")
	if objects == nil {
		return errors.New(errors.MissingParam, "objects")
	}

	for _, obj := range objects.GetAllChildContentS() {
		metricName, display, _, _ := util.ParseMetric(obj)

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

	// setup batchSize for request
	if my.client.IsClustered() {
		if b := my.Params.GetChildContentS("batch_size"); b != "" {
			if _, err := strconv.Atoi(b); err == nil {
				my.batchSize = b
			}
		} else {
			my.batchSize = BatchSize
		}
	}

	return nil
}

func (my *Qtree) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {
	var (
		request, response *node.Node
		quotas            []*node.Node
		output            []*matrix.Matrix
		ad, pd            time.Duration // Request/API time, Parse time, Fetch time
		err               error
		numMetrics        int
	)

	apiT := 0 * time.Second
	parseT := 0 * time.Second

	// Purge and reset data
	my.data.PurgeInstances()
	my.data.Reset()

	// Set all global labels from zapi.go if already not exist
	my.data.SetGlobalLabels(data.GetGlobalLabels())

	request = node.NewXMLS(my.query)

	if my.client.IsClustered() && my.batchSize != "" {
		request.NewChildS("max-records", my.batchSize)
		// Fetching only tree quotas
		query := request.NewChildS("query", "")
		quotaQuery := query.NewChildS("quota", "")
		quotaQuery.NewChildS("quota-type", "tree")
	}

	tag := "initial"

	for {
		response, tag, ad, pd, err = my.client.InvokeBatchWithTimers(request, tag)

		if err != nil {
			return nil, err
		}

		if response == nil {
			break
		}

		apiT += ad
		parseT += pd

		if my.client.IsClustered() {
			if x := response.GetChildS("attributes-list"); x != nil {
				quotas = x.GetChildren()
			}
		} else {
			quotas = response.SearchChildren([]string{"quota"})
		}

		if len(quotas) == 0 {
			return nil, errors.New(errors.ErrNoInstance, "no quota instances found")
		}

		my.Logger.Debug().Int("quotas", len(quotas)).Msg("fetching quotas")

		for quotaIndex, quota := range quotas {
			var vserver, quotaInstanceKey string
			var qtreeInstance *matrix.Instance

			tree := quota.GetChildContentS("tree")
			volume := quota.GetChildContentS("volume")
			if my.client.IsClustered() {
				vserver = quota.GetChildContentS("vserver")
			}

			for attribute, m := range my.data.GetMetrics() {

				objectElem := quota.GetChildS(attribute)
				if objectElem == nil {
					continue
				}

				if attrValue := quota.GetChildContentS(attribute); attrValue != "" {
					if my.client.IsClustered() {
						qtreeInstance = data.GetInstance(tree + "." + volume + "." + vserver)
					} else {
						qtreeInstance = data.GetInstance(volume + "." + tree)
					}
					if qtreeInstance == nil {
						my.Logger.Warn().
							Str("tree", tree).
							Str("volume", volume).
							Str("vserver", vserver).
							Msg("No instance matching tree.volume.vserver")
						continue
					}
					if !qtreeInstance.IsExportable() {
						continue
					}
					// Ex. InstanceKey: SVMA.vol1Abc.qtree1.5.disk-limit
					if my.client.IsClustered() {
						quotaInstanceKey = vserver + "." + volume + "." + tree + "." + strconv.Itoa(quotaIndex) + "." + attribute
					} else {
						quotaInstanceKey = volume + "." + tree + "." + strconv.Itoa(quotaIndex) + "." + attribute
					}
					quotaInstance, err := my.data.NewInstance(quotaInstanceKey)

					if err != nil {
						my.Logger.Debug().Msgf("add (%s) instance: %v", attribute, err)
						return nil, err
					}

					my.Logger.Debug().Msgf("add (%s) instance: %s.%s.%s", attribute, vserver, volume, tree)

					for _, label := range my.data.GetExportOptions().GetChildS("instance_keys").GetAllChildContentS() {
						if value := qtreeInstance.GetLabel(label); value != "" {
							quotaInstance.SetLabel(label, value)
							numMetrics++
						}
					}

					// If the Qtree is the volume itself, then qtree label is empty, so copy the volume name to qtree.
					if tree == "" {
						quotaInstance.SetLabel("qtree", volume)
						numMetrics++
					}

					// populate numeric data
					if value := strings.Split(attrValue, " ")[0]; value != "" {
						// Few quota metrics would have value '-' which means unlimited (ex: disk-limit)
						if value == "-" {
							value = "0"
						}
						if err := m.SetValueString(quotaInstance, value); err != nil {
							my.Logger.Debug().Msgf("(%s) failed to parse value (%s): %v", attribute, value, err)
						} else {
							numMetrics++
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

	my.Logger.Info().
		Int("numQtrees", len(data.GetInstances())).
		Int("numQuotas", len(quotas)).
		Int("metrics", numMetrics).
		Str("apiD", apiT.Round(time.Millisecond).String()).
		Str("parseD", parseT.Round(time.Millisecond).String()).
		Str("batchSize", my.batchSize).
		Msg("Collected")
	return output, nil
}
