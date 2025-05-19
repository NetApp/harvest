package qtree

import (
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	client "github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"log/slog"
	"testing"
)

func NewQtree(historicalLabels bool, testFileName string) plugin.Plugin {
	params := node.NewS("Qtree")
	pp := node.NewS("QtreeParent")
	pp.NewChildS("poller_name", "test")
	pp.NewChildS("addr", "1.2.3.4")
	o := options.Options{IsTest: true}
	q := &Qtree{AbstractPlugin: plugin.New("qtree", &o, params, pp, "qtree", nil)}
	q.SLogger = slog.Default()
	q.historicalLabels = historicalLabels
	q.data = matrix.New(q.Parent+".Qtree", "quota", "quota")
	q.client = client.NewTestClient()
	q.testFilePath = testFileName
	exportOptions := node.NewS("export_options")
	instanceKeys := exportOptions.NewChildS("instance_keys", "")
	// apply all instance keys, instance labels from qtree.yaml to all quota metrics
	keys := []string{"export_policy", "oplocks", "security_style", "status"}
	for _, key := range keys {
		instanceKeys.NewChildS("", key)
	}
	q.data.SetExportOptions(exportOptions)
	_, _ = q.data.NewMetricFloat64("disk-limit", "disk_limit")
	_, _ = q.data.NewMetricFloat64("disk-used", "disk_used")
	return q
}

func TestRunForAllImplementations(t *testing.T) {
	testFileName := "testdata/quotas.xml"
	testCases := []struct {
		name                 string
		createQtree          func(historicalLabels bool, testFileName string) plugin.Plugin
		historicalLabels     bool
		expectedQuotaCount   int
		expectedQtreeCount   int
		quotaInstanceKey     string
		expectedQuotaLabels  int
		withNonExportedQtree bool
	}{
		{
			// Case 1: with historicalLabels = false, all qtrees were exported, total 4 quotas: 2 user/group quota, 1 empty qtree tree quota and 1 non-empty tree quota,
			// 1 empty tree quota will be skipped and 5 labels[qtree, svm, type, unit, volume] would be exported for tree quota.
			name:                 "historicalLabels=false,withNonExportedQtree=false",
			createQtree:          NewQtree,
			historicalLabels:     false,
			withNonExportedQtree: false,
			expectedQuotaCount:   6, // Only 3 quotas each with 2 metrics
			expectedQtreeCount:   3, // All 3 qtrees
			quotaInstanceKey:     "astra_300.trident_qtree_pool_trident_TIXRBILLKA.trident_pvc_2a6d71d9_1c78_4e9a_84a2_59d316adfae9..disk-limit.tree",
			expectedQuotaLabels:  5,
		},
		{
			// Case 2: with historicalLabels = false, 2 qtrees were not exported, total 4 quotas: 2 user/group quota, 1 empty qtree tree quota and 1 non-empty tree quota,
			// 1 empty tree quota will be skipped, 2 qtree will be skipped and 6 labels[qtree, svm, type, unit, user, volume] would be exported for user/group quota.
			name:                 "historicalLabels=false,withNonExportedQtree=true",
			createQtree:          NewQtree,
			historicalLabels:     false,
			withNonExportedQtree: true,
			expectedQuotaCount:   6, // Only 3 quotas each with 2 metrics
			expectedQtreeCount:   1, // Only 1 qtrees, because 2 qtree is not exported
			quotaInstanceKey:     "abcde.abcd_root..root.disk-used.user",
			expectedQuotaLabels:  6,
		},
		{
			// Case 3: with historicalLabels = true, all qtrees were exported, total 4 quotas: 2 user/group quota, 1 empty qtree tree quota and 1 non-empty tree quota,
			// all quotas with 9 labels [export_policy, oplocks, qtree, security_style, status, svm, type, unit, volume] would be exported.
			name:                 "historicalLabels=true,withNonExportedQtree=false",
			createQtree:          NewQtree,
			historicalLabels:     true,
			withNonExportedQtree: false,
			expectedQuotaCount:   8, // All 4 quotas each with 2 metrics
			expectedQtreeCount:   3, // All 3 qtrees
			quotaInstanceKey:     "svm1.volume1...disk-used.tree",
			expectedQuotaLabels:  9,
		},
		{
			// Case 4: with historicalLabels = true, 2 qtrees were not exported, total 4 quotas: 2 user/group quota, 1 empty qtree tree quota and 1 non-empty tree quota,
			// 2 qtree will be skipped, 3 quotas will be skipped and specified quota instancekey hasn't been exported because that qtree is not being exported.
			name:                 "historicalLabels=true,withNonExportedQtree=true",
			createQtree:          NewQtree,
			historicalLabels:     true,
			withNonExportedQtree: true,
			expectedQuotaCount:   2, // Only 1 quotas each with 2 metrics
			expectedQtreeCount:   1, // Only 1 qtrees, because 2 qtree is not exported
			quotaInstanceKey:     "svm1.volume1...disk-used.tree",
			expectedQuotaLabels:  0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runQtreeTest(t, tc.createQtree, tc.historicalLabels, tc.expectedQuotaCount, tc.expectedQtreeCount, tc.quotaInstanceKey, tc.expectedQuotaLabels, tc.withNonExportedQtree, testFileName)
		})
	}
}

