package auditlog

import (
	"fmt"
	"gopkg.in/yaml.v3"
)

type Filter struct {
	State string `yaml:"state"`
}

type Volume struct {
	Action []string `yaml:"action"`
	Filter Filter   `yaml:"filter"`
}

type Login struct {
	Action []string `yaml:"action"`
	Filter Filter   `yaml:"filter"`
}

type Config struct {
	Volume Volume `yaml:"Volume"`
	Login  Login  `yaml:"Login"`
}

type RootConfig struct {
	AuditLog Config `yaml:"AuditLog"`
}

func InitAuditLogConfig() (RootConfig, error) {
	data := `
AuditLog:
  Volume:
    action:
      - PATCH /api/storage/volumes
      - POST /api/application/applications
      - POST /api/storage/volumes
      - volume create
      - volume modify
      - volume delete
      - POST /api/private/cli/volume
      - DELETE /api/private/cli/volume
      - POST /api/private/cli/volume/rename
      - DELETE /api/storage/volumes
    filter:
      state: success
`

	var config RootConfig
	err := yaml.Unmarshal([]byte(data), &config)
	if err != nil {
		fmt.Printf("Error parsing YAML: %s\n", err)
		return RootConfig{}, err
	}

	return config, nil
}
