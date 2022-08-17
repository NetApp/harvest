// Copyright NetApp Inc, 2021 All rights reserved

package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bbrks/wrap/v2"
	"github.com/go-openapi/spec"
	tw "github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v3"
	"html"
	"io"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type ontap struct {
	apis            []string
	swagger         spec.Swagger
	collectionsApis map[string]spec.PathItem
}

type propArgs struct {
	table     *tw.Table
	name      string
	schema    spec.Schema
	seen      map[string]interface{}
	ontapSwag ontap
}

func doSwagger(args Args) {
	ontapSwag, err := readSwagger(args)
	if err != nil {
		return
	}

	switch args.Item {
	case "apis":
		showApis(ontapSwag)
	case "params":
		showParams(args, ontapSwag)
	case "models":
		showModels(args, ontapSwag)
	}
}

const (
	descriptionWidth = 80
	nameWidth        = 30
	xNtapIntroduced  = "x-ntap-introduced"
)

type model struct {
	schema spec.Schema
	name   string
}

// showModels prints the swagger definitions matching the specified api
// Most models have a reference and response. For example: volume has the following three definitions
// volume, volume_reference, volume_response
//
// In total, there are ~148 responses that return models
// cat ontapswagger.yaml | dasel -r yaml -w json | gron | rg 'json.definitions.(.*?)_response.properties.records.items..ref' | wc -l
// The api argument is tested against the schema name as well as a collection's schema when present
func showModels(a Args, ontapSwag ontap) {
	compile, err := regexp.Compile(a.API)
	if err != nil {
		fmt.Printf("Error compiling api regex param=[%s] %v\n", a.API, err)
		return
	}
	seen := map[string]interface{}{}
	var collected []model
	for _, name := range sortApis(ontapSwag.collectionsApis) {
		pathItem := ontapSwag.collectionsApis[name]
		if compile.MatchString(name) {
			if pathItem.Get.OperationProps.Responses != nil {
				schema := (*pathItem.Get.OperationProps.Responses).ResponsesProps.StatusCodeResponses[200].ResponseProps.Schema
				if schema != nil {
					responseModel := schemaFromRef(schema.Ref.GetURL())
					if responseModel != "" {
						collectModels(responseModel, ontapSwag, &collected)
					}
				}
			}
		}
	}
	// Sort by api name
	sort.Slice(collected, func(i, j int) bool {
		return collected[i].name < collected[j].name
	})

	for _, m := range collected {
		fmt.Printf("\n# Model: %s\n", m.name)

		table := newTable("name", "type", "description")
		table.SetColMinWidth(2, descriptionWidth)
		table.SetColWidth(descriptionWidth)

		for _, orderedItem := range m.schema.Properties.ToOrderedSchemaItems() {
			args := propArgs{
				table:     table,
				name:      orderedItem.Name,
				schema:    orderedItem.Schema,
				seen:      seen,
				ontapSwag: ontapSwag,
			}
			printProperty(args)
		}
		table.Render()
	}
}

func collectModels(name string, ontapSwag ontap, collected *[]model) {
	def, modelName := walkRefs(ontapSwag, name)
	if def == nil {
		return
	}
	*collected = append(*collected, model{
		schema: *def,
		name:   modelName,
	})
}

func walkRefs(ontapSwag ontap, name string) (*spec.Schema, string) {
	def, ok := ontapSwag.swagger.Definitions[name]
	if !ok {
		return nil, ""
	}
	records, hasRecords := def.SchemaProps.Properties["records"]
	if hasRecords {
		schemaProps := records.Items.Schema.SchemaProps
		if schemaProps.Type == nil {
			ref := schemaFromRef(schemaProps.Ref.GetURL())
			if len(ref) > 0 {
				schema, ok := ontapSwag.swagger.Definitions[ref]
				if ok {
					return &schema, ref
				}
			}
		} else {
			ref := schemaFromRef(records.Items.Schema.Ref.GetURL())
			if len(ref) == 0 {
				ref = name
			}
			return records.Items.Schema, ref
		}
	}
	return nil, ""
}

func newTable(headers ...string) *tw.Table {
	table := tw.NewWriter(os.Stdout)
	table.SetBorder(false)
	table.SetRowLine(true)
	table.SetHeader(headers)
	return table
}

