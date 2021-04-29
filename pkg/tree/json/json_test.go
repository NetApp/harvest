/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */
package json

import (
	"goharvest2/pkg/tree/json"
	"goharvest2/pkg/tree/node"
	"testing"
)

func TestLoad(t *testing.T) {

	var data []byte
	var err error
	var root *node.Node

	data = []byte(`{"id":61,"uid":"jratWpwMz","title":"Harvest 2.0","url":"/dashboards/f/jratWpwMz/harvest-2-0","hasAcl":false,"canSave":true,"canEdit":true,"canAdmin":false,"createdBy":"Anonymous","created":"2021-03-21T16:11:38+04:00","updatedBy":"Anonymous","updated":"2021-03-21T16:11:38+04:00","version":1}`)

	if root, err = json.Load(data); err != nil {
		t.Fatal(err)
	}

	if root == nil {
		t.Fatalf("node is nil")
	}

	if len(root.GetChildren()) != 13 {
		t.Errorf("parsed node with %d children (13 were expected):", len(root.GetChildren()))
	} else {
		t.Logf("parsed node with %d children:", len(root.GetChildren()))
	}

	root.Print(0)
}

func TestDump(t *testing.T) {

	var root *node.Node
	var data, dump []byte

	data = []byte(`{"title":"harvest","version":2.0,"admin":true}`)

	root = node.NewS("")
	root.NewChildS("title", "harvest")
	root.NewChildS("version", "2.0")
	root.NewChildS("admin", "true")

	t.Logf("dumping node:")
	root.Print(0)

	dump = json.Dump(root)

	t.Logf("%-10s [%s]", "expected:", string(data))

	if string(dump) != string(data) {
		t.Errorf("%-10s [%s]", "got:", string(dump))
	}
}
