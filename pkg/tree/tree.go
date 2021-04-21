//
// Copyright NetApp Inc, 2021 All rights reserved
//
// Package Description:
//
// Examples:
//
package tree

import (
    "errors"
    "goharvest2/pkg/tree/json"
    "goharvest2/pkg/tree/node"
    "goharvest2/pkg/tree/xml"
    "goharvest2/pkg/tree/yaml"
    "io/ioutil"
)

func Print(n *node.Node) {
    n.Print(0)
}

func Import(format, filepath string) (*node.Node, error) {

    data, err := ioutil.ReadFile(filepath)

    if err != nil {
        return nil, err
    }

    switch format {
    case "yaml":
        return yaml.Load(data)
    case "xml":
        return xml.Load(data)
    case "json":
        return json.Load(data)
    }

    return nil, errors.New("unknown format: " + format)
}

func Export(n *node.Node, format, filepath string) error {

    var data []byte
    var err error

    switch format {
    case "yaml":
        data, err = yaml.Dump(n)
    case "xml":
        data, err = xml.Dump(n)
    case "json":
        data = json.Dump(n)
    default:
        err = errors.New("unknown format: " + format)
    }

    if err == nil {
        err = ioutil.WriteFile(filepath, data, 0644)
    }
    return err
}

func LoadYaml(data []byte) (*node.Node, error) {
    return yaml.Load(data)
}

func DumpYaml(n *node.Node) ([]byte, error) {
    return yaml.Dump(n)
}

func LoadXml(data []byte) (*node.Node, error) {
    return xml.Load(data)
}

func DumpXml(n *node.Node) ([]byte, error) {
    return xml.Dump(n)
}
