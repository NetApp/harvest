// Package zapi Copyright NetApp Inc, 2021 All rights reserved
package zapi

import (
	"errors"
	"fmt"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"regexp"
	"strconv"
)

type system struct {
	name        string
	serial      string
	clusterUUID string
	release     string
	version     [3]int
	clustered   bool
}

// See system_test.go for examples
func (s *system) parse7mode(release string) error {
	// NetApp Release 8.2P4 7-Mode: Tue Oct 1 11:24:04 PDT 2013
	r := regexp.MustCompile(`NetApp Release (\d+)\.(\d+)\w(\d+)`)
	matches := r.FindStringSubmatch(release)
	if len(matches) == 4 {
		setInt(&s.version[0], matches[1])
		setInt(&s.version[1], matches[2])
		setInt(&s.version[2], matches[3])
		return nil
	}

	r = regexp.MustCompile(`NetApp Release (\d+)\.(\d+)\.(\d+)`)
	matches = r.FindStringSubmatch(release)
	if len(matches) == 4 {
		setInt(&s.version[0], matches[1])
		setInt(&s.version[1], matches[2])
		setInt(&s.version[2], matches[3])
		return nil
	}

	r = regexp.MustCompile(`NetApp Release (\d+)\.(\d+)\.(\d+)\.(\d+)`)
	matches = r.FindStringSubmatch(release)
	if len(matches) == 5 {
		// 7.0.0.1 becomes 7.0.0
		setInt(&s.version[0], matches[1])
		setInt(&s.version[1], matches[2])
		setInt(&s.version[2], matches[3])
		return nil
	}

	return fmt.Errorf("no valid version tuple found for=[%s]", release)
}

func setInt(i *int, s string) {
	value, err := strconv.Atoi(s)
	if err != nil {
		return
	}
	*i = value
}

// getSystem connects to ONTAP system and retrieves its identity and version
// this works for Clustered and 7-mode systems
func (c *Client) getSystem() error {
	var (
		s        system
		response *node.Node
		err      error
	)

	s = system{}

	// fetch system version and model
	if response, err = c.InvokeRequestString("system-get-version"); err != nil {
		return err
	}

	s.release = response.GetChildContentS("version")

	if version := response.GetChildS("version-tuple"); version != nil {
		if tuple := version.GetChildS("system-version-tuple"); tuple != nil {

			for i, v := range []string{"generation", "major", "minor"} {
				if n, err := strconv.ParseInt(tuple.GetChildContentS(v), 0, 16); err == nil {
					s.version[i] = int(n)
				}
			}
		}
	}

	// if version tuple is missing try to parse from the release string
	// this is usually the case with 7mode systems
	// e.g. NetApp Release 8.2P4 7-Mode: Tue Oct 1 11:24:04 PDT 2013
	if s.version[0] == 0 {
		err := s.parse7mode(s.release)
		if err != nil {
			return err
		}
	}

	if clustered := response.GetChildContentS("is-clustered"); clustered == "" {
		return errors.New("missing attribute [is-clustered]")
	} else if clustered == "true" {
		s.clustered = true
	} else {
		s.clustered = false
	}

	// fetch system name and serial number
	request := "cluster-identity-get"
	if !s.clustered {
		request = "system-get-info"
	}

	if response, err = c.InvokeRequestString(request); err != nil {
		return err
	}

	if s.clustered {
		if attrs := response.GetChildS("attributes"); attrs != nil {
			if info := attrs.GetChildS("cluster-identity-info"); info != nil {
				s.name = info.GetChildContentS("cluster-name")
				s.clusterUUID = info.GetChildContentS("cluster-uuid")
			}
		}
	} else {
		if info := response.GetChildS("system-info"); info != nil {
			s.name = info.GetChildContentS("system-name")
			s.serial = info.GetChildContentS("system-serial-number")
			// There is no uuid for non cluster mode, using system-id.
			s.clusterUUID = info.GetChildContentS("system-id")
		}
	}
	c.system = &s
	return nil
}
