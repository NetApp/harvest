/*
 * Copyright NetApp Inc, 2021 All rights reserved

 Run basic tests to check if CLI args are consistent
*/
package main

import (
	"bytes"
	"goharvest2/pkg/config"
	"io/ioutil"
	"os/exec"
	"path"
	"testing"
)

var home string = config.GetHarvestHome()
var binsExist bool = binDirExists()
var commands = [][]string{
	{path.Join(home, "bin/harvest")},
	{path.Join(home, "bin/harvest"), "manager"},
	{path.Join(home, "bin/harvest"), "config"},
	{path.Join(home, "bin/harvest"), "new"},
	{path.Join(home, "bin/poller")},
	{path.Join(home, "bin/zapi")},
	{path.Join(home, "bin/grafana")},
}

// all bins should print usage if no args are provided
func TestPrintUsage(t *testing.T) {
	var (
		out []byte
		err error
	)

	if !binsExist {
		t.Log("no binaries compiled, can't run tests")
		return
	}

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

	if !binsExist {
		t.Log("no binaries compiled, can't run tests")
		return
	}
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

func binDirExists() bool {
	if fs, err := ioutil.ReadDir(path.Join(home, "bin/")); err == nil && len(fs) != 0 {
		return true
	}
	return false
}
