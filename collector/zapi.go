
package collector

import (
    "fmt"
    "errors"
    "strconv"
    "local.host/api"
    "local.host/matrix"
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

type Zapi struct {
    Class string
    Name string
    Connection api.Connection
    System SystemInfo
    Data matrix.Matrix
}

func (z Zapi) GetSystemInfo() (SystemInfo, error) {
    var sys SystemInfo
    var err error
    var status, reason, request string
    var found bool

    sys = SystemInfo{}

    // fetch system version and mode
    z.Log("Fetching system version")

    node := api.NewNode("system-get-version")
    xml, err := node.Build()
    if err != nil { return sys, err }

    re, err := z.Connection.InvokeAPI(string(xml))
    if err != nil { return sys, err }

    node, err = api.Parse(re)
    if err != nil { return sys, err }

    if status, found = node.GetAttribute("status"); !found {
        z.Log("Error: attribute status")
        return sys, errors.New("Status not found")
    } else if status != "passed" {
        reason, _ = node.GetAttribute("reason")
        z.Log(fmt.Sprintf("Request rejected: %s", reason))
        return sys, errors.New("Request rejected")
    }

    release, _ := node.GetChildContent("version")
    sys.Release = string(release)

    version, found := node.GetChild("version-tuple")
    if found == true {
        tuple, found := version.GetChild("system-version-tuple")
        if found == true {

            gen, _ := tuple.GetChildContent("generation")
            maj, _ := tuple.GetChildContent("major")
            min, _ := tuple.GetChildContent("minor")

            z.Log(fmt.Sprintf("convertion version tuple: %s %s %s", string(gen), string(maj), string(min)))
            genint, _ := strconv.ParseInt(string(gen), 0, 16)
            majint, _ := strconv.ParseInt(string(maj), 0, 16)
            minint, _ := strconv.ParseInt(string(min), 0, 16)

            sys.Version[0] = int(genint)
            sys.Version[1] = int(majint)
            sys.Version[2] = int(minint)

        } else {
            z.Log("Not found [system-version-tuple]")
        }
    } else {
        z.Log("Not found [version-tuple]")
    }

    clustered, found := node.GetChildContent("is-clustered")
    if !found {
        z.Log("Not found [is-clustered]")
    } else if string(clustered) == "true" {
        sys.Clustered = true
    } else {
        sys.Clustered = false
    }


    // fetch system name and serial number
    z.Log("Fetching system identity")

    if sys.Clustered {
        request = "cluster-identity-get"
    } else {
        request = "system-get-info"
    }

    node = api.NewNode(request)
    xml, err = node.Build()
    if err != nil { return sys, err }

    re, err = z.Connection.InvokeAPI(string(xml))
    if err != nil { return sys, err }

    node, err = api.Parse(re)
    if err != nil { return sys, err }

    if status, found = node.GetAttribute("status"); !found {
        z.Log("Error: attribute status")
        return sys, errors.New("Status not found")
    } else if status != "passed" {
        reason, _ = node.GetAttribute("reason")
        z.Log(fmt.Sprintf("Request rejected: %s", reason))
        return sys, errors.New("Request rejected")
    }

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
                z.Log("Not found [cluster-identity-info]")
            }
        } else {
            z.Log("Not found [attributes]")
        }
    } else {
        id, found := node.GetChild("system-info")
        if found == true {
            name, _ := id.GetChildContent("system-name")
            serial, _ := id.GetChildContent("system-serial-number")

            sys.Name = string(name)
            sys.SerialNumber = string(serial)
        } else {
            z.Log("Not found [system-info]")
        }
    }

    z.Log("Collected system info!")
    return sys, nil
}

func (z Zapi) Log(msg string) {
    //fmt.Printf("Log called for %s", c.Name)
    fmt.Printf("[%s:%s] %s\n", z.Class, z.Name, msg)
}

func (z Zapi) Init(p api.ConnectionParams) error {

    var err error

    z.Log("Intializing!")

    z.Connection, err = api.NewConnection(p)
    if err != nil {
        z.Log("Error connection")
        return err
    }

    z.System, err = z.GetSystemInfo()
    if err != nil {
        z.Log("Error fetching system info")
        return err
    }

    z.Log(fmt.Sprintf("Start-up success! Connected to: %s", z.System.ToString()))
    return err
}
