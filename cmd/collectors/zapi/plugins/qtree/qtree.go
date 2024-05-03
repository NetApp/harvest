// Copyright NetApp Inc, 2021 All rights reserved

package qtree

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/conf"
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
	data             *matrix.Matrix
	instanceKeys     map[string]string
	instanceLabels   map[string]map[string]string
	batchSize        string
	client           *zapi.Client
	query            string
	quotaType        []string
	historicalLabels bool // supports labels, metrics for 22.05
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Qtree{AbstractPlugin: p}
}

func (q *Qtree) Init() error {

	var err error

	if err := q.InitAbc(); err != nil {
		return err
	}

	if q.client, err = zapi.New(conf.ZapiPoller(q.ParentParams), q.Auth); err != nil {
		q.Logger.Error().Stack().Err(err).Msg("connecting")
		return err
	}

	if err := q.client.Init(5); err != nil {
		return err
	}

	if q.client.IsClustered() {
		q.query = "quota-report-iter"
	} else {
		q.query = "quota-report"
	}
	q.Logger.Debug().Msg("plugin connected!")

	q.data = matrix.New(q.Parent+".Qtree", "quota", "quota")
	q.instanceKeys = make(map[string]string)
	q.instanceLabels = make(map[string]map[string]string)
	q.historicalLabels = false

	if q.Params.HasChildS("historicalLabels") {
		exportOptions := node.NewS("export_options")
		instanceKeys := exportOptions.NewChildS("instance_keys", "")

		// apply all instance keys, instance labels from parent (qtree.yaml) to all quota metrics
		if exportOption := q.ParentParams.GetChildS("export_options"); exportOption != nil {
			// parent instancekeys would be added in plugin metrics
			if parentKeys := exportOption.GetChildS("instance_keys"); parentKeys != nil {
				for _, parentKey := range parentKeys.GetAllChildContentS() {
					instanceKeys.NewChildS("", parentKey)
				}
			}
			// parent instacelabels would be added in plugin metrics
			if parentLabels := exportOption.GetChildS("instance_labels"); parentLabels != nil {
				for _, parentLabel := range parentLabels.GetAllChildContentS() {
					instanceKeys.NewChildS("", parentLabel)
				}
			}
		}

		instanceKeys.NewChildS("", "type")
		instanceKeys.NewChildS("", "index")
		instanceKeys.NewChildS("", "unit")

		q.data.SetExportOptions(exportOptions)
		q.historicalLabels = true
	}

	quotaType := q.Params.GetChildS("quotaType")
	if quotaType != nil {
		q.quotaType = []string{}
		q.quotaType = append(q.quotaType, quotaType.GetAllChildContentS()...)
	} else {
		q.quotaType = []string{"user", "group", "tree"}
	}

	objects := q.Params.GetChildS("objects")
	if objects == nil {
		return errs.New(errs.ErrMissingParam, "objects")
	}

	for _, obj := range objects.GetAllChildContentS() {

		metricName, display, _, _ := util.ParseMetric(obj)

		_, err := q.data.NewMetricFloat64(metricName, display)
		if err != nil {
			q.Logger.Error().Stack().Err(err).Msg("add metric")
			return err
		}
	}

	q.Logger.Debug().Msgf("added data with %d metrics", len(q.data.GetMetrics()))

	// setup batchSize for request
	q.batchSize = BatchSize
	if q.client.IsClustered() {
		if b := q.Params.GetChildContentS("batch_size"); b != "" {
			if _, err := strconv.Atoi(b); err == nil {
				q.batchSize = b
			}
		}
	}

	return nil
}

