// Copyright NetApp Inc, 2021 All rights reserved

package rest

import (
	"encoding/json"
	"fmt"
	"github.com/bbrks/wrap/v2"
	"github.com/go-openapi/spec"
	tw "github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v3"
	"html"
	"io/ioutil"
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

// showModels prints the swagger definitions matching the specified api
// Most models have a reference and response. For example: volume has the following three definitions
// volume, volume_reference, volume_response
//
// In total, there are ~148 responses that return models
// cat ontapswagger.yaml | dasel -r yaml -w json | gron | rg 'json.definitions.(.*?)_response.properties.records.items..ref' | wc -l
// The api argument is tested against the schema name as well as a collection's schema when present
func showModels(a Args, ontapSwag ontap) {
	compile, err := regexp.Compile(a.Api)
	if err != nil {
		fmt.Printf("Error compiling api regex param=[%s] %v\n", a.Api, err)
		return
	}
	sortedKeys := sortSchema(ontapSwag.swagger.Definitions)
	seen := map[string]interface{}{}
	for _, respName := range sortedKeys {
		if compile.MatchString(respName) {
			printModelTable(respName, seen, ontapSwag)
		}
	}
}

func printModelTable(name string, seen map[string]interface{}, ontapSwag ontap) {
	_, ok := seen[name]
	if ok {
		return
	}
	def, ok := ontapSwag.swagger.Definitions[name]
	if ok {
		seen[name] = true
		// determine if this is a collection wrapper or a model
		// if it's a collection wrapper, unwrap the underlying model
		props := def.SchemaProps.Properties
		records, hasRecords := def.SchemaProps.Properties["records"]
		if hasRecords {
			props = records.Items.Schema.Properties
		}
		if props == nil {
			return
		}
		fmt.Printf("\n# Model: %s\n", name)

		table := newTable("name", "type", "description")
		table.SetColMinWidth(2, descriptionWidth)
		table.SetColWidth(descriptionWidth)

		for _, orderedItem := range props.ToOrderedSchemaItems() {
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
		refModel := schemaFromRef(args.schema)
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

func sortSchema(schema map[string]spec.Schema) []string {
	var keys []string
	for name := range schema {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	return keys
}

func showParams(a Args, ontapSwag ontap) {
	compile, err := regexp.Compile(a.Api)
	if err != nil {
		fmt.Printf("Error compiling api regex param=[%s] %v\n", a.Api, err)
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
					responseModel = schemaFromRef(*schema)
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

func schemaFromRef(schema spec.Schema) string {
	url := schema.Ref.GetURL()
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
			if p1.Required {
				return true
			}
			return false
		}
		return p1.Name < p2.Name
	}
}

// readSwagger reads the yaml swagger file from the local filesystem and converts it into a spec.Swagger struct
// and pulls out the collection apis
// This is done by:
//	1. unmarshalling the yaml swagger into a map
//  2. converting the map into a json []byte
//  3. converting the []byte into the spec.Swagger struct
// Step 2 is required because spec.Swagger (un)marshalling code is written with json tags instead of yaml tags
func readSwagger(args Args) (ontap, error) {
	contents, err := ioutil.ReadFile(args.SwaggerPath)
	if err != nil {
		fmt.Printf("error reading swagger file=[%s] err=%+v\n", args.SwaggerPath, err)
		return ontap{}, err
	}
	var ontapSwag spec.Swagger
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
	err = json.Unmarshal(b, &ontapSwag)
	if err != nil {
		fmt.Printf("error unmarshalling %s into swagger err=%+v\n", args.SwaggerPath, err)
		return ontap{}, err
	}

	// These are not tagged with an operations props id of "_collection_"
	// but should be included
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

func value(ptr *string, nilValue string) string {
	if ptr == nil {
		return nilValue
	}
	return *ptr
}
