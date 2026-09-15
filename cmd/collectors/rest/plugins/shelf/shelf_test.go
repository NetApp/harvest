package shelf

import (
	"testing"

	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
)

// newShelf builds the plugin the same way a poller does. This plugin has no
// client, so Init can be called for real.
func newShelf(t *testing.T) *Shelf {
	t.Helper()
	params := node.NewS("Shelf")
	p := plugin.New("Rest", options.New(), params, nil, "shelf", nil)
	s := &Shelf{AbstractPlugin: p}
	if err := s.Init(conf.Remote{}); err != nil {
		t.Fatalf("Init: %v", err)
	}
	return s
}

// TestRunSetsIsEmbedded covers the isEmbedded label that the shelf dashboards
// and the shelf-power calculation both key off. Mislabelling an embedded shelf
// is what produced the double-counted power in GitHub issues #4251 "Harvest is
// reporting double power for shelf" and #3757 "MetroCluster Shelfs - duplicate
// entry".
func TestRunSetsIsEmbedded(t *testing.T) {
	tests := []struct {
		name       string
		model      string
		moduleType string
		want       string
	}{
		// Rule 1: a module type ending in E means embedded.
		{name: "module type ending in E", model: "DS224-12", moduleType: "IOM12E", want: "Yes"},
		{name: "module type ending in E, lowercase", model: "ds224-12", moduleType: "iom12e", want: "Yes"},

		// Rule 2: the explicitly listed model/module combinations.
		{name: "listed FS424-12 with IOM12F", model: "FS424-12", moduleType: "IOM12F", want: "Yes"},
		{name: "listed DS212-12 with IOM12G", model: "DS212-12", moduleType: "IOM12G", want: "Yes"},
		{name: "listed combination, lowercase", model: "fs424-12", moduleType: "iom12f", want: "Yes"},

		// Non-embedded: the same module on an unlisted model must not match.
		{name: "IOM12F on an unlisted model", model: "DS224-12", moduleType: "IOM12F", want: "No"},
		{name: "IOM12G on an unlisted model", model: "DS224-12", moduleType: "IOM12G", want: "No"},
		{name: "plain external shelf", model: "DS224-12", moduleType: "IOM12", want: "No"},
		{name: "crossed listed combination", model: "FS424-12", moduleType: "IOM12G", want: "No"},
		{name: "empty labels", model: "", moduleType: "", want: "No"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := newShelf(t)

			data := matrix.New("Shelf", "shelf", "shelf")
			inst, err := data.NewInstance("shelf-1")
			assert.Nil(t, err)
			inst.SetLabel("model", tc.model)
			inst.SetLabel("module_type", tc.moduleType)

			got, meta, err := s.Run(map[string]*matrix.Matrix{"shelf": data})

			assert.Nil(t, err)
			assert.Nil(t, got)
			assert.Nil(t, meta)
			assert.Equal(t, inst.GetLabel("isEmbedded"), tc.want)
		})
	}
}

// TestRunSkipsNonExportableInstances pins that a filtered-out shelf is left
// alone, so it cannot acquire a label that would make it look like real data
// downstream.
func TestRunSkipsNonExportableInstances(t *testing.T) {
	s := newShelf(t)

	data := matrix.New("Shelf", "shelf", "shelf")

	exported, err := data.NewInstance("shelf-1")
	assert.Nil(t, err)
	exported.SetLabel("model", "FS424-12")
	exported.SetLabel("module_type", "IOM12F")

	skipped, err := data.NewInstance("shelf-2")
	assert.Nil(t, err)
	skipped.SetLabel("model", "FS424-12")
	skipped.SetLabel("module_type", "IOM12F")
	skipped.SetExportable(false)

	_, _, err = s.Run(map[string]*matrix.Matrix{"shelf": data})
	assert.Nil(t, err)

	assert.Equal(t, exported.GetLabel("isEmbedded"), "Yes")
	assert.Equal(t, skipped.GetLabel("isEmbedded"), "")
}

// TestRunIsIdempotent pins that a second poll does not flip the label, since
// Run mutates the collector's live matrix in place across polls.
func TestRunIsIdempotent(t *testing.T) {
	s := newShelf(t)

	data := matrix.New("Shelf", "shelf", "shelf")
	inst, err := data.NewInstance("shelf-1")
	assert.Nil(t, err)
	inst.SetLabel("model", "DS224-12")
	inst.SetLabel("module_type", "IOM12")

	dataMap := map[string]*matrix.Matrix{"shelf": data}

	_, _, err = s.Run(dataMap)
	assert.Nil(t, err)
	assert.Equal(t, inst.GetLabel("isEmbedded"), "No")

	_, _, err = s.Run(dataMap)
	assert.Nil(t, err)
	assert.Equal(t, inst.GetLabel("isEmbedded"), "No")
}