func (q *Qtree) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {
	var (
		request, response *node.Node
		quotas            []*node.Node
		ad, pd            time.Duration // Request/API time, Parse time, Fetch time
		err               error
		numMetrics        int
	)

	data := dataMap[q.Object]
	q.client.Metadata.Reset()

	apiT := 0 * time.Second
	parseT := 0 * time.Second

	// Purge and reset data
	q.data.PurgeInstances()
	q.data.Reset()

	// Set all global labels from zapi.go if already not exist
	q.data.SetGlobalLabels(data.GetGlobalLabels())

	request = node.NewXMLS(q.query)

	if q.client.IsClustered() && q.batchSize != "" {
		request.NewChildS("max-records", q.batchSize)
		// add quota filter
		query := request.NewChildS("query", "")
		quotaQuery := query.NewChildS("quota", "")
		quotaQuery.NewChildS("quota-type", strings.Join(q.quotaType, "|"))
	}

	tag := "initial"
	quotaIndex := 0

	// In 22.05, all qtrees were exported
	if q.historicalLabels {
		for _, qtreeInstance := range data.GetInstances() {
			qtreeInstance.SetExportable(true)
		}
	}

	cluster := data.GetGlobalLabels()["cluster"]

	for {
		response, tag, ad, pd, err = q.client.InvokeBatchWithTimers(request, tag)

		if err != nil {
			return nil, nil, err
		}

		if response == nil {
			break
		}

		apiT += ad
		parseT += pd

		if q.client.IsClustered() {
			if x := response.GetChildS("attributes-list"); x != nil {
				quotas = x.GetChildren()
			}
		} else {
			quotas = response.SearchChildren([]string{"quota"})
		}

		if len(quotas) == 0 {
			q.Logger.Debug().Msg("no quota instances found")
			return nil, q.client.Metadata, nil
		}

		q.Logger.Debug().Int("quotas", len(quotas)).Msg("fetching quotas")

		if q.historicalLabels {
			// In 22.05, populate metrics with qtree prefix and old labels
			err = q.handlingHistoricalMetrics(quotas, data, cluster, &quotaIndex, &numMetrics)
		} else {
			// Populate metrics with quota prefix and current labels
			err = q.handlingQuotaMetrics(quotas, cluster, &quotaIndex, &numMetrics)
		}

		if err != nil {
			return nil, nil, err
		}
	}

	q.Logger.Info().
		Int("numQuotas", quotaIndex).
		Int("metrics", numMetrics).
		Str("apiD", apiT.Round(time.Millisecond).String()).
		Str("parseD", parseT.Round(time.Millisecond).String()).
		Str("batchSize", q.batchSize).
		Msg("Collected")

	// metrics with qtree prefix and quota prefix are available to support backward compatibility
	qtreePluginData := q.data.Clone(matrix.With{Data: true, Metrics: true, Instances: true, ExportInstances: true})
	qtreePluginData.UUID = q.Parent + ".Qtree"
	qtreePluginData.Object = "qtree"
	qtreePluginData.Identifier = "qtree"
	return []*matrix.Matrix{qtreePluginData, q.data}, q.client.Metadata, nil
}

