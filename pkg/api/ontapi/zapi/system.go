// Package zapi Copyright NetApp Inc, 2021 All rights reserved
package zapi

import (
	"errors"
	"fmt"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"regexp"
	"strconv"
)

// See system_test.go for examples
func parse7mode(release string) (int, int, int, error) {
	// NetApp Release 8.2P4 7-Mode: Tue Oct 1 11:24:04 PDT 2013
	r := regexp.MustCompile(`NetApp Release (\d+)\.(\d+)\w(\d+)`)
	matches := r.FindStringSubmatch(release)
	if len(matches) == 4 {
		return toInt(matches[1]), toInt(matches[2]), toInt(matches[3]), nil
	}

	r = regexp.MustCompile(`NetApp Release (\d+)\.(\d+)\.(\d+)`)
	matches = r.FindStringSubmatch(release)
	if len(matches) == 4 {
		return toInt(matches[1]), toInt(matches[2]), toInt(matches[3]), nil
	}

	r = regexp.MustCompile(`NetApp Release (\d+)\.(\d+)\.(\d+)\.(\d+)`)
	matches = r.FindStringSubmatch(release)
	if len(matches) == 5 {
		// 7.0.0.1 becomes 7.0.0
		return toInt(matches[1]), toInt(matches[2]), toInt(matches[3]), nil
	}

	return 0, 0, 0, fmt.Errorf("no valid version tuple found for=[%s]", release)
}

func toInt(s string) int {
	value, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return value
}

// getSystem connects to ONTAP system and retrieves its identity and version
// this works for Clustered and 7-mode systems
func (c *Client) getSystem() error {
	var (
		r           conf.Remote
		release     string
		isClustered bool
		versionT    [3]int
		response    *node.Node
		err         error
	)

	r = conf.Remote{}

	// fetch system version and model
	if response, err = c.InvokeRequestString("system-get-version"); err != nil {
		return err
	}

	release = response.GetChildContentS("version")

	if version := response.GetChildS("version-tuple"); version != nil {
		if tuple := version.GetChildS("system-version-tuple"); tuple != nil {

			for i, v := range []string{"generation", "major", "minor"} {
				if n, err := strconv.ParseInt(tuple.GetChildContentS(v), 0, 16); err == nil {
					versionT[i] = int(n)
				}
			}
		}
	}

	// if version tuple is missing, try to parse from the release string,
	// this is usually the case with 7mode systems
	// e.g., NetApp Release 8.2P4 7-Mode: Tue Oct 1 11:24:04 PDT 2013
	if versionT[0] == 0 {
		i0, i1, i2, err := parse7mode(release)
		if err != nil {
			return err
		}
		versionT[0] = i0
		versionT[1] = i1
		versionT[2] = i2
	}

	clustered := response.GetChildContentS("is-clustered")
	if clustered == "" {
		return errors.New("missing attribute [is-clustered]")
	}
	isClustered = clustered == "true"

	// fetch system name and serial number
	request := "cluster-identity-get"
	if !isClustered {
		request = "system-get-info"
	}

	if response, err = c.InvokeRequestString(request); err != nil {
		return err
	}

	if isClustered {
		if attrs := response.GetChildS("attributes"); attrs != nil {
			if info := attrs.GetChildS("cluster-identity-info"); info != nil {
				r.Name = info.GetChildContentS("cluster-name")
				r.UUID = info.GetChildContentS("cluster-uuid")
			}
		}
	} else {
		if info := response.GetChildS("system-info"); info != nil {
			r.Name = info.GetChildContentS("system-name")
			r.Serial = info.GetChildContentS("system-serial-number")
			// There is no uuid for non-cluster mode, using system-id.
			r.UUID = info.GetChildContentS("system-id")
		}
	}

	r.Version = strconv.Itoa(versionT[0]) + "." + strconv.Itoa(versionT[1]) + "." + strconv.Itoa(versionT[2])
	if isClustered {
		r.Model = "cdot"
	} else {
		r.Model = "7mode"
	}
	r.IsClustered = isClustered
	r.ZAPIsExist = true
	r.Release = release
	c.remote = r

	return nil
}
