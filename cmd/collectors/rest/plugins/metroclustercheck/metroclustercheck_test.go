package metroclustercheck

import (
	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"log/slog"
	"testing"
)

func TestUpdate(t *testing.T) {
	m := populateData()
	data := matrix.New("metrocluster_check", "metrocluster_check", "metrocluster_check")
	localClusterName := "stiA250mccip-htp-00605_siteA"
	data.SetGlobalLabel("cluster", localClusterName)
	instance1, _ := data.NewInstance("")
	clusterData := "{\n        \"details\": [\n            {\n                \"cluster\": {\n                    \"name\": \"stiA250mccip-htp-00807_siteB\",\n                    \"_links\": {\n                        \"self\": {\n                            \"href\": \"/api/cluster/37f1b2df-5777-11f0-af6c-d039ea4ecf6d\"\n                        }\n                    }\n                },\n                \"timestamp\": \"2025-10-10T08:15:00-04:00\",\n                \"checks\": [\n                    {\n                        \"name\": \"negotiated_switchover_ready\",\n                        \"result\": \"not_applicable\",\n                        \"additional_info\": \"Disaster recovery readiness checks are not performed as part of periodic MetroCluster check. To run these checks, use the \\\"metrocluster check run\\\" command.\"\n                    },\n                    {\n                        \"name\": \"switchback_ready\",\n                        \"result\": \"not_applicable\",\n                        \"additional_info\": \"Disaster recovery readiness checks are not performed as part of periodic MetroCluster check. To run these checks, use the \\\"metrocluster check run\\\" command.\"\n                    },\n                    {\n                        \"name\": \"job_schedules\",\n                        \"result\": \"ok\"\n                    },\n                    {\n                        \"name\": \"licenses\",\n                        \"result\": \"ok\"\n                    },\n                    {\n                        \"name\": \"periodic_check_enabled\",\n                        \"result\": \"ok\"\n                    },\n                    {\n                        \"name\": \"onboard_key_management\",\n                        \"result\": \"ok\",\n                        \"additional_info\": \"success\"\n                    },\n                    {\n                        \"name\": \"external_key_management\",\n                        \"result\": \"ok\",\n                        \"additional_info\": \"success\"\n                    }\n                ]\n            },\n            {\n                \"cluster\": {\n                    \"uuid\": \"ae943971-573c-11f0-b855-d039ea51ec76\",\n                    \"name\": \"stiA250mccip-htp-00605_siteA\",\n                    \"_links\": {\n                        \"self\": {\n                            \"href\": \"/api/cluster/ae943971-573c-11f0-b855-d039ea51ec76\"\n                        }\n                    }\n                },\n                \"timestamp\": \"2025-10-10T08:15:00-04:00\",\n                \"checks\": [\n                    {\n                        \"name\": \"negotiated_switchover_ready\",\n                        \"result\": \"not_applicable\",\n                        \"additional_info\": \"Disaster recovery readiness checks are not performed as part of periodic MetroCluster check. To run these checks, use the \\\"metrocluster check run\\\" command.\"\n                    },\n                    {\n                        \"name\": \"switchback_ready\",\n                        \"result\": \"not_applicable\",\n                        \"additional_info\": \"Disaster recovery readiness checks are not performed as part of periodic MetroCluster check. To run these checks, use the \\\"metrocluster check run\\\" command.\"\n                    },\n                    {\n                        \"name\": \"job_schedules\",\n                        \"result\": \"ok\"\n                    },\n                    {\n                        \"name\": \"licenses\",\n                        \"result\": \"ok\"\n                    },\n                    {\n                        \"name\": \"periodic_check_enabled\",\n                        \"result\": \"ok\"\n                    },\n                    {\n                        \"name\": \"onboard_key_management\",\n                        \"result\": \"ok\",\n                        \"additional_info\": \"success\"\n                    },\n                    {\n                        \"name\": \"external_key_management\",\n                        \"result\": \"ok\",\n                        \"additional_info\": \"success\"\n                    }\n                ]\n            }\n        ],\n        \"summary\": {\n            \"message\": \"\"\n        }\n    }"
	instance1.SetLabel("cluster", clusterData)
	dataMap := map[string]*matrix.Matrix{
		"metrocluster_check": data,
	}

	output, _, err := m.Run(dataMap)
	assert.Nil(t, err)

	outputData := output[0]
	assert.Equal(t, len(outputData.GetInstances()), 14)

	remoteInstance := outputData.GetInstance("stiA250mccip-htp-00807_siteBnegotiated_switchover_ready")
	assert.Equal(t, remoteInstance.GetLabel("type"), "remote")

	localInstance := outputData.GetInstance("stiA250mccip-htp-00605_siteAnegotiated_switchover_ready")
	assert.Equal(t, localInstance.GetLabel("type"), "local")
}

func populateData() plugin.Plugin {
	params := node.NewS("MetroClusterCheck")
	mm := node.NewS("MetroclusterCheckParent")
	mm.NewChildS("poller_name", "test")
	opts := options.New()
	m := &MetroclusterCheck{AbstractPlugin: plugin.New("metrocluster_check", opts, params, mm, "metrocluster_check", nil)}
	m.SLogger = slog.Default()
	m.data = matrix.New(m.Parent+".Metrocluster", "metrocluster_check", "metrocluster_check")
	exportOptions := node.NewS("export_options")
	instanceKeys := exportOptions.NewChildS("instance_keys", "")
	pluginLabels := []string{"result", "name", "node", "aggregate", "volume", "type"}
	for _, label := range pluginLabels {
		instanceKeys.NewChildS("", label)
	}
	m.data.SetExportOptions(exportOptions)
	_, _ = m.data.NewMetricFloat64("cluster_status", "cluster_status")
	return m
}
