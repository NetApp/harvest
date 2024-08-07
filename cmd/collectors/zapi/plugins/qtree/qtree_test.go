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

	// Case 1: with historicalLabels = false, total 4 quotas, 2 user/group quota, 1 empty qtree tree quota and 1 non-empty tree quota,
	// 1 empty tree quota will be skipped and 5 labels[qtree, svm, type, unit, volume] would be exported for tree quota.
	q1 := NewQtree()
	testLabels(t, q1, false, quotas, nil, "astra_300.trident_qtree_pool_trident_TIXRBILLKA.trident_pvc_2a6d71d9_1c78_4e9a_84a2_59d316adfae9..disk-limit.tree", 3, 6, 5)

	// Case 2: with historicalLabels = true, total 4 quotas, 2 user/group quota, 1 empty qtree tree quota and 1 non-empty tree quota,
	// 1 empty tree quota will be skipped and 6 labels[qtree, svm, type, unit, user, volume] would be exported for user/group quota.
	q2 := NewQtree()
	testLabels(t, q2, false, quotas, nil, "abcde.abcd_root..root.disk-used.user", 3, 6, 6)
	//testLabels(t, q2, false, quotas, data3, "abcde.abcd_root..1.disk-used", 3, 6, 10)

	// Case 3: with historicalLabels = true, total 4 quotas, 2 user/group quota, 1 empty qtree tree quota and 1 non-empty tree quota,
	// all quotas with 9 labels [export_policy, oplocks, qtree, security_style, status, svm, type, unit, volume] would be exported.
	q3 := NewQtree()
	data := matrix.New(q3.Parent+".Qtree", "qtree", "qtree")
	q3.data.SetExportOptions(exportOptions)
	qtreeInstance1, _ := data.NewInstance("" + "." + "volume1" + "." + "svm1")
	addLabels(qtreeInstance1)
	qtreeInstance2, _ := data.NewInstance("trident_pvc_2a6d71d9_1c78_4e9a_84a2_59d316adfae9" + "." + "trident_qtree_pool_trident_TIXRBILLKA" + "." + "astra_300")
	addLabels(qtreeInstance2)
	qtreeInstance3, _ := data.NewInstance("" + "." + "abcd_root" + "." + "abcde")
	addLabels(qtreeInstance3)
	testLabels(t, q3, true, quotas, data, "svm1.volume1...disk-used.tree", 4, 8, 9)
}

func addLabels(qtreeInstance *matrix.Instance) {
	qtreeInstance.SetLabel("export_policy", "default")
	qtreeInstance.SetLabel("oplocks", "enabled")
	qtreeInstance.SetLabel("security_style", "unix")
	qtreeInstance.SetLabel("status", "normal")
}

func testLabels(t *testing.T, q *Qtree, historicalLabels bool, quotas []*node.Node, data *matrix.Matrix, quotaInstanceKey string, expectedQuotaCount int, expectedQuotaMetricCount int, expectedQuotaLabels int) {
	quotaCount := 0
	numMetrics := 0
	quotaLabels := 0
	var err error

	q.historicalLabels = historicalLabels
	err = q.handlingQuotaMetrics(quotas, data, &quotaCount, &numMetrics)

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
