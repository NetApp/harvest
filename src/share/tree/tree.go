package tree

import (
    "io/ioutil"
    "goharvest2/share/tree/node"
    "goharvest2/share/tree/yaml"
    "goharvest2/share/tree/xml"
)

func Print(n *node.Node) {
    n.Print(0)
}

func ImportYaml(filepath string) (*node.Node, error) {
    if data, err := ioutil.ReadFile(filepath); err != nil {
        return nil, err
    } else {
        return yaml.Load(data)
    }
}

func ExportYaml(n *node.Node, filepath string) error {
    if data, err := yaml.Dump(n); err != nil {
        return err
    } else {
        return ioutil.WriteFile(filepath, data, 0644)
    }
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


