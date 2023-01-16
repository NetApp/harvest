// Copyright NetApp Inc, 2021 All rights reserved

package rest

import (
	"encoding/json"
	"fmt"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	validShowArgs = []string{"apis", "params", "models", "data"}
)

type check struct {
	isValid bool
}

var args = &Args{}

type Args struct {
	Item          string
	Poller        string
	API           string
	Endpoint      string
	Config        string
	SwaggerPath   string
	Fields        string
	Field         []string
	QueryField    string
	QueryValue    string
	DownloadAll   bool
	MaxRecords    string
	ForceDownload bool
	Verbose       bool
}

var Cmd = &cobra.Command{
	Use:   "rest",
	Short: "ONTAP Rest Utility",
	Long:  "ONTAP Rest Utility - Explore available ONTAP REST APIs",
}

var showCmd = &cobra.Command{
	Use:       "show",
	Short:     "item to show, one of: " + strings.Join(validShowArgs, ", "),
	Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	ValidArgs: validShowArgs,
	Run:       doShow,
}

func ReadOrDownloadSwagger(pName string) (string, error) {
	var (
		poller         *conf.Poller
		err            error
		addr           string
		shouldDownload = true
		swagTime       time.Time
	)

	if poller, addr, err = GetPollerAndAddr(pName); err != nil {
		return "", err
	}

	tmp := os.TempDir()
	swaggerPath := filepath.Join(tmp, addr+"-swagger.yaml")
	fileInfo, err := os.Stat(swaggerPath)

	if os.IsNotExist(err) {
		fmt.Printf("%s does not exist downloading\n", swaggerPath)
	} else if !args.ForceDownload {
		swagTime = fileInfo.ModTime()
		twoWeeksAgo := swagTime.Local().AddDate(0, 0, -14)
		if swagTime.Before(twoWeeksAgo) {
			fmt.Printf("%s is more than two weeks old, re-download", swaggerPath)
		} else {
			shouldDownload = false
		}
	}
	if shouldDownload {
		url := "https://" + addr + "/docs/api/swagger.yaml"
		bytesDownloaded, err := downloadSwagger(poller, swaggerPath, url, args.Verbose)
		if err != nil {
			fmt.Printf("error downloading swagger %s\n", err)
			return "", err
		}
		fmt.Printf("downloaded %d bytes from %s\n", bytesDownloaded, url)
	}
	fmt.Printf("Using downloaded file %s with timestamp %s\n", swaggerPath, swagTime)
	return swaggerPath, nil
}

func doShow(_ *cobra.Command, a []string) {
	c := validateArgs(a)
	if !c.isValid {
		return
	}
	err := conf.LoadHarvestConfig(args.Config)
	if err != nil {
		log.Fatal(err)
	}
	if args.SwaggerPath != "" {
		doSwagger(*args)
	} else {
		doCmd()
	}
}

func validateArgs(strings []string) check {
	// One of Poller or SwaggerPath are allowed, but not both
	if args.Poller != "" && args.SwaggerPath != "" {
		fmt.Printf("Both poller and swagger are set. Only one or the other can be set, not both\n")
		return check{isValid: false}
	}
	if len(strings) == 0 {
		args.Item = ""
	} else {
		args.Item = strings[0]
	}
	qfSet := args.QueryField != ""
	qvSet := args.QueryValue != ""
	if args.Item == "data" && (qfSet != qvSet) {
		fmt.Printf(`Both "query-fields" and "query-value" must be specified if either is specified.` + "\n")
		return check{isValid: false}
	}
	return check{isValid: true}
}

func doCmd() {
	switch args.Item {
	case "apis", "params", "models":
		swaggerPath, err := ReadOrDownloadSwagger(args.Poller)
		if err != nil {
			return // everything logged earlier
		}
		args.SwaggerPath = swaggerPath
		doSwagger(*args)
	case "data":
		doData()
	}
}

type Pagination struct {
	Records    []any `json:"records"`
	NumRecords int   `json:"num_records"`
	Links      *struct {
		Next struct {
			Href string `json:"href"`
		} `json:"next"`
	} `json:"_links,omitempty"`
}

type PerfRecord struct {
	Records   gjson.Result `json:"records"`
	Timestamp int64        `json:"time"`
}

func doData() {
	var (
		poller *conf.Poller
		err    error
		client *Client
	)

	if poller, _, err = GetPollerAndAddr(args.Poller); err != nil {
		return
	}

	timeout, _ := time.ParseDuration(DefaultTimeout)
	if client, err = New(*poller, timeout); err != nil {
		fmt.Printf("error creating new client %+v\n", err)
		os.Exit(1)
	}

	// strip leading slash
	args.API = strings.TrimPrefix(args.API, "/")

	var records []any
	href := BuildHref(args.API, args.Fields, args.Field, args.QueryField, args.QueryValue, args.MaxRecords, "", args.Endpoint)
	stderr("fetching href=[%s]\n", href)

	err = FetchForCli(client, href, &records, args.DownloadAll)
	if err != nil {
		stderr("error %+v\n", err)
		return
	}
	all := Pagination{
		Records:    records,
		NumRecords: len(records),
	}
	pretty, err := json.MarshalIndent(all, "", " ")
	if err != nil {
		stderr("error marshalling json %+v\n", err)
		return
	}
	fmt.Printf("%s\n", pretty)
}

