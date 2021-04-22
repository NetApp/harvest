//
// Copyright NetApp Inc, 2021 All rights reserved
//
// Package Description:
//
// Examples:
//
package template

import (
	"goharvest2/pkg/tree"
	"goharvest2/pkg/tree/node"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

func Run() {

	var harvest_path string
	if harvest_path = os.Getenv("HARVEST_HOME"); harvest_path == "" {
		harvest_path = "/opt/harvest/"
	}

	var err error

	if len(os.Args) < 3 {
		fmt.Println("What to create? (choose: collector, plugin, exporter)")
		os.Exit(1)
	}

	if os.Args[2] == "collector" {
		err = new_collector(harvest_path)
	} else if os.Args[2] == "plugin" {
		err = new_plugin(harvest_path)
	} else {
		fmt.Printf("Sorry, can't create %s\n", os.Args[2])
		os.Exit(1)
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}

func new_plugin(harvest_path string) error {
	var (
		name, collector string
		err             error
		data            []byte
	)

	name = get_name()

	fmt.Printf("collector: ")
	if _, err = fmt.Scanln(&collector); err != nil {
		return err
	}

	if data, err = ioutil.ReadFile(path.Join(harvest_path, "template/", "plugin.go")); err != nil {
		return err
	}

	fp := path.Join(harvest_path, "src/", "collectors/", strings.ToLower(collector), "plugins/", strings.ToLower(name))

	if err = os.MkdirAll(fp, 0755); err != nil {
		return err
	}

	fp = path.Join(fp, strings.ToLower(name)+".go")

	data = bytes.ReplaceAll(data, []byte("PLUGIN"), []byte(name))

	if err = ioutil.WriteFile(fp, data, 0644); err != nil {
		return err
	}

	fmt.Printf("Created source file [%s]\n", fp)
	fmt.Printf("Run \"harvest build collector %s\" when you are ready\n", strings.ToLower(name))
	fmt.Println("Happy coding!")

	return nil
}

func get_name() string {

	var name string

	if len(os.Args) > 3 {
		return os.Args[3]
	} else {
		fmt.Printf("name: ")
		fmt.Scanln(&name)
	}
	return name
}

func new_collector(harvest_path string) error {
	var (
		name, object string
		data         []byte
		err          error
	)

	name = get_name()

	fmt.Printf("object: ")
	if _, err = fmt.Scanln(&object); err != nil {
		return err
	}

	if tfp, err := create_template(harvest_path, name, object); err != nil {
		return err
	} else {
		fmt.Printf("Created collector config [%s]\n", tfp)
	}

	if data, err = ioutil.ReadFile(path.Join(harvest_path, "template/", "collector.go")); err != nil {
		return err
	}

	fp := path.Join(harvest_path, "src/", "collectors/", strings.ToLower(name))

	if err = os.MkdirAll(fp, 0755); err != nil {
		return err
	}

	fp = path.Join(fp, strings.ToLower(name)+".go")

	data = bytes.ReplaceAll(data, []byte("COLLECTOR"), []byte(name))

	if err = ioutil.WriteFile(fp, data, 0644); err != nil {
		return err
	}

	fmt.Printf("Created source file [%s]\n", fp)
	fmt.Printf("Run \"harvest build collector %s\" when you are ready\n", strings.ToLower(name))
	fmt.Println("Happy coding!")

	return nil
}

func create_template(harvest_path, collector, object string) (string, error) {

	t := node.NewS("")
	t.NewChildS("collector", collector)
	t.NewChildS("object", object)

	freq := ""
	fmt.Printf("schedule: ")
	if _, err := fmt.Scanln(&freq); err != nil {
		return "", err
	}

	if !strings.HasSuffix(freq, "s") {
		freq += "s"
	}

	schedule := t.NewChildS("schedule", "")
	schedule.NewChildS("data", freq)

	export := t.NewChildS("export_options", "")
	export.NewChildS("include_all_labels", "True")

	fp := path.Join(harvest_path, "conf/", strings.ToLower(collector))

	if err := os.MkdirAll(fp, 0755); err != nil {
		return "", err
	}

	fp = path.Join(fp, "default.yaml")

	return fp, tree.Export(t, "yaml", fp)
}
