package installer

import (
	"errors"
)

const (
	RHEL                   = "rpm"
	NATIVE                 = "tar"
	HarvestConfigFile      = "harvest.yml"
	HarvestAdminConfigFile = "harvest_admin.yml"
	GRAFANA                = "grafana"
	PROMETHEUS             = "prometheus"
)

func GetInstaller(installType string, path string) (Installer, error) {
	switch installType {
	case RHEL:
		d := new(RPM)
		d.Init(path)
		return d, nil
	case NATIVE:
		d := new(Native)
		d.Init(path)
		return d, nil
	case GRAFANA:
		d := new(Grafana)
		d.Init(path)
		return d, nil
	case PROMETHEUS:
		d := new(Prometheus)
		d.Init(path)
		return d, nil
	}
	return nil, errors.New("Wrong installer type passed " + installType)
}
