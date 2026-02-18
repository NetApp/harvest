// Copyright NetApp Inc, 2021 All rights reserved

package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/netapp/harvest/v2/pkg/auth"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/requests"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"github.com/netapp/harvest/v2/third_party/tidwall/sjson"
	"github.com/spf13/cobra"
	"log"
	"log/slog"
	"net/url"
	"os"
	"os/signal"
	"strconv"
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
	Item        string
	Poller      string
	API         string
	Endpoint    string
	Config      string
	Fields      string
	Field       []string
	QueryField  string
	QueryValue  string
	DownloadAll bool
	MaxRecords  string
	Timeout     string
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

func OntapRestAPIHref(pName string) (string, error) {
	var (
		err  error
		addr string
	)

	if _, addr, err = GetPollerAndAddr(pName); err != nil {
		return "", err
	}

	return "https://" + addr + "/docs/api/", nil
}

func doShow(_ *cobra.Command, a []string) {
	c := validateArgs(a)
	if !c.isValid {
		return
	}
	_, err := conf.LoadHarvestConfig(args.Config)
	if err != nil {
		log.Fatal(err)
	}
	doCmd()
}

func validateArgs(slice []string) check {
	if len(slice) == 0 {
		args.Item = ""
	} else {
		args.Item = slice[0]
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
		ontapRestAPI, err := OntapRestAPIHref(args.Poller)
		if err != nil {
			fmt.Printf("error %+v\n", err)
			return
		}
		fmt.Printf("Find the ONTAP REST API reference for %s here: %s\n", args.Poller, ontapRestAPI)
	case "data":
		doData()
	}
}

func fetchData(poller *conf.Poller, timeout time.Duration) (*Results, error) {
	var (
		err    error
		client *Client
	)

	if client, err = New(poller, timeout, auth.NewCredentials(poller, slog.Default())); err != nil {
		return nil, fmt.Errorf("poller=%s %w", poller.Name, err)
	}

	// Init is called to get the cluster version
	err = client.Init(1, conf.Remote{})
	if err != nil {
		if re, ok := errors.AsType[*errs.RestError](err); ok {
			return nil, fmt.Errorf("poller=%s statusCode=%d", poller.Name, re.StatusCode)
		}
		return nil, fmt.Errorf("poller=%s %w", poller.Name, err)
	}

	// strip leading slash
	args.API = strings.TrimPrefix(args.API, "/")

	now := time.Now()
	var records []any
	var curls []string

	hrefBuilder := NewHrefBuilder().
		APIPath(args.API).
		Fields(strings.Split(args.Fields, ",")).
		Filter(args.Field).
		QueryFields(args.QueryField).
		QueryValue(args.QueryValue)

	if args.MaxRecords != "" {
		_, err := strconv.Atoi(args.MaxRecords)
		if err != nil {
			return nil, fmt.Errorf("--max-records should be numeric %s", args.MaxRecords)
		}
		hrefBuilder.MaxRecords(args.MaxRecords)
	}

	href := hrefBuilder.Build()

	err = FetchForCli(client, href, &records, args.DownloadAll, &curls)
	if err != nil {
		return nil, fmt.Errorf("poller=%s %w", poller.Name, err)
	}
	for _, curl := range curls {
		stderr("%s # %s\n", curl, poller.Name)
	}
	results := &Results{
		Poller:         poller.Name,
		Addr:           poller.Addr,
		API:            args.API,
		Version:        client.remote.Version,
		ClusterName:    client.remote.Name,
		Records:        records,
		NumRecords:     len(records),
		PollDurationMs: time.Since(now).Milliseconds(),
	}
	if len(records) == 0 {
		results.Records = []any{}
	}
	return results, nil
}

type Results struct {
	Poller         string `json:"poller,omitempty"`
	Addr           string `json:"addr,omitempty"`
	API            string `json:"api,omitempty"`
	Version        string `json:"version,omitempty"`
	ClusterName    string `json:"cluster_name,omitempty"`
	Records        []any  `json:"records"`
	NumRecords     int    `json:"num_records"`
	PollDurationMs int64  `json:"poll_ms"`
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
		err     error
		results []*Results
		timeout time.Duration
	)

	timeout, err = time.ParseDuration(args.Timeout)
	if err != nil {
		stderr("Unable to parse timeout=%s using default %s\n", args.Timeout, DefaultTimeout)
		timeout, _ = time.ParseDuration(DefaultTimeout)
	}

	resultChan := make(chan *Results)
	errChan := make(chan error)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	pollers := make([]string, 0)
	if args.Poller == "*" {
		pollers = append(pollers, conf.Config.PollersOrdered...)
	} else {
		pollers = append(pollers, args.Poller)
	}

	// Prime the credential cache before forking goroutines
	for _, pollerName := range pollers {
		p, _, err := GetPollerAndAddr(pollerName)
		if err != nil {
			stderr("failed to get poller %s err: %+v\n", pollers[0], err)
			continue
		}
		cred := auth.NewCredentials(p, slog.Default())
		_, _ = cred.GetPollerAuth()
	}

	for _, pollerName := range pollers {
		go func(pollerName string) {
			var (
				poller *conf.Poller
			)
			if poller, _, err = GetPollerAndAddr(pollerName); err != nil {
				errChan <- err
				return
			}
			data, err := fetchData(poller, timeout)
			if err != nil {
				errChan <- err
				return
			}
			resultChan <- data
		}(pollerName)
	}

