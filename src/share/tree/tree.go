package tree

import (
	"errors"
	"io/ioutil"
	"goharvest2/share/tree/node"
	"goharvest2/share/tree/xml"
	"goharvest2/share/tree/yaml"
	"goharvest2/share/tree/json"
)

func Print(n *node.Node) {
	n.Print(0)
}
func Import(format, filepath string) (*node.Node, error) {
	if data, err := ioutil.ReadFile(filepath); err != nil {
		return nil, err
	}
	
	switch format {
	case "yaml":
		return yaml.Load(data)
	case "xml":
		return xml.Load(data)
	case "json":
		return json.Load(data)
	default:
		return nil, errors.New("unknown format: " + format)
	}
}

func Export(n *node.Node, format, filepath string) error {

	var data []byte
	var err error

	switch format {
	case "yaml":
		data, err = yaml.Dump(node)
	case "xml":
		data, err = xml.Load(data)
	case "json":
		data, err = json.Load(data)
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