func (q *Qtree) handlingHistoricalMetrics(quotas []*node.Node, data *matrix.Matrix, cluster string, quotaIndex *int, numMetrics *int) error {
	for qIndex, quota := range quotas {
		var vserver, quotaInstanceKey string
		var qtreeInstance *matrix.Instance

		quotaType := quota.GetChildContentS("quota-type")
		tree := quota.GetChildContentS("tree")
		volume := quota.GetChildContentS("volume")
		if q.client.IsClustered() {
			vserver = quota.GetChildContentS("vserver")
		}

		*quotaIndex++
		for attribute, m := range q.data.GetMetrics() {

			objectElem := quota.GetChildS(attribute)
			if objectElem == nil {
				continue
			}

			if attrValue := quota.GetChildContentS(attribute); attrValue != "" {
				if q.client.IsClustered() {
					qtreeInstance = data.GetInstance(tree + "." + volume + "." + vserver)
				} else {
					qtreeInstance = data.GetInstance(volume + "." + tree)
				}
				if qtreeInstance == nil {
					q.Logger.Warn().
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
				if q.client.IsClustered() {
					quotaInstanceKey = vserver + "." + volume + "." + tree + "." + strconv.Itoa(qIndex) + "." + attribute
				} else {
					quotaInstanceKey = volume + "." + tree + "." + strconv.Itoa(qIndex) + "." + attribute
				}
				quotaInstance, err := q.data.NewInstance(quotaInstanceKey)

				if err != nil {
					q.Logger.Debug().Msgf("add (%s) instance: %v", attribute, err)
					return err
				}

				for _, label := range q.data.GetExportOptions().GetChildS("instance_keys").GetAllChildContentS() {
					if value := qtreeInstance.GetLabel(label); value != "" {
						quotaInstance.SetLabel(label, value)
					}
				}

				// set labels
				quotaInstance.SetLabel("type", quotaType)
				quotaInstance.SetLabel("qtree", tree)
				quotaInstance.SetLabel("volume", volume)
				quotaInstance.SetLabel("svm", vserver)
				quotaInstance.SetLabel("index", cluster+"_"+strconv.Itoa(*quotaIndex))

				// If the Qtree is the volume itself, then qtree label is empty, so copy the volume name to qtree.
				if tree == "" {
					quotaInstance.SetLabel("qtree", volume)
				}

				// populate numeric data
				if value := strings.Split(attrValue, " ")[0]; value != "" {
					// Few quota metrics would have value '-' which means unlimited (ex: disk-limit)
					if value == "-" {
						value = "0"
					}
					if attribute == "soft-disk-limit" || attribute == "disk-limit" || attribute == "disk-used" {
						quotaInstance.SetLabel("unit", "Kbyte")
					}
					if err := m.SetValueString(quotaInstance, value); err != nil {
						q.Logger.Debug().Msgf("(%s) failed to parse value (%s): %v", attribute, value, err)
					} else {
						*numMetrics++
					}
				}

			}
		}
	}
	return nil
}

func (q *Qtree) handlingQuotaMetrics(quotas []*node.Node, cluster string, quotaCount *int, numMetrics *int) error {
	for _, quota := range quotas {
		var vserver, quotaInstanceKey, uid, uName string

		quotaType := quota.GetChildContentS("quota-type")
		tree := quota.GetChildContentS("tree")
		volume := quota.GetChildContentS("volume")
		if q.client.IsClustered() {
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
		switch {
		case quotaType == "user":
			if (uName == "*" && uid == "*") || (uName == "" && uid == "") {
				continue
			}
		case quotaType == "group":
			if uName == "*" || uName == "" {
				continue
			}
		case quotaType == "tree":
			if tree == "" {
				continue
			}
		}
		*quotaCount++
		quotaIndex := volume + "." + tree + "." + uName

		for attribute, m := range q.data.GetMetrics() {

			objectElem := quota.GetChildS(attribute)
			if objectElem == nil {
				continue
			}

			if attrValue := quota.GetChildContentS(attribute); attrValue != "" {

				// Ex. InstanceKey: SVMA.vol1Abc.qtree1.5.disk-limit
				if q.client.IsClustered() {
					quotaInstanceKey = vserver + "." + quotaIndex + "." + attribute
				} else {
					quotaInstanceKey = quotaIndex + "." + attribute
				}
				quotaInstance, err := q.data.NewInstance(quotaInstanceKey)
				if err != nil {
					q.Logger.Debug().Msgf("add (%s) instance: %v", attribute, err)
					return err
				}
				// set labels
				quotaInstance.SetLabel("type", quotaType)
				quotaInstance.SetLabel("qtree", tree)
				quotaInstance.SetLabel("volume", volume)
				quotaInstance.SetLabel("svm", vserver)
				quotaInstance.SetLabel("index", cluster+"_"+quotaIndex)

				if quotaType == "user" {
					if uName != "" {
						quotaInstance.SetLabel("user", uName)
					} else if uid != "" {
						quotaInstance.SetLabel("user", uid)
					}
				} else if quotaType == "group" {
					if uName != "" {
						quotaInstance.SetLabel("group", uName)
					} else if uid != "" {
						quotaInstance.SetLabel("group", uid)
					}
				}

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
						q.Logger.Debug().Msgf("(%s) failed to parse value (%s): %v", attribute, value, err)
					} else {
						*numMetrics++
					}
				}
			}
		}
	}
	return nil
}