func printProperty(args propArgs) {
	// Ignore this metadata
	if strings.HasSuffix(args.name, "_links") || strings.HasSuffix(args.name, " self") {
		return
	}
	if args.schema.Type == nil {
		// check if this is a reference to another schema
		refModel := schemaFromRef(args.schema.Ref.GetURL())
		if len(refModel) > 0 {
			_, seen := args.seen[refModel]
			if !seen {
				refSchema, ok := args.ontapSwag.swagger.Definitions[refModel]
				if ok {
					args.schema = refSchema
					if refModel != args.name {
						// use space as a separator to help wordwrap
						args.name = args.name + " " + refModel
					}
				}
			}
		}
	}
	cleanDesc := html.UnescapeString(args.schema.Description)
	w := wrap.NewWrapper()
	w.StripTrailingNewline = true
	text := w.Wrap(cleanDesc, descriptionWidth)
	kind := strings.Join(args.schema.SchemaProps.Type, ", ")

	args.table.Append([]string{w.Wrap(args.name, nameWidth), kind, text})
	for _, orderedItem := range args.schema.SchemaProps.Properties.ToOrderedSchemaItems() {
		argsChild := propArgs{
			table:     args.table,
			name:      args.name + " " + orderedItem.Name,
			schema:    orderedItem.Schema,
			seen:      args.seen,
			ontapSwag: args.ontapSwag,
		}
		printProperty(argsChild)
	}
}