outer:
	for range pollers {
		select {
		case r := <-resultChan:
			results = append(results, r)
		case err := <-errChan:
			stderr("failed to fetch data err: %+v\n", err)
		case <-sigChan:
			break outer
		}
	}

	if results != nil {
		pretty, err := json.MarshalIndent(results, "", " ")
		if err != nil {
			stderr("error marshalling json %+v\n", err)
			return
		}
		fmt.Printf("%s\n", pretty)
	}
}

func GetPollerAndAddr(pName string) (*conf.Poller, string, error) {
	var (
		poller *conf.Poller
		err    error
	)
	if poller, err = conf.PollerNamed(pName); err != nil {
		return nil, "", fmt.Errorf("poller=%s does not exist, err: %w", pName, err)
	}
	if poller.Addr == "" {
		return nil, "", fmt.Errorf("poller=%s has blank addr", pName)
	}
	return poller, poller.Addr, nil
}

// FetchForCli used for CLI only
func FetchForCli(client *Client, href string, records *[]any, downloadAll bool, curls *[]string) error {

	var prevLink string
	nextLink := href

	for {
		getRest, err := client.GetRest(nextLink)
		if err != nil {
			return fmt.Errorf("error making request %w", err)
		}

		pollerAuth, err := client.auth.GetPollerAuth()
		if err != nil {
			return err
		}
		*curls = append(*curls, fmt.Sprintf("curl --user %s --insecure '%s%s'", pollerAuth.Username, client.baseURL, nextLink))

		isNonIterRestCall := false
		value := gjson.GetBytes(getRest, "records")
		if value.ClonedString() == "" {
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
			break
		}

		// The pagination struct is used to pretty print the JSON output
		var page Pagination
		err = json.Unmarshal(getRest, &page)
		if err != nil {
			return fmt.Errorf("error unmarshalling json %w", err)
		}

		*records = append(*records, page.Records...)

		// If all results are desired and there is a next link, follow it
		next := ""
		if page.Links != nil {
			next = page.Links.Next.Href
		}

		prevLink = nextLink
		nextLink = next
		// strip leading slash
		nextLink = strings.TrimPrefix(nextLink, "/")

		if nextLink == "" || nextLink == prevLink || !downloadAll {
			// no nextLink, nextLink is the same as the previous link, or not all records are desired, exit
			break
		}
	}

	return nil
}

// FetchAll collects all records.
// If you want to limit the number of records returned, use FetchSome.
func FetchAll(client *Client, href string, headers ...map[string]string) ([]gjson.Result, error) {
	var (
		records []gjson.Result
		result  []gjson.Result
		err     error
	)

	err = fetchAll(client, href, &records, headers...)
	if err != nil {
		return nil, err
	}
	for _, r := range records {
		result = append(result, r.Array()...)
	}
	return result, nil
}

func FetchAnalytics(client *Client, href string) ([]gjson.Result, gjson.Result, error) {
	var (
		records   []gjson.Result
		analytics = &gjson.Result{}
		result    []gjson.Result
		err       error
	)
	downloadAll := true
	maxRecords := 0
	if strings.Contains(href, "max_records") {
		mr, err := requests.GetQueryParam(href, "max_records")
		if err != nil {
			return []gjson.Result{}, gjson.Result{}, err
		}
		if mr != "" {
			mri, err := strconv.Atoi(mr)
			if err != nil {
				return []gjson.Result{}, gjson.Result{}, err
			}
			maxRecords = mri
		}
		downloadAll = maxRecords == 0
	}
	err = fetchAnalytics(client, href, &records, analytics, downloadAll, int64(maxRecords))
	if err != nil {
		return nil, gjson.Result{}, err
	}
	for _, r := range records {
		result = append(result, r.Array()...)
	}

	if len(result) == 0 {
		return []gjson.Result{}, gjson.Result{}, nil
	}

	return result, *analytics, nil
}

