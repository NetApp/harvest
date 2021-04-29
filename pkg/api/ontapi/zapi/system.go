/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */
package zapi

import (
	"errors"
	"fmt"
	"strconv"
)

type System struct {
	Name         string
	SerialNumber string
	Release      string
	Version      [3]int
	Clustered    bool
}

func (sys *System) String() string {
	var model, version string
	if sys.Clustered == true {
		model = "CDOT"
	} else {
		model = "7MODE"
	}

	version = fmt.Sprintf("(%s version %d.%d.%d)", model, sys.Version[0], sys.Version[1], sys.Version[2])
	return fmt.Sprintf("%s %s (serial %s) (%s)", sys.Name, version, sys.SerialNumber, sys.Release)
}

func (c *Client) GetSystem() (*System, error) {
	var sys System
	var err error

	sys = System{}

	// fetch system version and model
	if err := c.build_request_string("system-get-version", true); err != nil {
		return &sys, err
	}

	response, err := c.Invoke()
	if err != nil {
		return &sys, err
	}

	sys.Release = response.GetChildContentS("version")

	if version := response.GetChildS("version-tuple"); version != nil {
		if tuple := version.GetChildS("system-version-tuple"); tuple != nil {

			gener, _ := strconv.ParseInt(tuple.GetChildContentS("generation"), 0, 16)
			major, _ := strconv.ParseInt(tuple.GetChildContentS("major"), 0, 16)
			minor, _ := strconv.ParseInt(tuple.GetChildContentS("minor"), 0, 16)

			sys.Version[0] = int(gener)
			sys.Version[1] = int(major)
			sys.Version[2] = int(minor)

		}
	}

	// if version tuple is missing try to parse manually
	// this is usually the case with 7mode systems
	if sys.Version[0] == 0 {
		if _, err = fmt.Sscanf(sys.Release, "NetApp Release %d.%d.%d", &sys.Version[0], &sys.Version[1], &sys.Version[2]); err != nil {
			return &sys, errors.New("no valid version tuple found")
		}
	}

	if clustered := response.GetChildContentS("is-clustered"); clustered == "" {
		return &sys, errors.New("Not found [is-clustered]")
	} else if clustered == "true" {
		sys.Clustered = true
	} else {
		sys.Clustered = false
	}

	// fetch system name and serial number
	request := "cluster-identity-get"
	if !sys.Clustered {
		request = "system-get-info"
	}

	if err := c.build_request_string(request, true); err != nil {
		return &sys, err
	}

	response, err = c.Invoke()
	if err != nil {
		return &sys, err
	}

	if sys.Clustered {
		if attrs := response.GetChildS("attributes"); attrs != nil {
			if info := attrs.GetChildS("cluster-identity-info"); info != nil {
				sys.Name = info.GetChildContentS("cluster-name")
				sys.SerialNumber = info.GetChildContentS("cluster-serial-number")
			}
		}
	} else {
		if info := response.GetChildS("system-info"); info != nil {

			sys.Name = info.GetChildContentS("system-name")
			sys.SerialNumber = info.GetChildContentS("system-serial-number")
		}
	}
	c.System = &sys
	return &sys, nil
}
