/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */
package stub

import (
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
	"goharvest2/pkg/conf"
	"goharvest2/pkg/tree"
	"goharvest2/pkg/tree/node"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

var (
	harvestHomePath = conf.GetHarvestHomePath()
)

var NewCmd = &cobra.Command{
	Use:   "new",
	Short: "Create new collector, plugin or exporter (for developers)",
}

var collectorCmd = &cobra.Command{
	Use:   "collector",
	Short: "create collector",
	Run: func(cmd *cobra.Command, args []string) {
		run("collector")
	},
}

var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "create plugin",
	Run: func(cmd *cobra.Command, args []string) {
		run("plugin")
	},
}

var exporterCmd = &cobra.Command{
	Use:   "exporter",
	Short: "create exporter",
	Run: func(cmd *cobra.Command, args []string) {
		run("exporter")
	},
}

func run(cmd string) {
	var err error

	switch cmd {
	case "collector":
		err = newCollector()
	case "plugin":
		err = newPlugin()
	case "exporter":
		err = newExporter()
	default:
		fmt.Printf("Sorry, can't create %s\n", cmd)
		os.Exit(1)
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}

func newPlugin() error {
	var (
		name, collector string
		err             error
		data            []byte
	)

	name = getName()

	fmt.Println("For which collector is this plugin?")
	fmt.Printf("collector: ")
	if _, err = fmt.Scanln(&collector); err != nil {
		return err
	}

	if data, err = ioutil.ReadFile(path.Join(harvestHomePath, "cmd/harvest/stub/_plugin.go_")); err != nil {
		return err
	}

	fp := path.Join(harvestHomePath, "cmd/collectors/", strings.ToLower(collector), "plugins/", strings.ToLower(name))

	if err = os.MkdirAll(fp, 0755); err != nil {
		return err
	}

	data = bytes.ReplaceAll(data, []byte("PLUGIN"), []byte(name))

	fp = path.Join(fp, strings.ToLower(name)+".go")
	if err = ioutil.WriteFile(fp, data, 0644); err != nil {
		return err
	}

	fmt.Printf("Created source file [%s]\n", fp)
	fmt.Println("Run \"make collectors\" when you are ready")
	fmt.Println("Happy coding!")

	return nil
}

func newExporter() error {
	var (
		name string
		data []byte
		err  error
	)

	name = getName()

	if data, err = ioutil.ReadFile(path.Join(harvestHomePath, "cmd/harvest/stub/_exporter.go_")); err != nil {
		return err
	}

	fp := path.Join(harvestHomePath, "cmd/exporters/", strings.ToLower(name))

	if err = os.MkdirAll(fp, 0755); err != nil {
		return err
	}

	data = bytes.ReplaceAll(data, []byte("EXPORTER"), []byte(name))

	fp = path.Join(fp, strings.ToLower(name)+".go")
	if err = ioutil.WriteFile(fp, data, 0644); err != nil {
		return err
	}

	fmt.Printf("Created source file [%s]\n", fp)
	fmt.Println("Run \"make exporters\" when you are ready")
	fmt.Println("Happy coding!")

	return nil
}

func newCollector() error {
	var (
		name, object string
		data         []byte
		err          error
	)

	name = getName()

	fmt.Println("What object does this collector collect?")
	fmt.Println("(choose a name that best describes your metrics/instances)")
	fmt.Printf("object: ")
	if _, err = fmt.Scanln(&object); err != nil {
		return err
	}

	if tfp, err := createTemplate(name, object); err != nil {
		return err
	} else {
		fmt.Printf("Created collector config [%s]\n", tfp)
	}

	if data, err = ioutil.ReadFile(path.Join(harvestHomePath, "cmd/harvest/stub/_collector.go_")); err != nil {
		return err
	}

	fp := path.Join(harvestHomePath, "cmd/", "collectors/", strings.ToLower(name))

	if err = os.MkdirAll(fp, 0755); err != nil {
		return err
	}

	fp = path.Join(fp, strings.ToLower(name)+".go")

	data = bytes.ReplaceAll(data, []byte("COLLECTOR"), []byte(name))

	if err = ioutil.WriteFile(fp, data, 0644); err != nil {
		return err
	}

	fmt.Printf("Created source file [%s]\n", fp)
	fmt.Println("Run \"make collectors\" when you are ready")
	fmt.Println("Happy coding!")

	return nil
}

func getName() string {

	var name string

	if len(os.Args) > 3 {
		return os.Args[3]
	} else {
		fmt.Printf("name: ")
		_, _ = fmt.Scanln(&name)
	}
	return name
}

func createTemplate(collector, object string) (string, error) {

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

	fp := path.Join(harvestHomePath, "conf/", strings.ToLower(collector))

	if err := os.MkdirAll(fp, 0755); err != nil {
		return "", err
	}

	fp = path.Join(fp, "default.yaml")

	return fp, tree.Export(t, "yaml", fp)
}

func init() {
	NewCmd.AddCommand(collectorCmd, pluginCmd, exporterCmd)
}
