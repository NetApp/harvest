package conf

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
)

var patternRe = regexp.MustCompile(`\$(|__\w+){([^}]+)}`)

func ExpandVars(in []byte) ([]byte, error) {
	empty := []byte("")
	matches := patternRe.FindAllSubmatch(in, -1)

	for _, match := range matches {
		if len(match) < 3 {
			return empty, fmt.Errorf("regex error, got %d results back for match, expected 3", len(match))
		}

		if bytes.Equal(match[1], []byte("__env")) || bytes.Equal(match[1], empty) {
			updated, err := expandEnv(match[2])
			if err != nil {
				return empty, err
			}

			in = bytes.Replace(in, match[0], updated, 1)
		}
	}

	return in, nil
}

func expandEnv(bytes []byte) ([]byte, error) {
	s := string(bytes)
	envValue := os.Getenv(s)

	// if env variable is hostname, and the var is empty, use os.Hostname
	if s == "HOSTNAME" && envValue == "" {
		hostname, err := os.Hostname()
		if err != nil {
			return nil, err
		}
		return []byte(hostname), nil
	}

	return []byte(envValue), nil
}
