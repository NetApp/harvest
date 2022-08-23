// Copyright NetApp Inc, 2021 All rights reserved

package qtree

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/dict"
	"github.com/netapp/harvest/v2/pkg/errs"
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
	quotaType      []string
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
	exportOptions.NewChildS("include_all_labels", "true")

	objects := my.Params.GetChildS("objects")
	if objects == nil {
		return errs.New(errs.ErrMissingParam, "objects")
	}

	quotaType := my.Params.GetChildS("quotaType")
	if quotaType != nil {
		my.quotaType = []string{}
		my.quotaType = append(my.quotaType, quotaType.GetAllChildContentS()...)
	} else {
		my.quotaType = []string{"user", "group", "tree"}
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
		// add quota filter
		query := request.NewChildS("query", "")
		quotaQuery := query.NewChildS("quota", "")
		quotaQuery.NewChildS("quota-type", strings.Join(my.quotaType[:], "|"))
	}

	tag := "initial"
	quotaIndex := 0
	var cluster string
	if data.GetGlobalLabels().Has("cluster") {
		cluster = data.GetGlobalLabels().Get("cluster")
	}

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
			my.Logger.Debug().Msg("no quota instances found")
			return nil, nil
		}

		my.Logger.Debug().Int("quotas", len(quotas)).Msg("fetching quotas")

		for _, quota := range quotas {
			var vserver, quotaInstanceKey, uid, uName string
			quotaType := quota.GetChildContentS("quota-type")
			tree := quota.GetChildContentS("tree")
			volume := quota.GetChildContentS("volume")
			if my.client.IsClustered() {
				vserver = quota.GetChildContentS("vserver")
			}
			users := quota.GetChildS("quota-users")
			if users != nil {
				quotaUser := users.GetChildS("quota-user")
				if quotaUser != nil {
					uid = quotaUser.GetChildContentS("quota-user-id")
					uName = quotaUser.GetChildContentS("quota-user-name")
				}
			}

			// ignore default quotas and set user/group
			// Rest uses service side filtering to remove default records
			if quotaType == "user" {
				if (uName == "*" && uid == "*") || (uName == "" && uid == "") {
					continue
				}
			} else if quotaType == "group" {
				if uName == "*" || uName == "" {
					continue
				}
			} else if quotaType == "tree" {
				if tree == "" {
					continue
				}
			}

			quotaIndex++

			for attribute, m := range my.data.GetMetrics() {

				objectElem := quota.GetChildS(attribute)
				if objectElem == nil {
					continue
				}

				if attrValue := quota.GetChildContentS(attribute); attrValue != "" {

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
					//set labels
					quotaInstance.SetLabel("type", quotaType)
					quotaInstance.SetLabel("qtree", tree)
					quotaInstance.SetLabel("volume", volume)
					quotaInstance.SetLabel("svm", vserver)
					quotaInstance.SetLabel("index", cluster+"_"+strconv.Itoa(quotaIndex))

					if quotaType == "user" {
						if uName != "" {
							quotaInstance.SetLabel("user", uName)
						} else if uid != "" {
							quotaInstance.SetLabel("user", uid)
						}
						quotaInstance.SetLabel("user_id", uid)
					} else if quotaType == "group" {
						if uName != "" {
							quotaInstance.SetLabel("group", uName)
						} else if uid != "" {
							quotaInstance.SetLabel("group", uid)
						}
						quotaInstance.SetLabel("group_id", uid)
					}

					my.Logger.Debug().Msgf("add (%s) instance: %s.%s.%s", attribute, vserver, volume, tree)

					// populate numeric data
					if value := strings.Split(attrValue, " ")[0]; value != "" {
						// Few quota metrics would have value '-' which means unlimited (ex: disk-limit)
						if value == "-" {
							value = "-1"
						}

						if attribute == "soft-disk-limit" || attribute == "disk-limit" || attribute == "disk-used" {
							quotaInstance.SetLabel("unit", "Kbyte")
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
			}
		}
	}

	my.Logger.Info().
		Int("numQuotas", quotaIndex).
		Int("metrics", numMetrics).
		Str("apiD", apiT.Round(time.Millisecond).String()).
		Str("parseD", parseT.Round(time.Millisecond).String()).
		Str("batchSize", my.batchSize).
		Msg("Collected")
	return []*matrix.Matrix{my.data}, nil
}
