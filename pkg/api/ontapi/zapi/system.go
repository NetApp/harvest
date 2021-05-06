/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */
package zapi

import (
	"errors"
	"fmt"
	"goharvest2/pkg/tree/node"
	"strconv"
)

type system struct {
	name      string
	serial    string
	release   string
	version   [3]int
	clustered bool
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

	// if version tuple is missing try to parse from the release strirng
	// this is usually the case with 7mode systems
	if s.version[0] == 0 {
		if _, err = fmt.Sscanf(s.release, "NetApp Release %d.%d.%d", &s.version[0], &s.version[1], &s.version[2]); err != nil {
			return errors.New("no valid version tuple found")
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
				s.serial = info.GetChildContentS("cluster-serial-number")
			}
		}
	} else {
		if info := response.GetChildS("system-info"); info != nil {
			s.name = info.GetChildContentS("system-name")
			s.serial = info.GetChildContentS("system-serial-number")
		}
	}
	c.system = &s
	return nil
}
