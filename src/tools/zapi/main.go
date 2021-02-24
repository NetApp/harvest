package main

import (
	"os"
	//"os/exec"
	"flag"
	"fmt"
	"strings"

	client "goharvest2/poller/api/zapi"

	"goharvest2/share/util"
	"goharvest2/share/config"
	//"goharvest2/share/tree"
	"goharvest2/share/tree/node"
	"goharvest2/poller/struct/set"
)

var ACTIONS = set.NewFrom([]string{"show", "add", "export"})
var ITEMS = set.NewFrom([]string{"system", "apis", "attrs", "data", "objects", "counters", "instances"})

var options *args
var connection *client.Client
var system *client.System

type args struct {
	Action string
	Item string
	Poller string
	Path string
	Query string
	Object string
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

	fmt.Printf("\n%s%-30s %-10s %-15s %-30s %20s %15s %20s%s\n\n", util.Bold, "name", "scalar", "unit", "properties", "base counter", "deprecated", "replacement", util.End)
}
func (c *counter) print() {
	fmt.Printf("%s%s%-30s%s %-10v %-15s %-30s %20s %15v %20s\n", util.Bold, util.Cyan, c.name, util.End, c.scalar, c.unit, strings.Join(c.properties, ", "), c.base, c.deprecated, c.replacement)
	if !c.scalar {
		fmt.Printf("%sarray labels 1D%s: %s%v%s\n", util.Pink, util.End, util.Grey, c.labels1, util.End)
		if len(c.labels2) > 0 {
			fmt.Printf("%sarray labels 2D%s: %s%v%s\n", util.Pink, util.End, util.Grey, c.labels2, util.End)
		}
	}
	fmt.Printf("%s%s%s\n", util.Yellow, c.info, util.End)
}

func (a *args) Print() {
	fmt.Printf("action = %s\n", a.Action)
	fmt.Printf("item   = %s\n", a.Item)
	fmt.Printf("path = %s\n", a.Path)
	fmt.Printf("poller = %s\n", a.Poller)
	fmt.Printf("query  = %s\n", a.Query)
	fmt.Printf("object = %s\n", a.Object)
}

