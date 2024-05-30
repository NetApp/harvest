package qtree

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/logging"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/tidwall/gjson"
	"os"
	"testing"
)

func NewQtree() *Qtree {
	q := &Qtree{AbstractPlugin: plugin.New("qtree", nil, nil, nil, "qtree", nil)}
	q.Logger = logging.Get()
	q.data = matrix.New(q.Parent+".Qtree", "quota", "quota")
	_, _ = q.data.NewMetricFloat64("space.hard_limit", "disk_limit")
	_, _ = q.data.NewMetricFloat64("space.used.total", "disk_used")
	return q
}

func TestHandlingQuotaMetrics(t *testing.T) {
	q := NewQtree()

	jsonResponse, err := os.ReadFile("testdata/quota.json")
	if err != nil {
		t.Fatalf("Failed to read JSON response from file: %v", err)
	}

	result := gjson.Get(string(jsonResponse), "records").Array()

	quotaCount := 0
	numMetrics := 0

	err = q.handlingQuotaMetrics(result, &quotaCount, &numMetrics)
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
