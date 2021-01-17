
package collector

import (
	"fmt"
	"strconv"
	"local.host/xmltree"
)

type SystemInfo struct {
    Name string
    SerialNumber string
    Release string
    Version [3]int
    Clustered bool
}

func (sys SystemInfo) ToString() string {
    var model, version string
    if sys.Clustered == true {
        model = "CDOT"
    } else {
        model = "7MODE"
    }

    version = fmt.Sprintf("(%s version %d.%d.%d)", model, sys.Version[0], sys.Version[1], sys.Version[2])
    return fmt.Sprintf("%s %s (serial %s) (%s)", sys.Name, version, sys.SerialNumber, sys.Release)
}

func (c *Collector) GetSystemInfo() (SystemInfo, error) {
    var sys SystemInfo
    var node *xmltree.Node
    var err error
    var request string
    var found bool

    sys = SystemInfo{}

    // fetch system version and mode
    Log.Debug("Fetching system version")

    c.Client.BuildRequest(xmltree.New("system-get-version"))

    node, err = c.Client.InvokeRequest()
    if err != nil { return sys, err }

    release, _ := node.GetChildContent("version")
    sys.Release = string(release)

    version, found := node.GetChild("version-tuple")
    if found == true {
        tuple, found := version.GetChild("system-version-tuple")
        if found == true {

            gen, _ := tuple.GetChildContent("generation")
            maj, _ := tuple.GetChildContent("major")
            min, _ := tuple.GetChildContent("minor")

            Log.Debug(fmt.Sprintf("convertion version tuple: %s %s %s", string(gen), string(maj), string(min)))
            genint, _ := strconv.ParseInt(string(gen), 0, 16)
            majint, _ := strconv.ParseInt(string(maj), 0, 16)
            minint, _ := strconv.ParseInt(string(min), 0, 16)

            sys.Version[0] = int(genint)
            sys.Version[1] = int(majint)
            sys.Version[2] = int(minint)

        } else {
            Log.Debug("Not found [system-version-tuple]")
        }
    } else {
        Log.Debug("Not found [version-tuple]")
    }

    clustered, found := node.GetChildContent("is-clustered")
    if !found {
        Log.Debug("Not found [is-clustered]")
    } else if string(clustered) == "true" {
        sys.Clustered = true
    } else {
        sys.Clustered = false
    }

    // fetch system name and serial number
    Log.Debug("Fetching system identity")

    if sys.Clustered {
        request = "cluster-identity-get"
    } else {
        request = "system-get-info"
    }

    err = c.Client.BuildRequest(xmltree.New(request))
    if err != nil { return sys, err }

    node, err = c.Client.InvokeRequest()
    if err != nil { return sys, err }

    if sys.Clustered {
        id, found := node.GetChild("attributes")
        if found == true {
            info, found := id.GetChild("cluster-identity-info")
            if found {
                name, _ := info.GetChildContent("cluster-name")
                serial, _ := info.GetChildContent("cluster-serial-number")

                sys.Name = string(name)
                sys.SerialNumber = string(serial)
            } else {
                Log.Debug("Not found [cluster-identity-info]")
            }
        } else {
            Log.Debug("Not found [attributes]")
        }
    } else {
        id, found := node.GetChild("system-info")
        if found == true {
            name, _ := id.GetChildContent("system-name")
            serial, _ := id.GetChildContent("system-serial-number")

            sys.Name = string(name)
            sys.SerialNumber = string(serial)
        } else {
            Log.Debug("Not found [system-info]")
        }
    }

    Log.Debug("Collected system info!")
    return sys, nil
}
