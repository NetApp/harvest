// Copyright NetApp Inc, 2021 All rights reserved

package qtree

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
	"log/slog"
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
	historicalLabels bool   // supports labels, metrics for 22.05
	qtreeMetrics     bool   // supports quota metrics with qtree prefix
	testFilePath     string // Used only from unit test
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Qtree{AbstractPlugin: p}
}

func (q *Qtree) Init(remote conf.Remote) error {

	var err error

	if err := q.InitAbc(); err != nil {
		return err
	}

	if q.Options.IsTest {
		return nil
	}

	if q.client, err = zapi.New(conf.ZapiPoller(q.ParentParams), q.Auth); err != nil {
		q.SLogger.Error("connecting", slogx.Err(err))
		return err
	}

	if err := q.client.Init(5, remote); err != nil {
		return err
	}

	if q.client.IsClustered() {
		q.query = "quota-report-iter"
	} else {
		q.query = "quota-report"
	}
	q.SLogger.Debug("plugin connected!")

	q.data = matrix.New(q.Parent+".Qtree", "quota", "quota")
	q.instanceKeys = make(map[string]string)
	q.instanceLabels = make(map[string]map[string]string)
	q.historicalLabels = false

	if q.Params.HasChildS("qtreeMetrics") {
		q.qtreeMetrics = true
	}

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
		instanceKeys.NewChildS("", "unit")
		instanceKeys.NewChildS("", "user")
		instanceKeys.NewChildS("", "userId")
		instanceKeys.NewChildS("", "group")
		instanceKeys.NewChildS("", "groupId")

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
			q.SLogger.Error("add metric", slogx.Err(err))
			return err
		}
	}

	q.SLogger.Debug("added data with metrics", slog.Int("metrics", len(q.data.GetMetrics())))

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

	for {
		responseData, err := q.client.InvokeBatchRequest(request, tag, q.testFilePath)
		if err != nil {
			return nil, nil, err
		}
		response = responseData.Result
		tag = responseData.Tag
		ad = responseData.Rd
		pd = responseData.Pd

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
			q.SLogger.Debug("no quota instances found")
			return nil, q.client.Metadata, nil
		}

		q.SLogger.Debug("fetching quotas", slog.Int("quotas", len(quotas)))

		// Populate metrics with quota prefix
		err = q.handlingQuotaMetrics(quotas, data, &quotaIndex, &numMetrics)

		if err != nil {
			return nil, nil, err
		}
	}

	q.client.Metadata.PluginInstances = uint64(quotaIndex) //nolint:gosec

	q.SLogger.Info(
		"Collected",
		slog.Int("numQuotas", quotaIndex),
		slog.Int("metrics", numMetrics),
		slog.String("apiD", apiT.Round(time.Millisecond).String()),
		slog.String("parseD", parseT.Round(time.Millisecond).String()),
		slog.String("batchSize", q.batchSize),
	)

	if q.qtreeMetrics || q.historicalLabels {
		// metrics with qtree prefix and quota prefix are available to support backward compatibility
		qtreePluginData := q.data.Clone(matrix.With{Data: true, Metrics: true, Instances: true, ExportInstances: true})
		qtreePluginData.UUID = q.Parent + ".Qtree"
		qtreePluginData.Object = "qtree"
		qtreePluginData.Identifier = "qtree"
		return []*matrix.Matrix{qtreePluginData, q.data}, q.client.Metadata, nil
	}
	return []*matrix.Matrix{q.data}, q.client.Metadata, nil
}

func (q *Qtree) handlingQuotaMetrics(quotas []*node.Node, data *matrix.Matrix, quotaCount *int, numMetrics *int) error {
	for _, quota := range quotas {
		var vserver, quotaInstanceKey, uid, uName string
		var qtreeInstance *matrix.Instance

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
		switch quotaType {
		case "user":
			if (uName == "*" && uid == "*") || (uName == "" && uid == "") {
				continue
			}
		case "group":
			if uName == "*" || uName == "" {
				continue
			}
		case "tree":
			if !q.historicalLabels && tree == "" {
				continue
			}
		}

		// In 22.05, populate metrics with qtree labels
		if q.historicalLabels {
			if q.client.IsClustered() {
				qtreeInstance = data.GetInstance(tree + "." + volume + "." + vserver)
			} else {
				qtreeInstance = data.GetInstance(volume + "." + tree)
			}
			if qtreeInstance == nil {
				q.SLogger.Warn(
					"no instance matching tree.volume.vserver",
					slog.String("tree", tree),
					slog.String("volume", volume),
					slog.String("vserver", vserver),
				)
				continue
			}
			if !qtreeInstance.IsExportable() {
				continue
			}
		}
		*quotaCount++

		for attribute, m := range q.data.GetMetrics() {
			objectElem := quota.GetChildS(attribute)
			if objectElem == nil {
				continue
			}

			if attrValue := quota.GetChildContentS(attribute); attrValue != "" {
				// Ex. InstanceKey: SVMA.vol1Abc.qtree1.5.disk-limit
				if q.client.IsClustered() {
					quotaInstanceKey = vserver + "." + volume + "." + tree + "." + uName + "." + uid + "." + attribute + "." + quotaType
				} else {
					quotaInstanceKey = volume + "." + tree + "." + uName + "." + uid + "." + attribute
				}
				quotaInstance, err := q.data.NewInstance(quotaInstanceKey)
				if err != nil {
					q.SLogger.Debug("add instance", slogx.Err(err), slog.String("attribute", attribute))
					return err
				}

				// set labels
				quotaInstance.SetLabel("type", quotaType)
				quotaInstance.SetLabel("qtree", tree)
				quotaInstance.SetLabel("volume", volume)
				quotaInstance.SetLabel("svm", vserver)

				// In 22.05, populate metrics with qtree labels
				if q.historicalLabels {
					if qtreeInstance != nil {
						for _, label := range q.data.GetExportOptions().GetChildS("instance_keys").GetAllChildContentS() {
							if value := qtreeInstance.GetLabel(label); value != "" {
								quotaInstance.SetLabel(label, value)
							}
						}
					}
					// If the Qtree is the volume itself, then qtree label is empty, so copy the volume name to qtree.
					if tree == "" {
						quotaInstance.SetLabel("qtree", volume)
					}
				}

				switch quotaType {
				case "user":
					quotaInstance.SetLabel("user", uName)
					quotaInstance.SetLabel("userId", uid)
				case "group":
					quotaInstance.SetLabel("group", uName)
					quotaInstance.SetLabel("groupId", uid)
				}

				// populate numeric data
				if value := strings.Split(attrValue, " ")[0]; value != "" {
					// Few quota metrics would have value '-' which means unlimited (ex: disk-limit)
					if value == "-" {
						// In 22.05, populate metrics value with 0
						if q.historicalLabels {
							value = "0"
						} else {
							value = "-1"
						}
					}

					if attribute == "soft-disk-limit" || attribute == "disk-limit" || attribute == "disk-used" {
						quotaInstance.SetLabel("unit", "Kbyte")
					}
					if err := m.SetValueString(quotaInstance, value); err != nil {
						q.SLogger.Debug(
							"failed to parse value",
							slogx.Err(err),
							slog.String("attribute", attribute),
							slog.String("value", value),
						)
					} else {
						*numMetrics++
					}
				}
			}
		}
	}
	return nil
}
