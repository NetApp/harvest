package main

import (
	"goharvest2/pkg/conf"
	"os"
	"text/template"
)

type Poller struct {
	PollerName string
	Port int
}

type Test struct {
	Pollers []Poller
}

func main() {
	test := Test{}

	configPath := "/home/rahulg2/code/github/harvest/harvest.yml"
	conf.LoadHarvestConfig(configPath)

	for k,_ := range *conf.Config.Pollers {
		port , _ := conf.GetPrometheusExporterPorts(k)
		test.Pollers = append(test.Pollers, Poller{k, port})
	}

	t, err := template.New("docker-compose.tmpl").ParseFiles("docker-compose.tmpl")
	if err != nil {
		panic(err)
	}
	// Create the file
	f, err := os.Create("docker-compose.yaml")
	if err != nil {
		panic(err)
	}
	err = t.Execute(f, test)
	if err != nil {
		panic(err)
	}
}
