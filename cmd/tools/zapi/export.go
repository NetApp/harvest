// Package zapi Copyright NetApp Inc, 2021 All rights reserved
package zapi

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	client "github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/tree/yaml"
	"os"
	"path/filepath"
	"strings"
)

func export(n *node.Node, c *client.Client, args *Args) error {
	switch args.Item {
	case "attrs":
		return exportAttrs()
	case "counters":
		return exportCounters(n, c, args)
	default:
		return errs.New(errs.ErrInvalidItem, args.Item)
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
		if ch.GetChildContentS("is-deprecated") == "true" {
			continue
		}
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
	if dump, err = yaml.Dump(template); err != nil {
		fmt.Println(err)
		return err
	}
	var fp []string

	harvestHomePath = conf.Path("")
	fp = append(fp, harvestHomePath, "conf/", "zapiperf/")

	if c.IsClustered() {
		fp = append(fp, "cdot")
	} else {
		fp = append(fp, "7mode")
	}

	fp = append(fp, "9.8.0", strings.ReplaceAll(args.Object, ":", "_")+".yaml")

	if err = os.MkdirAll(filepath.Join(fp[:5]...), 0750); err != nil {
		fmt.Println("mkdirall")
		return err
	}

	templateFp := filepath.Join(fp...)

	if err = os.WriteFile(templateFp, dump, 0600); err != nil {
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

	if custom, err = collector.ImportTemplate([]string{"conf"}, "custom.yaml", "zapiperf"); err != nil {
		custom = node.NewS("")
		custom.NewChildS("collector", "ZapiPerf")
		custom.NewChildS("objects", "")
	}

	customFp := filepath.Join(harvestHomePath, "conf", "zapiperf", "custom.yaml")

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

	if err = os.WriteFile(customFp, dump, 0600); err != nil {
		fmt.Println("write custom.yaml")
		return err
	}

	fmt.Printf("added template to [%s]\n", customFp)

	return nil
}

func renderObjectName(rawName string) string {
	var name strings.Builder
	name.WriteString(strings.ToUpper(string(rawName[0])))
	i := 1
	size := len(rawName)

	for i < size {

		c := string(rawName[i])

		if c == "_" || c == ":" || c == "-" {
			if i < size-1 {
				name.WriteString(strings.ToUpper(string(rawName[i+1])))
			}
			i += 2
		} else {
			name.WriteString(c)
			i++
		}
	}

	return name.String()
}