func FetchRestPerfDataStream(client *Client, href string, processBatch func([]PerfRecord) error, headers ...map[string]string) error {
	var prevLink string
	nextLink := href
	recordsFound := false
	for {
		response, err := client.GetRest(nextLink, headers...)
		if err != nil {
			return fmt.Errorf("error making request %w", err)
		}

		// extract returned records since paginated records need to be merged into a single list
		output := gjson.ParseBytes(response)
		data := output.Get("records")
		numRecords := output.Get("num_records")
		next := output.Get("_links.next.href")

		if numRecords.Int() > 0 {
			recordsFound = true
			p := PerfRecord{Records: data, Timestamp: time.Now().UnixNano()}
			if err := processBatch([]PerfRecord{p}); err != nil {
				return err
			}
		}

		prevLink = nextLink
		nextLink = next.ClonedString()

		if nextLink == "" || nextLink == prevLink {
			// no nextLink or nextLink is the same as the previous link, no progress is being made, exit
			break
		}
	}
	if !recordsFound {
		return errs.New(errs.ErrNoInstance, "no instances found")
	}

	return nil
}

func FetchPost(client *Client, endpoint string, body []byte, headers ...map[string]string) ([]gjson.Result, error) {
	response, err := client.PostRest(endpoint, body, headers...)
	if err != nil {
		return nil, fmt.Errorf("error in POST request: %w", err)
	}

	output := gjson.ParseBytes(response)
	data := output.Get("output")
	var results []gjson.Result

	if data.Exists() {
		results = data.Array()
	} else {
		results = []gjson.Result{}
	}

	return results, nil
}

func FetchAllStream(client *Client, href string, processBatch func([]gjson.Result, int64) error, headers ...map[string]string) error {
	var prevLink string
	nextLink := href
	recordsFound := false

	for {
		var records []gjson.Result
		response, err := client.GetRest(nextLink, headers...)
		if err != nil {
			return fmt.Errorf("error making request %w", err)
		}

		output := gjson.ParseBytes(response)
		data := output.Get("records")
		numRecords := output.Get("num_records")
		next := output.Get("_links.next.href")

		if data.Exists() {
			if numRecords.Int() > 0 {
				recordsFound = true
				// Process the current batch of records
				if err := processBatch(data.Array(), time.Now().UnixNano()/collector.BILLION); err != nil {
					return err
				}
			}

			prevLink = nextLink
			// If there is a next link, follow it
			nextLink = next.ClonedString()
			if nextLink == "" || nextLink == prevLink {
				// no nextLink or nextLink is the same as the previous link, no progress is being made, exit
				break
			}
		} else {
			contentJSON := `{"records":[]}`
			response, err := sjson.SetRawBytes([]byte(contentJSON), "records.-1", response)
			if err != nil {
				return fmt.Errorf("error setting record %w", err)
			}
			value := gjson.GetBytes(response, "records")
			records = append(records, value.Array()...)
			if len(records) > 0 {
				recordsFound = true
				// Process the current batch of records
				if err := processBatch(records, time.Now().UnixNano()/collector.BILLION); err != nil {
					return err
				}
			}
			break
		}
	}
	if !recordsFound {
		return errs.New(errs.ErrNoInstance, "no instances found")
	}

	return nil
}

func fetchAll(client *Client, href string, records *[]gjson.Result, headers ...map[string]string) error {

	var prevLink string
	nextLink := href

	for {
		response, err := client.GetRest(nextLink, headers...)
		if err != nil {
			return fmt.Errorf("error making request %w", err)
		}

		output := gjson.ParseBytes(response)
		data := output.Get("records")
		numRecords := output.Get("num_records")
		next := output.Get("_links.next.href")

		if data.Exists() {
			// extract returned records since paginated records need to be merged into a single lists
			if numRecords.Int() > 0 {
				*records = append(*records, data)
			}

			prevLink = nextLink
			// If there is a next link, follow it
			nextLink = next.ClonedString()
			if nextLink == "" || nextLink == prevLink {
				// no nextLink or nextLink is the same as the previous link, no progress is being made, exit
				break
			}
		} else {
			contentJSON := `{"records":[]}`
			response, err := sjson.SetRawBytes([]byte(contentJSON), "records.-1", response)
			if err != nil {
				return fmt.Errorf("error setting record %w", err)
			}
			value := gjson.GetBytes(response, "records")
			*records = append(*records, value)
			break
		}
	}

	return nil
}

// FetchSome collects at most recordsWanted records, following pagination links as needed.
// Use batchSize to limit the number of records returned in a single response.
// If recordsWanted is -1, all records are collected.
func FetchSome(client *Client, href string, recordsWanted int, batchSize string) ([]gjson.Result, error) {
	var (
		records []gjson.Result
		result  []gjson.Result
		err     error
	)

	// Set max_records to batchSize to limit the number of records returned in a single response.
	// If recordsWanted is < batchSize, set batchSize to recordsWanted.
	batch, _ := strconv.Atoi(batchSize)
	if recordsWanted < batch {
		batch = recordsWanted
	}

	u, err := url.Parse(href)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("max_records", strconv.Itoa(batch))
	encoded := q.Encode()
	u.RawQuery = encoded
	href = u.String()

	err = fetchLimit(client, href, &records, recordsWanted)
	if err != nil {
		return nil, err
	}
	for _, r := range records {
		result = append(result, r.Array()...)
	}
	return result, nil
}

