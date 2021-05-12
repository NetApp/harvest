/*
 * Copyright NetApp Inc, 2021 All rights reserved

 Run basic tests to check if CLI args are consistent
*/
package main

import (
	"os/exec"
	"bytes"
	"path"
	"testing"
	"goharvest2/pkg/config"
)

var home string = config.GetHarvestHome()

var commands = [][]string{
	[]string{path.Join(home, "bin/harvest")},
	[]string{path.Join(home, "bin/harvest"), "manager"},
	[]string{path.Join(home, "bin/harvest"), "config"},
	[]string{path.Join(home, "bin/harvest"), "new"},
	[]string{path.Join(home, "bin/poller")},
	[]string{path.Join(home, "bin/zapi")},
	[]string{path.Join(home, "bin/grafana")},
}

// all bins should print usage if no args are provided
func TestPrintUsage(t *testing.T) {
	var (
		out []byte
		err error
	)

	for _, c := range commands {

		t.Logf("exec %v", c)
		if out, err = exec.Command(c[0], c[1:]...).CombinedOutput(); err != nil {
			t.Error(err)
		} else if bytes.HasPrefix(out, []byte("Usage: ")) {
			t.Log("  -> OK: usage printed")
		} else {
			t.Error("  -> FAIL: usage not printed")
		}
	}
}


// all bins should print help if first arg is "help"
func TestPrintHelp(t *testing.T) {
	var (
		out []byte
		err error
	)
	for _, c := range commands {

		h := append(c, "help")
		t.Logf("exec %v", h)
		if out, err = exec.Command(h[0], h[1:]...).CombinedOutput(); err != nil {
			t.Error(err)
		} else if len(out) == 0 {
			t.Error("  -> FAIL: no text printed")
		} else if len(bytes.Split(out, []byte("\n"))) < 2 {
			t.Error("  -> FAIL: probably not help text")
		} else {
			t.Log("  -> OK: multi-line text printed")
		}
	}
}