/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package tree

import (
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/tree/xml"
	"os"
)

func ImportYaml(filepath string) (*node.Node, error) {
	data, err := os.ReadFile(filepath)

	if err != nil {
		return nil, err
	}

	return LoadYaml(data)
}

func LoadYaml(data []byte) (*node.Node, error) {
	astFile, err := parser.ParseBytes(data, 0)
	if err != nil {
		return nil, err
	}
	// treat an empty file as an error
	if len(astFile.Docs) == 0 {
		return nil, errs.New(errs.ErrConfig, "template file is empty or does not exist")
	}

	r := node.New([]byte("Root"))
	consume(r, "", astFile.Docs[0].Body, false)
	return r, nil
}

func consume(r *node.Node, key string, y ast.Node, makeNewChild bool) {
	switch y.Type() { //nolint:exhaustive
	case ast.StringType:
		sn := y.(*ast.StringNode)
		r.NewChildS(key, sn.Value)
	case ast.IntegerType, ast.FloatType, ast.BoolType, ast.LiteralType, ast.NullType:
		r.NewChildS(key, y.String())
	case ast.MappingType:
		var s = r
		if key != "" || makeNewChild {
			s = r.NewChildS(key, "")
		}
		mn := y.(*ast.MappingNode)
		for _, child := range mn.Values {
			k := child.Key.String()
			// special case to handle incorrectly indented LabelAgent
			if k == "LabelAgent" && isScalar(child.Value) {
				s = r.NewChildS(k, "")
				continue
			}
			consume(s, k, child.Value, false)
		}
	case ast.DocumentType, ast.SequenceType, ast.AliasType:
		s := r.NewChildS(key, "")
		sn := y.(*ast.SequenceNode)
		for _, child := range sn.Values {
			makeNewChild := false
			if child.Type() == ast.MappingType {
				makeNewChild = key == "endpoints" || key == "events" || key == "matches"
			}
			consume(s, "", child, makeNewChild)
		}
	default:
		// ignore
	}
}

func isScalar(n ast.Node) bool {
	switch n.Type() { //nolint:exhaustive
	case ast.StringType, ast.IntegerType, ast.FloatType, ast.BoolType, ast.LiteralType, ast.NullType:
		return true
	default:
		return false
	}
}

func LoadXML(data []byte) (*node.Node, error) {
	return xml.Load(data)
}

func DumpXML(n *node.Node) ([]byte, error) {
	return xml.Dump(n)
}

func ImportXML(filepath string) (*node.Node, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	return LoadXML(data)
}