// Common test logic for Qtree plugin
func runQtreeTest(t *testing.T, createQtree func(historicalLabels bool, testFileName string) plugin.Plugin, historicalLabels bool, expectedQuotaCount int, expectedQtreeCount int, quotaInstanceKey string, expectedQuotaLabels int, withNonExportedQtree bool, testFileName string) {
	q := createQtree(historicalLabels, testFileName)

	// Initialize the plugin
	if err := q.Init(conf.Remote{}); err != nil {
		t.Fatalf("failed to initialize plugin: %v", err)
	}

	// Create test data
	qtreeData := matrix.New("qtree", "qtree", "qtree")
	if withNonExportedQtree {
		qInstance1, _ := qtreeData.NewInstance("" + "." + "volume1" + "." + "svm1")
		addLabels(qInstance1)
		qInstance1.SetExportable(false)
		qInstance2, _ := qtreeData.NewInstance("trident_pvc_2a6d71d9_1c78_4e9a_84a2_59d316adfae9" + "." + "trident_qtree_pool_trident_TIXRBILLKA" + "." + "astra_300")
		addLabels(qInstance2)
		qInstance3, _ := qtreeData.NewInstance("" + "." + "abcd_root" + "." + "abcde")
		addLabels(qInstance3)
		qInstance3.SetExportable(false)
	} else {
		qInstance1, _ := qtreeData.NewInstance("" + "." + "volume1" + "." + "svm1")
		addLabels(qInstance1)
		qInstance2, _ := qtreeData.NewInstance("trident_pvc_2a6d71d9_1c78_4e9a_84a2_59d316adfae9" + "." + "trident_qtree_pool_trident_TIXRBILLKA" + "." + "astra_300")
		addLabels(qInstance2)
		qInstance3, _ := qtreeData.NewInstance("" + "." + "abcd_root" + "." + "abcde")
		addLabels(qInstance3)
	}

	dataMap := map[string]*matrix.Matrix{
		"qtree": qtreeData,
	}

	// Run the plugin
	output, _, err := q.Run(dataMap)
	if err != nil {
		t.Fatalf("Run method failed: %v", err)
	}

	quotaOutput := output[0]
	verifyInstanceCount(t, quotaOutput, expectedQuotaCount)
	verifyInstanceCount(t, qtreeData, expectedQtreeCount)
	verifyLabelCount(t, quotaOutput, quotaInstanceKey, expectedQuotaLabels)
}

func addLabels(qtreeInstance *matrix.Instance) {
	qtreeInstance.SetLabel("export_policy", "default")
	qtreeInstance.SetLabel("oplocks", "enabled")
	qtreeInstance.SetLabel("security_style", "unix")
	qtreeInstance.SetLabel("status", "normal")
}

func verifyInstanceCount(t *testing.T, output *matrix.Matrix, expectedCount int) {
	// count exportable instances
	currentCount := 0
	for _, i := range output.GetInstances() {
		if i.IsExportable() {
			currentCount++
		}
	}

	// Verify the number of instances
	if currentCount != expectedCount {
		t.Errorf("expected %d instances, got %d", expectedCount, currentCount)
	}
}

func verifyLabelCount(t *testing.T, quotaOutput *matrix.Matrix, quotaInstanceKey string, expectedQuotaLabels int) {
	quotaLabels := 0
	if quotaInstance := quotaOutput.GetInstance(quotaInstanceKey); quotaInstance != nil {
		quotaLabels = len(quotaInstance.GetLabels())
	}

	if quotaLabels != expectedQuotaLabels {
		t.Errorf("labels = %d; want %d", quotaLabels, expectedQuotaLabels)
	}
}

func TestUserIdentifierHandling(t *testing.T) {
	testFileName := "testdata/quotas2.xml"
	testCases := []struct {
		name                string
		expectedInstanceKey []string
	}{
		{
			name: "User identified by user ID",
			expectedInstanceKey: []string{
				"abcde.vol0..0.disk-limit.user",
				"abcde.vol0..0.disk-used.user",
				"abcde.vol0..1.disk-used.user",
				"abcde.vol0..1.disk-limit.user"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q := NewQtree(false, testFileName)

			if err := q.Init(conf.Remote{}); err != nil {
				t.Fatalf("failed to initialize plugin: %v", err)
			}

			qtreeData := matrix.New("qtree", "qtree", "qtree")
			qtreeInstance, _ := qtreeData.NewInstance("svm1.volume1.qtree1")
			addLabels(qtreeInstance)

			dataMap := map[string]*matrix.Matrix{
				"qtree": qtreeData,
			}

			output, _, err := q.Run(dataMap)
			if err != nil {
				t.Fatalf("Run method failed: %v", err)
			}

			quotaOutput := output[0]
			for _, iKey := range tc.expectedInstanceKey {
				if quotaInstance := quotaOutput.GetInstance(iKey); quotaInstance == nil {
					t.Errorf("expected instance key %s not found", tc.expectedInstanceKey)
				}
			}
		})
	}
}