func fetchLimit(client *Client, href string, records *[]gjson.Result, recordsWanted int) error {

	var prevLink string
	nextLink := href

	for {
		getRest, err := client.GetRest(nextLink)
		if err != nil {
			return fmt.Errorf("error making request %w", err)
		}

		output := gjson.ParseBytes(getRest)
		data := output.Get("records")
		numRecords := output.Get("num_records")
		next := output.Get("_links.next.href")

		if data.Exists() {
			// extract returned records since paginated records need to be merged into a single lists
			if numRecords.Int() > 0 {
				*records = append(*records, data)

				if recordsWanted != -1 {
					recordsWanted -= int(numRecords.Int())
					if recordsWanted <= 0 {
						return nil
					}
				}
			}

			prevLink = nextLink
			nextLink = next.ClonedString()

			if nextLink == "" || nextLink == prevLink {
				// no nextLink or nextLink is the same as the previous link, no progress is being made, exit
				break
			}
			// Follow the next link
		} else {
			contentJSON := `{"records":[]}`
			response, err := sjson.SetRawBytes([]byte(contentJSON), "records.-1", getRest)
			if err != nil {
				return fmt.Errorf("error setting record %w", err)
			}
			value := gjson.GetBytes(response, "records")
			*records = append(*records, value)
			break
		}
	}

	return nil
}

func fetchAnalytics(client *Client, href string, records *[]gjson.Result, analytics *gjson.Result, downloadAll bool, maxRecords int64) error {

	var prevLink string
	nextLink := href

	for {
		getRest, err := client.GetRest(nextLink)
		if err != nil {
			return fmt.Errorf("error making request %w", err)
		}

		output := gjson.ParseBytes(getRest)
		data := output.Get("records")
		numRecords := output.Get("num_records")
		next := output.Get("_links.next.href")
		*analytics = output.Get("analytics")

		// extract returned records since paginated records need to be merged into a single lists
		if numRecords.Int() > 0 {
			*records = append(*records, data)
			if !downloadAll {
				maxRecords -= numRecords.Int()
				if maxRecords <= 0 {
					return nil
				}
			}
		}

		prevLink = nextLink
		nextLink = next.ClonedString()

		if nextLink == "" || nextLink == prevLink || !downloadAll {
			// no nextLink, nextLink is the same as the previous link, or not all records are desired, exit
			break
		}
	}

	return nil
}

func stderr(format string, a ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format, a...)
}

func init() {
	configPath := conf.Path(conf.HarvestYML)

	Cmd.AddCommand(showCmd)
	flags := Cmd.PersistentFlags()
	flags.StringVarP(&args.Poller, "poller", "p", "", "Name of poller (cluster), as defined in your harvest config. * for all pollers")
	flags.StringVar(&args.Config, "config", configPath, "Harvest config file path")
	flags.StringVarP(&args.Timeout, "timeout", "t", DefaultTimeout, "Duration to wait before giving up")

	showFlags := showCmd.Flags()
	showFlags.StringVarP(&args.API, "api", "a", "", "REST API PATTERN to show")
	showFlags.BoolVar(&args.DownloadAll, "all", false, "Collect all records by walking pagination links")
	showFlags.StringVarP(&args.MaxRecords, "max-records", "m", "", "Limit the number of records returned before providing pagination link")
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
  # Print cluster infinity's' ONTAP REST API Online Reference 
  bin/harvest rest -p infinity show apis
  
  # Query cluster infinity for stopped svms.
  bin/harvest rest -p infinity show data --api svm/svms --field "state=stopped"

  # Query cluster infinity for all volumes where physical_used_percent is > 70% and total_footprint is >= 400G. The response should contain name, svm, and space attributes of matching volumes.  	
  bin/harvest rest -p infinity show data --api storage/volumes --field "space.physical_used_percent=>70" --field "space.total_footprint=>=400G" --fields "name,svm,space"

  # Query cluster infinity for all volumes where the name of any volume or child resource matches io_load or scale.
  bin/harvest rest -p infinity show data --api storage/volumes --query-field "name" --query-value "io_load|scale"

  # Query all clusters, in your harvest.yml file, for all qos policies. Pipe the results to jq, and print as CSV.	
  bin/harvest rest -p '*' show data --api storage/qos/policies | jq -r '.[] | [.poller, .addr, .num_records, .version, .cluster_name, .poll_ms, .api] |  @csv' | column -ts,
`)
}