func GetPollerAndAddr(pName string) (*conf.Poller, string, error) {
	var (
		poller *conf.Poller
		err    error
	)
	if poller, err = conf.PollerNamed(pName); err != nil {
		fmt.Printf("Poller named [%s] does not exist\n", pName)
		return nil, "", err
	}
	if poller.Addr == "" {
		fmt.Printf("Poller named [%s] does not have a valid addr=[]\n", pName)
		return nil, "", err
	}
	return poller, poller.Addr, nil
}

// FetchForCli used for CLI only
func FetchForCli(client *Client, href string, records *[]any, downloadAll bool) error {
	getRest, err := client.GetRest(href)
	if err != nil {
		return fmt.Errorf("error making request %w", err)
	}

	isNonIterRestCall := false
	value := gjson.GetBytes(getRest, "records")
	if value.String() == "" {
		isNonIterRestCall = true
	}

	if isNonIterRestCall {
		contentJSON := `{"records":[]}`
		response, err := sjson.SetRawBytes([]byte(contentJSON), "records.-1", getRest)
		if err != nil {
			return fmt.Errorf("error setting record %w", err)
		}
		var page Pagination
		err = json.Unmarshal(response, &page)
		if err != nil {
			return fmt.Errorf("error unmarshalling json %w", err)
		}
		*records = append(*records, page.Records...)
	} else {
		// extract returned records since paginated records need to be merged into a single list
		var page Pagination
		err := json.Unmarshal(getRest, &page)
		if err != nil {
			return fmt.Errorf("error unmarshalling json %w", err)
		}

		*records = append(*records, page.Records...)

		// If all results are desired and there is a next link, follow it
		if downloadAll && page.Links != nil {
			nextLink := page.Links.Next.Href
			if nextLink != "" {
				if nextLink == href {
					// nextLink is same as previous link, no progress is being made, exit
					return nil
				}
				err := FetchForCli(client, nextLink, records, downloadAll)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// Fetch collects all records
func Fetch(client *Client, href string) ([]gjson.Result, error) {
	var (
		records []gjson.Result
		result  []gjson.Result
		err     error
	)
	err = fetch(client, href, &records, true)
	if err != nil {
		return nil, err
	}
	for _, r := range records {
		result = append(result, r.Array()...)
	}
	return result, nil
}

// FetchLimited collects records as specified in URL
func FetchLimited(client *Client, href string) ([]gjson.Result, error) {
	var (
		records []gjson.Result
		result  []gjson.Result
		err     error
	)
	err = fetch(client, href, &records, false)
	if err != nil {
		return nil, err
	}
	for _, r := range records {
		result = append(result, r.Array()...)
	}
	return result, nil
}

func fetch(client *Client, href string, records *[]gjson.Result, downloadAll bool) error {
	getRest, err := client.GetRest(href)
	if err != nil {
		return fmt.Errorf("error making request %w", err)
	}

	isNonIterRestCall := false
	output := gjson.GetManyBytes(getRest, "records", "num_records", "_links.next.href")
	data := output[0]
	numRecords := output[1]
	next := output[2]
	if !data.Exists() {
		isNonIterRestCall = true
	}

	if isNonIterRestCall {
		contentJSON := `{"records":[]}`
		response, err := sjson.SetRawBytes([]byte(contentJSON), "records.-1", getRest)
		if err != nil {
			return fmt.Errorf("error setting record %w", err)
		}
		value := gjson.GetBytes(response, "records")
		*records = append(*records, value)
	} else {
		// extract returned records since paginated records need to be merged into a single lists
		if numRecords.Exists() && numRecords.Int() > 0 {
			*records = append(*records, data)
		}

		// If all results are desired and there is a next link, follow it
		if next.Exists() && downloadAll {
			nextLink := next.String()
			if nextLink != "" {
				if nextLink == href {
					// nextLink is same as previous link, no progress is being made, exit
					return nil
				}
				err := fetch(client, nextLink, records, downloadAll)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func FetchAnalytics(client *Client, href string, records *[]gjson.Result, analytics *gjson.Result, downloadAll bool) error {
	getRest, err := client.GetRest(href)
	if err != nil {
		return fmt.Errorf("error making request %w", err)
	}

	output := gjson.GetManyBytes(getRest, "records", "num_records", "_links.next.href", "analytics")
	data := output[0]
	numRecords := output[1]
	next := output[2]
	*analytics = output[3]

	// extract returned records since paginated records need to be merged into a single lists
	if numRecords.Exists() && numRecords.Int() > 0 {
		*records = append(*records, data)
	}

	// If all results are desired and there is a next link, follow it
	if next.Exists() && downloadAll {
		nextLink := next.String()
		if nextLink != "" {
			if nextLink == href {
				// nextLink is same as previous link, no progress is being made, exit
				return nil
			}
			err := fetch(client, nextLink, records, downloadAll)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// FetchRestPerfData This method is used in PerfRest collector. This method returns timestamp per batch
func FetchRestPerfData(client *Client, href string, perfRecords *[]PerfRecord) error {
	getRest, err := client.GetRest(href)
	if err != nil {
		return fmt.Errorf("error making request %w", err)
	}

	// extract returned records since paginated records need to be merged into a single list
	output := gjson.GetManyBytes(getRest, "records", "num_records", "_links.next.href")

	data := output[0]
	numRecords := output[1]
	next := output[2]

	if numRecords.Exists() && numRecords.Int() > 0 {
		p := PerfRecord{Records: data, Timestamp: time.Now().UnixNano()}
		*perfRecords = append(*perfRecords, p)
	}

	// If all results are desired and there is a next link, follow it
	if next.Exists() {
		nextLink := next.String()
		if nextLink != "" {
			if nextLink == href {
				// nextLink is same as previous link, no progress is being made, exit
				return nil
			}
			err := FetchRestPerfData(client, nextLink, perfRecords)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func stderr(format string, a ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format, a...)
}

func init() {
	configPath := conf.GetDefaultHarvestConfigPath()

	Cmd.AddCommand(showCmd)
	flags := Cmd.PersistentFlags()
	flags.StringVarP(&args.Poller, "poller", "p", "", "name of poller (cluster), as defined in your harvest config")
	flags.StringVarP(&args.SwaggerPath, "swagger", "s", "", "path to Swagger (OpenAPI) file to read from")
	flags.StringVar(&args.Config, "config", configPath, "harvest config file path")

	showFlags := showCmd.Flags()
	showFlags.StringVarP(&args.API, "api", "a", "", "REST API PATTERN to show")
	showFlags.StringVar(&args.Endpoint, "endpoint", "", "By default, /api is appended to passed argument in --api. Use --endpoint instead to pass absolute path of url")
	showFlags.BoolVar(&args.DownloadAll, "all", false, "Collect all records by walking pagination links")
	showFlags.BoolVarP(&args.Verbose, "verbose", "v", false, "Be verbose")
	showFlags.StringVarP(&args.MaxRecords, "max-records", "m", "", "Limit the number of records returned before providing pagination link")
	showFlags.BoolVar(&args.ForceDownload, "download", false, "Force download Swagger file instead of using local copy")
	showFlags.StringVarP(&args.Fields, "fields", "f", "*", "Fields to return in the response <field>[,...].")
	showFlags.StringArrayVar(&args.Field, "field", []string{}, "Query a field by value (can be specified multiple times.)\n"+
		`If the value contains query characters (*|,!<>..), it must be quoted to avoid their special meaning
    *         wildcard
    < > <= >= comparisons
    3..10     range
    !water    negation
    3|5       matching value in a list
    {} and "" escape special characters`)
	showFlags.StringVarP(&args.QueryField, "query-field", "q", "", "Search fields named <string>, matching rows where the value of the field selected by <string> matches <query-value>.\n"+
		"comma-delimited list of fields, or * to search across all fields.")
	showFlags.StringVarP(&args.QueryValue, "query-value", "u", "", "Pattern to search for in all fields specified by <query-fields>\n"+
		"same query characters as <field> apply (see above)")

	Cmd.SetUsageTemplate(Cmd.UsageTemplate() + `
Examples:
  harvest rest -p infinity show apis                                        Query cluster infinity for available APIs
  harvest rest -p infinity show params --api svm/svms                       Query cluster infinity for svm parameters. These query parameters are used
                                                                            to filter requests.
  harvest rest -p infinity show models --api svm/svms                       Query cluster infinity for svm models. These describe the REST response
                                                                            received when sending the svm/svms GET request.
  harvest rest -p infinity show data --api svm/svms --field "state=stopped" Query cluster infinity for stopped svms.

  harvest rest -p infinity show data --api storage/volumes \                Query cluster infinity for all volumes where  
      --field "space.physical_used_percent=>70" \                               physical_used_percent is > 70% and 
      --field "space.total_footprint=>400G" \                                   total_footprint is > 400G 
      --fields "name,svm,space"                                             The response should contain name, svm, and space attributes of matching volumes.

  harvest rest -p infinity show data --api storage/volumes \                Query cluster infinity for all volumes where the name of any volume or child
      --query-field "name" --query-value "io_load|scale"                         resource matches io_load or scale.
`)
}