func main() {

	var err error

	options = get_args()

	if !ACTIONS.Has(options.Action) {
		fmt.Printf("action should be one of: %v", ACTIONS.Slice())
		os.Exit(1)
	}

	if !ITEMS.Has(options.Item) {
		fmt.Printf("item should be one of: %v", ITEMS.Slice())
		os.Exit(1)
	}

	if err = connect(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	system = get_system()

	switch options.Item {

	case "system":
		//get_system()
        fmt.Println("Done")
	case "data":
		get_data()
	case "counters":
		get_counters()
    case "objects":
        get_objects()
    case "apis":
        get_apis()
    default:
        fmt.Printf("invalid item: %s\n", options.Item)
        os.Exit(1)
	}
}

func connect() error {

	//cwd, _ := os.Getwd()
	//fmt.Printf("cwd = %s\n", cwd)

	params, err := config.GetPoller(options.Path, "config.yaml", options.Poller)
	if err != nil {
		return err
	}

	fmt.Printf("connecting to [%s]... ", params.GetChildContentS("url"))

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

func get_system() *client.System {

	var err error

	fmt.Printf("fetching system info for poller [%s]\n", options.Poller)

	if system, err = connection.GetSystem(); err == nil {
		fmt.Println(system.String())
	} else {
		fmt.Println(err)
	}

    return system
}

func get_query() {
	fmt.Printf("fetching data for zapi query [%s]\n", options.Query)

	if err := connection.BuildRequestString(options.Query); err != nil {
		fmt.Println(err)
        return
	}

	results, err := connection.Invoke()

	if err != nil {
		fmt.Println(err)
	} else {
		results.Print(0)
	}
}

func get_data() {

    var request *node.Node

    request = node.NewXmlS(options.Query)

    if options.Object != "" {
        fmt.Printf("fetching raw data of zapiperf object [%s]\n", options.Object)

        if system.Clustered {
            request = node.NewXmlS("perf-object-get-instances")
            //request.NewChildS("max-records", "100")
            instances := request.NewChildS("instances", "")
            instances.NewChildS("instance", "*")
        } else {
            request = node.NewXmlS("perf-object-get-instances")
        }
        request.NewChildS("objectname", options.Object)
    } else {
        fmt.Printf("fetching raw data of zapi api [%s]\n", options.Query)
    }

    if err := connection.BuildRequest(request); err != nil {
        fmt.Println(err)
        return
    }

    results, err := connection.Invoke()
    if err != nil {
        fmt.Println(err)
    } else {
        results.Print(0)
    }
}

func get_apis() {

    if err := connection.BuildRequestString("system-api-list"); err != nil {
        fmt.Println(err)
        return
    }

    results, err := connection.Invoke()
    if err != nil {
        fmt.Println(err)
        return
    }

    apis := results.GetChildS("apis")
    if apis == nil {
        fmt.Println("Missing [apis] element in response")
        return
    }

    fmt.Printf("%s%s%-70s %s %20s %15s\n\n", util.Bold, util.Pink, "API", util.End, "LICENSE", "STREAM")
    for _, a := range apis.GetChildren() {
        fmt.Printf("%s%s%-70s %s %20s %15s\n", util.Bold, util.Pink, a.GetChildContentS("name"), util.End, a.GetChildContentS("license"), a.GetChildContentS("is-streaming"))
    }


}

func get_objects() {

    if err := connection.BuildRequestString("perf-object-list-info"); err != nil {
        fmt.Println(err)
        return
    }

    results, err := connection.Invoke()
    if err != nil {
        fmt.Println(err)
        return
    }

    objects := results.GetChildS("objects")
    if objects == nil {
        fmt.Println("Missing [objects] element in response")
        return
    }

    fmt.Printf("%s%s%-50s %s %s %-20s %15s %15s %15s %s\n\n", util.Bold, util.Blue, "OBJECT", util.End, util.Bold, "PREFERRED KEY", "PRIV-LVL", "DEPREC", "REPLC", util.End)
    for _, o := range objects.GetChildren() {
        fmt.Printf("\n%s%s%-50s %s %s %-20s %15s %15s %15s %s\n",
            util.Bold,
            util.Blue,
            o.GetChildContentS("name"),
            util.End,
            util.Bold,
            o.GetChildContentS("get-instances-preferred-counter"),
            o.GetChildContentS("privilege-level"),
            o.GetChildContentS("is-deprecated"),
            o.GetChildContentS("replaced-by"),
            util.End,
        )
        fmt.Printf("%s     %s%s\n", util.Grey, o.GetChildContentS("description"), util.End)
    }
}

func get_counters() {
	counters := make([]counter, 0)

	request := node.NewXmlS("perf-object-counter-list-info")

	request.NewChildS("objectname", options.Object)

	connection.BuildRequest(request)

	results, err := connection.Invoke()
	if err != nil {
		fmt.Println(err)
		return
	}

	counters_elem := results.GetChildS("counters")
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

			//elem.Print(0)

			if labels := elem.GetChildS("labels"); labels != nil {

				label_elems := labels.GetChildren()

				if len(label_elems) > 0 {
					c.labels1 = strings.Split( node.DecodeHtml( label_elems[0].GetContentS() ), ",")

					if len(label_elems) > 1 {
						c.labels2 = strings.Split( node.DecodeHtml( label_elems[1].GetContentS() ), ",")
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

    a := args{}

	flag.StringVar(&a.Path, "path", "/home/imandes0/GoCode/goharvest2", "harvest directory path")
	flag.StringVar(&a.Poller, "poller", "", "poller name")
	flag.StringVar(&a.Query, "query", "", "API query")
	flag.StringVar(&a.Object, "object", "", "API object")

	flag.Parse()

	if flag.NArg() < 2 {
		fmt.Printf("missing arguments (%d): action item\n", flag.NArg())
		os.Exit(1)
	}

	a.Action = flag.Arg(0)
	a.Item = flag.Arg(1)

    a.Print()

	return &a
}
