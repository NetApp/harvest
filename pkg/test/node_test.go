package test

import (
	"goharvest2/pkg/tree"
	"testing"
)

func TestNode_Merge(t *testing.T) {
	defaultTemplate, _ := tree.Import("yaml", "testdata/default.yaml")
	customTemplate, _ := tree.Import("yaml", "testdata/custom.yaml")
	defaultTemplate.Merge(customTemplate)

	// count number of objects post merge
	want := 10
	got := 0
	if objects := defaultTemplate.GetChildS("objects"); objects != nil {
		for range objects.GetChildren() {
			got += 1
		}
		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	}

	// Compare overwritten values for object
	want1 := "node2.yaml"
	got1 := defaultTemplate.GetChildS("objects").GetChildS("Node").GetContentS()

	if got1 != want1 {
		t.Errorf("got %v, want %v", got1, want1)
	}

	// Check if additional object is added
	checkObject := "Sensor"
	if defaultTemplate.GetChildS("objects").GetChildS(checkObject) == nil {
		t.Errorf("missing object %v", checkObject)
	}

	// Compare overwritten values for schedule
	want2 := "650s"
	got2 := defaultTemplate.GetChildS("schedule").GetChildS("instance").GetContentS()

	if got2 != want2 {
		t.Errorf("got %v, want %v", got2, want2)
	}

}