func sortApis(schema map[string]spec.PathItem) []string {
	var keys []string
	for name := range schema {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	return keys
}

func showParams(a Args, ontapSwag ontap) {
	compile, err := regexp.Compile(a.API)
	if err != nil {
		fmt.Printf("Error compiling api regex param=[%s] %v\n", a.API, err)
		return
	}
	w := wrap.NewWrapper()
	w.StripTrailingNewline = true
	for _, name := range sortApis(ontapSwag.collectionsApis) {
		pathItem := ontapSwag.collectionsApis[name]
		table := newTable("  ", "type", "field", "description")
		table.SetColMinWidth(3, descriptionWidth)

		if compile.MatchString(name) {
			sort.SliceStable(pathItem.Get.OperationProps.Parameters, sortParams(pathItem))
			responseModel := ""
			if pathItem.Get.OperationProps.Responses != nil {
				schema := (*pathItem.Get.OperationProps.Responses).ResponsesProps.StatusCodeResponses[200].ResponseProps.Schema
				if schema != nil {
					responseModel = schemaFromRef(schema.Ref.GetURL())
				}
			}

			fmt.Printf("\n# %s\n", name)
			fmt.Printf("Response model: %s\n", responseModel)
			if len(pathItem.Get.OperationProps.Tags) > 0 {
				fmt.Printf("Tags: %s\n", pathItem.Get.OperationProps.Tags)
			}
			for _, param := range pathItem.Get.OperationProps.Parameters {
				description := param.Description
				cleanDesc := html.UnescapeString(description)
				text := w.Wrap(cleanDesc, descriptionWidth)

				if len(param.Name) > 0 {
					required := " "
					if param.ParamProps.Required {
						required = "*"
					}
					table.Append([]string{required, param.SimpleSchema.Type, param.Name, text})
				}
			}
			table.Render()
		}
	}
}

func schemaFromRef(url *url.URL) string {
	responseModel := ""
	if url != nil {
		responseModel = url.Fragment
		lastSlash := strings.LastIndex(responseModel, "/")
		if lastSlash > -1 && lastSlash+1 < len(responseModel) {
			responseModel = responseModel[lastSlash+1:]
		}
	}
	return responseModel
}

func sortParams(pathItem spec.PathItem) func(i int, j int) bool {
	return func(i, j int) bool {
		p1 := pathItem.Get.OperationProps.Parameters[i]
		p2 := pathItem.Get.OperationProps.Parameters[j]

		if p1.Required != p2.Required {
			return p1.Required
		}
		return p1.Name < p2.Name
	}
}

// readSwagger reads the yaml swagger file from the local filesystem and converts it into a spec.Swagger struct
// and pulls out the collection apis
// This is done by:
//  1. unmarshalling the yaml swagger into a map
//  2. converting the map into a json []byte
//  3. converting the []byte into the spec.Swagger struct
//
// Step 2 is required because spec.Swagger (un)marshalling code is written with json tags instead of yaml tags
func readSwagger(args Args) (ontap, error) {
	contents, err := os.ReadFile(args.SwaggerPath)
	if err != nil {
		fmt.Printf("error reading swagger file=[%s] err=%+v\n", args.SwaggerPath, err)
		return ontap{}, err
	}
	// read ONTAP swagger yaml and convert to JSON since swagger only has JSON unmarshalling
	node := make(map[string]interface{})
	err = yaml.Unmarshal(contents, node)
	if err != nil {
		fmt.Printf("error unmarshalling swagger file=[%s] err=%+v\n", args.SwaggerPath, err)
		return ontap{}, err
	}

	b, err := json.Marshal(node)
	if err != nil {
		fmt.Printf("error marshalling swagger file=[%s] to json err=%+v\n", args.SwaggerPath, err)
		return ontap{}, err
	}
	var ontapSwag spec.Swagger
	err = json.Unmarshal(b, &ontapSwag)
	if err != nil {
		// attempt to fix the swagger and try again
		swag, err2 := fixSwagger(args.SwaggerPath, b)
		if err2 != nil {
			fmt.Printf("error unmarshalling %s into swagger err=%+v\n", args.SwaggerPath, err)
			fmt.Printf("error attemping to fix swagger err=%+v\n", err2)
			return ontap{}, err2
		}
		ontapSwag = swag
	}

	// Include the following, even though they are not tagged with an operations props id of "_collection_"
	missingToAdd := map[string]interface{}{
		"/cluster/nodes": true,
	}
	collectionsApis := make(map[string]spec.PathItem)
	for name, path := range ontapSwag.Paths.Paths {
		if path.PathItemProps.Get != nil {
			if strings.Contains(path.PathItemProps.Get.OperationProps.ID, "_collection_") {
				collectionsApis[name] = path
			} else {
				_, ok := missingToAdd[name]
				if ok {
					collectionsApis[name] = path
				}
			}
		}
	}
	names := sortApis(collectionsApis)
	return ontap{apis: names, swagger: ontapSwag, collectionsApis: collectionsApis}, nil
}

// fixSwagger attempts to fix some known errors in ONTAP's swagger yaml file
// that cause go-openapi/spec to fail. The errors are found in visitNode
func fixSwagger(path string, b []byte) (spec.Swagger, error) {
	var nodes map[string]interface{}
	err := json.Unmarshal(b, &nodes)
	if err != nil {
		fmt.Printf("error unmarshalling %s into map err=%+v\n", args.SwaggerPath, err)
		return spec.Swagger{}, err
	}

	// visit every node correcting known issues then try again
	visitNode(nodes)
	var ontapSwag spec.Swagger
	nb, err := json.Marshal(nodes)
	if err != nil {
		fmt.Printf("error marshalling adjusted swagger file=[%s] to json err=%+v\n", args.SwaggerPath, err)
		return spec.Swagger{}, err
	}
	err = json.Unmarshal(nb, &ontapSwag)
	if err == nil {
		// the fixes worked, write out the changes
		out, err := os.Create(path)
		if err != nil {
			return spec.Swagger{}, fmt.Errorf("unable to create %s to save swagger.yaml", path)
		}
		defer func(out *os.File) { _ = out.Close() }(out)
		_, err = io.Copy(out, bytes.NewReader(nb))
		if err != nil {
			return spec.Swagger{}, fmt.Errorf("error while saving mutated swagger to %s err=%w", path, err)
		}
	}
	return ontapSwag, err
}

// visitNode corrects type errors in ONTAP's swagger.yaml file
// the properties listed in the case statement below should have a type
// of int instead of string - correct that so the yaml is valid and parses
func visitNode(v interface{}) {
	switch vv := v.(type) {
	case map[string]interface{}:
		for k, val := range vv {
			switch k {
			case "maxLength", "minLength", "maximum", "minimum", "minItems", "maxItems":
				if s, ok := val.(string); ok {
					num, err := strconv.Atoi(s)
					if err == nil {
						vv[k] = num
					}
				}
			case "readOnly":
				// seen on 9.7P9
				// readOnly should be a boolean. In some versions of ONTAP, a number was used as the value
				// instead of a boolean type. e.g. "readOnly": 1
				// detect that and fix it here
				if _, ok := val.(bool); !ok {
					if s, ok2 := val.(float64); ok2 {
						if s > 0 {
							vv[k] = true
						} else {
							vv[k] = false
						}
					}
				}
			}
			visitNode(val)
		}
	case []interface{}:
		for _, val := range vv {
			visitNode(val)
		}
	}
}

func showApis(ontapSwag ontap) {
	fmt.Printf("# of collection apis %d\n", len(ontapSwag.apis))
	table := tw.NewWriter(os.Stdout)
	table.SetBorder(false)
	table.SetHeader([]string{"API", "Params", "Ver"})
	for _, name := range ontapSwag.apis {
		pathItem := ontapSwag.collectionsApis[name]
		count := strconv.Itoa(len(pathItem.Get.Parameters))
		version := ""
		if introduced, ok := pathItem.Get.VendorExtensible.Extensions[xNtapIntroduced]; ok {
			version = fmt.Sprintf("%s", introduced)
		}
		table.Append([]string{name, count, version})
	}
	table.Render()
}
