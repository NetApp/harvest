package storagegrid

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/collectors/storagegrid/rest"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/pkg/auth"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"os"
	"testing"
)

const (
	pollerName = "test"
)

// newStorageGrid initializes a new StorageGrid instance for testing
func newStorageGrid(object string, path string) (*StorageGrid, error) {
	opts := options.New(options.WithConfPath("testdata/conf"))
	opts.Poller = pollerName
	opts.HomePath = "testdata"
	opts.IsTest = true
	r := StorageGrid{}
	rest.NewClientFunc = func(_ string, _ string, _ *auth.Credentials) (*rest.Client, error) {
		return rest.NewDummyClient(), nil
	}
	ac := collector.New("StorageGrid", object, opts, collectors.Params(object, path), nil, conf.Remote{})
	err := r.Init(ac)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func Test_SGAddRemoveRestInstances(t *testing.T) {
	conf.TestLoadHarvestConfig("testdata/config.yml")

	sg, err := newStorageGrid("Tenant", "tenant.yaml")
	if err != nil {
		t.Fatalf("failed to create new StorageGrid: %v", err)
	}

	testFile(t, sg, "testdata/tenant.json", 9)
	testFile(t, sg, "testdata/tenant_delete.json", 8)
	testFile(t, sg, "testdata/tenant_add.json", 10)
}

func testFile(t *testing.T, sg *StorageGrid, filename string, expectedLen int) {
	output, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	data := gjson.Get(string(output), "data")
	_ = sg.handleResults(data.Array())

	got := len(sg.Matrix[sg.Object].GetInstances())
	if got != expectedLen {
		t.Errorf("length of matrix = %v, want %v", got, expectedLen)
	}
}
