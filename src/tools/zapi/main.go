package main

import (
	"os"
	//"os/exec"
	"flag"
	"fmt"
	"strings"

	client "goharvest2/poller/api/zapi"
	
	"goharvest2/poller/config"
	"goharvest2/poller/struct/xml"
	"goharvest2/poller/struct/set"
	"goharvest2/poller/share"
)

var ACTIONS = set.NewFrom([]string{"show", "add", "export"})
var ITEMS = set.NewFrom([]string{"system", "apis", "attrs", "data", "objects", "counters", "instances"})

var options *args
var connection *client.Client
var system *client.System

type args struct {
	action string
	item string
	poller string
	query string
	object string
}

type counter struct {
	name string
	info string
	scalar bool
	base string
	unit string
	properties []string
	labels1 []string
	labels2 []string
	privilege string
	deprecated bool
	replacement string
	key bool
}

func (c *counter) print_header() {

	fmt.Printf("\n%s%-30s %-10s %-15s %-30s %20s %15s %20s%s\n\n", share.Bold, "name", "scalar", "unit", "properties", "base counter", "deprecated", "replacement", share.End)
}
func (c *counter) print() {
	fmt.Printf("%s%s%-30s%s %-10v %-15s %-30s %20s %15v %20s\n", share.Bold, share.Cyan, c.name, share.End, c.scalar, c.unit, strings.Join(c.properties, ", "), c.base, c.deprecated, c.replacement)
	if !c.scalar {
		fmt.Printf("%sarray labels 1D%s: %s%v%s\n", share.Pink, share.End, share.Grey, c.labels1, share.End)
		if len(c.labels2) > 0 {
			fmt.Printf("%sarray labels 2D%s: %s%v%s\n", share.Pink, share.End, share.Grey, c.labels2, share.End)
		}
	}
}

func (a *args) print() {
	fmt.Printf("action = %s\n", a.action)
	fmt.Printf("item   = %s\n", a.item)
	fmt.Printf("poller = %s\n", a.poller)
	fmt.Printf("query  = %s\n", a.query)
	fmt.Printf("object = %s\n", a.object)
}

func main() {

	var err error

	get_args()
	//a.print()

	if !ACTIONS.Has(options.action) {
		fmt.Printf("action should be one of: %v", ACTIONS.Keys())
		os.Exit(1)
	}

	if !ITEMS.Has(options.item) {
		fmt.Printf("item should be one of: %v", ITEMS.Keys())
		os.Exit(1)
	}

	if err = connect(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	switch options.item {

	case "system":
		get_system()
	case "data":
		get_query()
	case "counters":
		get_counters()
	}
}

func connect() error {

	cwd, _ := os.Getwd()
	//fmt.Printf("cwd = %s\n", cwd)

	params, err := config.GetPoller(cwd, "config.yaml", options.poller)
	if err != nil {
		return err
	}

	fmt.Printf("connecting to [%s]... ", params.GetChildValue("url"))

	if connection, err = client.New(params); err != nil {
		return err
	}

	fmt.Println("OK")
	return nil
}

/*
func create_client_cert() {

	// check if openssl is installed

	// create harvest role

	req := xml.New("security-login-role-create")
	req.CreateChild("access-level", "readonly")
	req.CreateChild("command-directory-name", "DEFAULT")
	req.CreateChild("role-name", "harvest2-role")
	req.CreateChild("vserver", system.Name)

	// create harvest user

	req := xml.New("security-login-create")
	req.CreateChild("application", "ontapi")
	req.CreateChild("authentication-method", "cert")
	req.CreateChild("comment", "readonly user for harvest2")
	req.CreateChild("role-name", "harvest2-role")
	req.CreateChild("user-name", "harvest2-user")
	req.CreateChild("vserver", system.Name)

	// generate certificates
	cmd := exec.Command("openssl", "req", "-x509", "-nodes", "-days", "1095", "-newkey", "rsa:2048", "-keyout", "cert/jamaica.key", "-out", "cert/jamaica.pem", "-subj" "\"/CN=harvest2-user\"")
	if err := cmd.Run(); err != nil {
		
	}
	// install certificate
	req := xml.New("security-certificate-install")
	req.CreateChild("cert-name", "harvest2-ca-cert")
	req.CreateChild("certificate", publickey)
	req.CreateChild("type", "client")
}
*/

func get_system() {

	var err error

	fmt.Printf("fetching system info for poller [%s]\n", options.poller)

	if system, err = connection.GetSystem(); err == nil {
		fmt.Println(system.String())
	} else {
		fmt.Println(err)
	}
}

func get_query() {
	fmt.Printf("fetching data for zapi query [%s]\n", options.query)

	if err := connection.BuildRequestString(options.query); err != nil {
		fmt.Println(err)
	}
	
	results, err := connection.InvokeRequest()

	if err != nil {
		fmt.Println(err)
	} else {
		results.Print()
	}

}


func get_counters() {
	counters := make([]counter, 0)

	request := xml.New("perf-object-counter-list-info")

	request.CreateChild("objectname", options.object)
	
	connection.BuildRequest(request)

	results, err := connection.InvokeRequest()
	if err != nil {
		fmt.Println(err)
		return
	}

	counters_elem, _ := results.GetChild("counters")
	if counters_elem == nil {
		fmt.Println("no counters in response")
		return
	}

	for _, elem := range counters_elem.GetChildren() {

		name := elem.GetChildContentS("name")
		c := counter{ name : name }

		if elem.GetChildContentS("is-deprecated") == "true" {
			c.deprecated = true
			c.replacement = elem.GetChildContentS("replaced-by")
		}

		if elem.GetChildContentS("is-key") == "true" {
			c.key = true
		}

		c.info = elem.GetChildContentS("desc")
		c.base = elem.GetChildContentS("base-counter")
		c.unit = elem.GetChildContentS("unit")
		c.properties = strings.Split(elem.GetChildContentS("properties"), ",")
		c.privilege = elem.GetChildContentS("privelege-level")

		if elem.GetChildContentS("type") == "" {
			c.scalar = true
		} else {
			c.scalar = false

			elem.Print()

			if labels, _ := elem.GetChild("labels"); labels != nil {

				label_elems := labels.GetChildren()

				if len(label_elems) > 0 {
					c.labels1 = strings.Split( xml.DecodeHtml( label_elems[0].GetContentS() ), ",")

					if len(label_elems) > 1 {
						c.labels2 = strings.Split( xml.DecodeHtml( label_elems[1].GetContentS() ), ",")
					}
				}
			}
		}
		counters = append(counters, c)
	}

	if len(counters) > 0 {
		counters[0].print_header()
	}

	for _, c := range counters {
		c.print()
	}

	fmt.Println()
	fmt.Printf("showed metadata of %d counters\n", len(counters))
}

func get_args() *args {

	options = &args{}

	flag.StringVar(&options.poller, "p", "", "poller")
	flag.StringVar(&options.query, "q", "", "query")
	flag.StringVar(&options.object, "o", "", "object")

	flag.Parse()

	if flag.NArg() != 2 {
		fmt.Println("missing arguments: action show")
		os.Exit(1)
	}

	options.action = flag.Arg(0)
	options.item = flag.Arg(1)

	return options
}