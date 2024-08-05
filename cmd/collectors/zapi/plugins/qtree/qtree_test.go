package qtree

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	client "github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/logging"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"os"
	"testing"
)

func NewQtree() *Qtree {
	q := &Qtree{AbstractPlugin: plugin.New("qtree", nil, nil, nil, "qtree", nil)}
	q.Logger = logging.Get()
	q.data = matrix.New(q.Parent+".Qtree", "quota", "quota")
	q.client = client.NewTestClient()
	_, _ = q.data.NewMetricFloat64("disk-limit", "disk_limit")
	_, _ = q.data.NewMetricFloat64("disk-used", "disk_used")
	return q
}

func TestHandlingQuotaMetrics(t *testing.T) {
	xmlResponse, err := os.ReadFile("testdata/quotas.xml")
	if err != nil {
		t.Fatalf("Failed to read XML response from file: %v", err)
	}

	root, err := tree.LoadXML(xmlResponse)
	if err != nil {
		t.Fatalf("Failed to parse XML response: %v", err)
	}

	response := root.GetChildS("results")
	if response == nil {
		t.Fatalf("empty response: %v", err)
	}

	var quotas []*node.Node
	if x := response.GetChildS("attributes-list"); x != nil {
		quotas = x.GetChildren()
	}

	exportOptions := node.NewS("export_options")
	instanceKeys := exportOptions.NewChildS("instance_keys", "")
	// apply all instance keys, instance labels from qtree.yaml to all quota metrics
	keys := []string{"export_policy", "oplocks", "security_style", "status"}
	for _, key := range keys {
		instanceKeys.NewChildS("", key)
	}

	// Case 1: with historicalLabels = false
	q1 := NewQtree()
	q1.historicalLabels = false
	testLabels(t, q1, quotas, nil, "astra_300.trident_qtree_pool_trident_TIXRBILLKA.trident_pvc_2a6d71d9_1c78_4e9a_84a2_59d316adfae9..disk-limit.tree", 3, 6, 5)

	// Case 2: with historicalLabels = true, only 1 qtree with 2 quotas exist
	q2 := NewQtree()
	data := matrix.New(q2.Parent+".Qtree", "qtree", "qtree")
	qtreeInstance, _ := data.NewInstance("" + "." + "abcd_root" + "." + "abcde")
	qtreeInstance.SetLabel("export_policy", "default")
	qtreeInstance.SetLabel("oplocks", "enabled")
	qtreeInstance.SetLabel("security_style", "unix")
	qtreeInstance.SetLabel("status", "normal")
	q2.data.SetExportOptions(exportOptions)
	q2.historicalLabels = true
	testLabels(t, q2, quotas, data, "abcde.abcd_root..root.disk-used.user", 2, 4, 10)

	// Case 3: with historicalLabels = true and only 1 qtree with 2 quotas exist, but it's not exported
	q3 := NewQtree()
	data3 := matrix.New(q3.Parent+".Qtree", "qtree", "qtree")
	qtreeInstance3, _ := data3.NewInstance("" + "." + "abcd_root" + "." + "abcde")
	qtreeInstance3.SetLabel("export_policy", "default")
	qtreeInstance3.SetLabel("oplocks", "enabled")
	qtreeInstance3.SetLabel("security_style", "unix")
	qtreeInstance3.SetLabel("status", "normal")
	qtreeInstance3.SetExportable(false)
	q3.data.SetExportOptions(exportOptions)
	q3.historicalLabels = true
	testLabels(t, q3, quotas, data3, "abcde.abcd_root..root.disk-used.user", 0, 0, 0)
}

func testLabels(t *testing.T, q *Qtree, quotas []*node.Node, data *matrix.Matrix, quotaInstanceKey string, expectedQuotaCount int, expectedQuotaMetricCount int, expectedQuotaLabels int) {
	quotaCount := 0
	numMetrics := 0
	quotaLabels := 0
	err := q.handlingQuotaMetrics(quotas, data, &quotaCount, &numMetrics)
	if err != nil {
		t.Errorf("handlingQuotaMetrics returned an error: %v", err)
	}

	if quotaInstance := q.data.GetInstance(quotaInstanceKey); quotaInstance != nil {
		quotaLabels = len(quotaInstance.GetLabels())
	}

	if quotaCount != expectedQuotaCount {
		t.Errorf("quotaCount = %d; want %d", quotaCount, expectedQuotaCount)
	}
	if numMetrics != expectedQuotaMetricCount {
		t.Errorf("numMetrics = %d; want %d", numMetrics, expectedQuotaMetricCount)
	}

	if quotaLabels != expectedQuotaLabels {
		t.Errorf("labels = %d; want %d", quotaLabels, expectedQuotaLabels)
	}
}
