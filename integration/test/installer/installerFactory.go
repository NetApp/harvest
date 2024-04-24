package installer

import (
	"errors"
)

const (
	RHEL              = "rpm"
	NATIVE            = "tar"
	HarvestConfigFile = "harvest.yml"
	GRAFANA           = "grafana"
	PROMETHEUS        = "prometheus"
)

func GetInstaller(installType string, path string) (Installer, error) {
	switch {
	case installType == RHEL:
		d := new(RPM)
		d.Init(path)
		return d, nil
	case installType == NATIVE:
		d := new(Native)
		d.Init(path)
		return d, nil
	case installType == GRAFANA:
		d := new(Grafana)
		d.Init(path)
		return d, nil
	case installType == PROMETHEUS:
		d := new(Prometheus)
		d.Init(path)
		return d, nil
	}
	return nil, errors.New("Wrong installer type passed " + installType)
}
