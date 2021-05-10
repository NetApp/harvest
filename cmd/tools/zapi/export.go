/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */
package main

import (
	"fmt"
	"goharvest2/cmd/poller/collector"
	client "goharvest2/pkg/api/ontapi/zapi"
	"goharvest2/pkg/errors"
	"goharvest2/pkg/tree/node"
	"goharvest2/pkg/tree/yaml"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

func export(n *node.Node, c *client.Client, args *Args) error {
	switch args.Item {
	case "attrs":
		return exportAttrs(n, c, args)
	case "counters":
		return exportCounters(n, c, args)
	default:
		return errors.New(INVALID_ITEM, args.Item)
	}
}

func exportAttrs(item *node.Node, c *client.Client, args *Args) error {
	return nil
}
func exportCounters(item *node.Node, c *client.Client, args *Args) error {

	var (
		dump   []byte
		err    error
		custom *node.Node
	)

	template := node.NewS("")

	object := args.Object
	fmt.Printf("object: ")
	if _, err := fmt.Scanln(&object); err != nil {
		return err
	}

	object_display := renderObjectName(args.Object)
	template.NewChildS("name", object_display)
	template.NewChildS("query", args.Object)
	template.NewChildS("object", object)

	counters := template.NewChildS("counters", "")
	export_options := template.NewChildS("export_options", "")
	instance_keys := export_options.NewChildS("instance_keys", "")

	for _, ch := range item.GetChildren() {
		if ch.GetChildContentS("is-deprecated") != "true" {
			name := ch.GetChildContentS("name")
			prop := ch.GetChildContentS("properties")

			if strings.Contains(prop, "no-display") {
				continue
			}

			counters.NewChildS("", name)

			if name != "instance_uuid" && strings.Contains(prop, "string") {
				if name == "instance_name" {
					instance_keys.NewChildS("", object)
				} else {
					instance_keys.NewChildS("", name)
				}
			}
		}
	}
	/*
		fmt.Println("\n===========================================================================\n")
		template.Print(0)
		fmt.Println("\n===========================================================================\n")
	*/
	if dump, err = yaml.Dump(template); err != nil {
		fmt.Println(err)
		return err
	} /*/else {
		fmt.Println(string(dump))
	}
	fmt.Println("\n===========================================================================\n")
	*/
	fp := make([]string, 0)

	fp = append(fp, harvestConfPath)
	fp = append(fp, "conf/")
	fp = append(fp, "zapiperf/")

	if c.IsClustered() {
		fp = append(fp, "cdot")
	} else {
		fp = append(fp, "7mode")
	}

	//fp = append(fp, fmt.Sprintf("%d.%d.%d", c.System.Version[0], c.System.Version[1], c.System.Version[2]))
	fp = append(fp, "9.8.0")
	fp = append(fp, strings.ReplaceAll(args.Object, ":", "_")+".yaml")

	if err = os.MkdirAll(path.Join(fp[:5]...), 0755); err != nil {
		fmt.Println("mkdirall")
		return err
	}

	template_fp := path.Join(fp...)

	if err = ioutil.WriteFile(template_fp, dump, 0644); err != nil {
		fmt.Println("writefile")
		return err
	}

	fmt.Printf("exported to [%s]\n", template_fp)

	answer := ""
	fmt.Printf("enable template? [y/N]: ")
	if _, err := fmt.Scanln(&answer); err != nil {
		return err
	}

	if answer = strings.ToLower(answer); answer != "y" && answer != "yes" {
		return nil
	}

	if custom, err = collector.ImportTemplate(harvestConfPath, "custom.yaml", "zapiperf"); err != nil {
		custom = node.NewS("")
		custom.NewChildS("collector", "ZapiPerf")
		custom.NewChildS("objects", "")
	}

	custom_fp := path.Join(harvestConfPath, "conf/", "zapiperf/", "custom.yaml")

	if objects := custom.GetChildS("objects"); objects != nil {

		if objects.GetChildS(object_display) != nil {
			fmt.Printf("[%s] already in custom template [%s]\n", object_display, custom_fp)
			return nil
		}

		objects.NewChildS(object_display, strings.ReplaceAll(args.Object, ":", "_")+".yaml")
	}

	if dump, err = yaml.Dump(custom); err != nil {
		fmt.Println("dump custom")
		return err
	}

	if err = ioutil.WriteFile(custom_fp, dump, 0644); err != nil {
		fmt.Println("write custom.yaml")
		return err
	}

	fmt.Printf("added template to [%s]\n", custom_fp)

	return nil
}

func renderObjectName(raw_name string) string {

	name := strings.ToUpper(string(raw_name[0]))
	i := 1
	size := len(raw_name)

	for i < size {

		c := string(raw_name[i])

		if c == "_" || c == ":" || c == "-" {
			if i < size-1 {
				name += strings.ToUpper(string(raw_name[i+1]))
			}
			i += 2
		} else {
			name += c
			i += 1
		}
	}

	return name
}
