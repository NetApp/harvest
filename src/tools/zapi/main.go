package main

import (
	"os"
	"os/exec"
	"flag"
	"fmt"
	"goharvest2/poller/structs/set"
	"goharvest2/poller/config"
    client "goharvest2/poller/apis/zapi"
)

var ACTIONS = set.NewFrom([]string{"show", "add", "export"})
var ITEMS = set.NewFrom([]string{"system", "apis", "attrs", "data", "objects", "counters", "instances"})

type args struct {
	action string
	item string
	poller string
	query string
	object string
}


var a *args
var connection client.Client
var system client.SystemInfo

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

	if !ACTIONS.Has(a.action) {
		fmt.Printf("action should be one of: %v", ACTIONS.Keys())
		os.Exit(1)
	}

	if !ITEMS.Has(a.item) {
		fmt.Printf("item should be one of: %v", ITEMS.Keys())
		os.Exit(1)
	}

	if err = connect(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	switch a.item {

	case "system":
		get_system()
	case "data":
		get_query()
	}
}

func connect() error {

	cwd, _ := os.Getwd()
	//fmt.Printf("cwd = %s\n", cwd)

	params, err := config.GetPoller(cwd, "config.yaml", a.poller)
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
	cmd.Run()
	// install certificate
	req := xml.New("security-certificate-install")
	req.CreateChild("cert-name", "harvest2-ca-cert")
	req.CreateChild("certificate", publickey)
	req.CreateChild("type", "client")

}

func get_system() {

	var err error

	fmt.Printf("fetching system info for poller [%s]\n", a.poller)

	if system, err = connection.GetSystemInfo(); err == nil {
		fmt.Println(system.String())
	} else {
		fmt.Println(err.Error())
	}
}

func get_query() {
	fmt.Printf("fetching data for zapi query [%s]\n", a.query)

	if err := connection.BuildRequestString(a.query); err != nil {
		fmt.Println(err)
	}
	
	results, err := connection.InvokeRequest()

	if err != nil {
		fmt.Println(err)
	} else {
		results.Print()
	}

}

func get_args() *args {

	a = &args{}

	flag.StringVar(&a.poller, "p", "", "poller")
	flag.StringVar(&a.query, "q", "", "query")
	flag.StringVar(&a.object, "o", "", "object")

	flag.Parse()

	if flag.NArg() != 2 {
		fmt.Println("missing arguments: action show")
		os.Exit(1)
	}

	a.action = flag.Arg(0)
	a.item = flag.Arg(1)

	return a
}