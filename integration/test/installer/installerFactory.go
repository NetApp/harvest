package installer

import "fmt"

const (
	DOCKER              = "docker"
	RHEL                = "rpm"
	NATIVE              = "tar"
	HARVEST_CONFIG_FILE = "harvest.yml"
)

func GetInstaller(installType string, path string) (Installer, error) {
	if installType == DOCKER {
		d := new(Docker)
		d.Init(path)
		return d, nil
	} else if installType == RHEL {
		d := new(RPM)
		d.Init(path)
		return d, nil
	} else if installType == NATIVE {
		d := new(Native)
		d.Init(path)
		return d, nil
	}
	return nil, fmt.Errorf("Wrong installer type passed " + installType)
}
