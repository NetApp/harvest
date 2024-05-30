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
	q := NewQtree()

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

	quotaCount := 0
	numMetrics := 0

	err = q.handlingQuotaMetrics(quotas, &quotaCount, &numMetrics)
	if err != nil {
		t.Errorf("handlingQuotaMetrics returned an error: %v", err)
	}

	if quotaCount != 3 {
		t.Errorf("quotaCount = %d; want 3", quotaCount)
	}
	if numMetrics != 6 {
		t.Errorf("numMetrics = %d; want 6", numMetrics)
	}
}
