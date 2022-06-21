// Package zapi Copyright NetApp Inc, 2021 All rights reserved
package zapi

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	client "github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errors"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/tree/yaml"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

func export(n *node.Node, c *client.Client, args *Args) error {
	switch args.Item {
	case "attrs":
		return exportAttrs()
	case "counters":
		return exportCounters(n, c, args)
	default:
		return errors.New(InvalidItem, args.Item)
	}
}

func exportAttrs() error {
	return nil
}
func exportCounters(item *node.Node, c *client.Client, args *Args) error {

	var (
		dump            []byte
		err             error
		custom          *node.Node
		harvestHomePath string
	)

	template := node.NewS("")

	object := args.Object
	fmt.Printf("object: ")
	if _, err := fmt.Scanln(&object); err != nil {
		return err
	}

	objectDisplay := renderObjectName(args.Object)
	template.NewChildS("name", objectDisplay)
	template.NewChildS("query", args.Object)
	template.NewChildS("object", object)

	counters := template.NewChildS("counters", "")
	exportOptions := template.NewChildS("export_options", "")
	instanceKeys := exportOptions.NewChildS("instance_keys", "")

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
					instanceKeys.NewChildS("", object)
				} else {
					instanceKeys.NewChildS("", name)
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

	harvestHomePath = conf.GetHarvestHomePath()
	fp = append(fp, harvestHomePath)
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

	templateFp := path.Join(fp...)

	if err = ioutil.WriteFile(templateFp, dump, 0600); err != nil {
		fmt.Println("writefile")
		return err
	}

	fmt.Printf("exported to [%s]\n", templateFp)

	answer := ""
	fmt.Printf("enable template? [y/N]: ")
	if _, err := fmt.Scanln(&answer); err != nil {
		return err
	}

	if answer = strings.ToLower(answer); answer != "y" && answer != "yes" {
		return nil
	}

	if custom, err = collector.ImportTemplate(harvestHomePath, "custom.yaml", "zapiperf"); err != nil {
		custom = node.NewS("")
		custom.NewChildS("collector", "ZapiPerf")
		custom.NewChildS("objects", "")
	}

	customFp := path.Join(harvestHomePath, "conf/", "zapiperf/", "custom.yaml")

	if objects := custom.GetChildS("objects"); objects != nil {

		if objects.GetChildS(objectDisplay) != nil {
			fmt.Printf("[%s] already in custom template [%s]\n", objectDisplay, customFp)
			return nil
		}

		objects.NewChildS(objectDisplay, strings.ReplaceAll(args.Object, ":", "_")+".yaml")
	}

	if dump, err = yaml.Dump(custom); err != nil {
		fmt.Println("dump custom")
		return err
	}

	if err = ioutil.WriteFile(customFp, dump, 0600); err != nil {
		fmt.Println("write custom.yaml")
		return err
	}

	fmt.Printf("added template to [%s]\n", customFp)

	return nil
}

func renderObjectName(rawName string) string {

	name := strings.ToUpper(string(rawName[0]))
	i := 1
	size := len(rawName)

	for i < size {

		c := string(rawName[i])

		if c == "_" || c == ":" || c == "-" {
			if i < size-1 {
				name += strings.ToUpper(string(rawName[i+1]))
			}
			i += 2
		} else {
			name += c
			i++
		}
	}

	return name
}
