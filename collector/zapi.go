
package collector

import (
    "fmt"
    "errors"
    "strings"
    "strconv"
    "regexp"
    "io/ioutil"
    "path/filepath"
    "local.host/api"
    "local.host/matrix"
    "local.host/share"
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
    Data *matrix.Matrix
    Template *share.Element
}

func NewZapi(class, name string) *Zapi {
    var z Zapi
    z = Zapi{ Class : class, Name : name }
    //z.Connection = new(api.Connection)
    //z.System = new(SystemInfo)
    z.Data = new(matrix.Matrix)
    z.Template = new(share.Element)
    return &z
}

func (z Zapi) GetSystemInfo() (SystemInfo, error) {
    var sys SystemInfo
    var node *api.Node
    var err error
    var request string
    var found bool

    sys = SystemInfo{}

    // fetch system version and mode
    z.Log("Fetching system version")

    z.Connection.BuildRequest(api.NewNode("system-get-version"))

    node, err = z.Connection.InvokeRequest()
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

    err = z.Connection.BuildRequest(api.NewNode(request))
    if err != nil { return sys, err }

    node, err = z.Connection.InvokeRequest()
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

func (z Zapi) Log(format string, vars ...interface{}) {
    fmt.Printf("[%s:%s] ", z.Class, z.Name)
    fmt.Printf(format, vars...)
    fmt.Println()
}

func (z Zapi) Init(params api.Params, template *share.Element) error {

    var err error

    z.Log("Intializing!")

    z.Connection, err = api.NewConnection(params)
    if err != nil {
        z.Log("Error connecting: %s", err)
        return err
    }

    z.System, err = z.GetSystemInfo()
    if err != nil {
        z.Log("Error fetching system info: %s", err)
        return err
    }

    z.Template, err = z.LoadSubtemplate(params.Path, params.Template, params.Subtemplate, z.Class, z.System.Version)
    if err != nil {
        z.Log("Error importing subtemplate: %s", err)
        return err
    }
    p := z.Template
    z.Log("[Init] Address of pointer: %v (%v). Address of value: (%v)", z.Template, p, &p)
    //subtemplate.MergeFrom(template)

    z.Data = matrix.NewMatrix("volume")

    counters := z.Template.GetChild("counters")
    if counters == nil {
        z.Log("Error: subtemplate has no counters sections")
    } else {
        z.Log("Parsing subtemplate counter section: %d values, %d children", len(counters.Values), len(counters.Children))
        empty := make([]string, 0)
        z.ParseCounters(z.Data, counters, empty)
        z.Log("Built counter cache with %d Metrics and %d Labels", len(z.Data.Counters), len(z.Data.Instances))

        z.Log(fmt.Sprintf("Start-up success! Connected to: %s", z.System.ToString()))
    }

    query := z.Template.GetChildValue("query")
    z.Log("I got query: [%s] and template is [%v]", query, z.Template)
    return err
}

func (z Zapi) PollData() error {
    var err error
    var query string
    var node *api.Node

    z.Log("\n\nStarting data poll session: %s", z.System.ToString())

    if z.Template == nil {
        z.Log("template is [%v] and NIL!!", z.Template)
    } else {
        z.Log("template is [%v] and OK!", z.Template)
    }

    query = z.Template.GetChildValue("query")
    if query == "" { panic("missing query in template") }

    z.Connection.BuildRequest(api.NewNode(query))

    node, err = z.Connection.InvokeRequest()

    if err != nil {
        z.Log("Request for [%s] failed: %s", query, err)
    } else {
        api.PrintTree(node, 0)
    }
    return err
}


func (z Zapi) LoadSubtemplate(path, dir, filename, collector string, version [3]int) (*share.Element, error) {

    var err error
    var selected_version string
    var subtemplate *share.Element

    path_prefix := filepath.Join(path, "var/", strings.ToLower(collector), dir)
    z.Log("Looking for best-fitting template in [%s]", path_prefix)

    available := make(map[string]bool)
    files, _ := ioutil.ReadDir(path_prefix)
    for _, file := range files {
        if match, _ := regexp.MatchString(`\d+\.\d+\.\d+`, file.Name()); match == true && file.IsDir() {
            available[file.Name()] = true
        }
    }

    vers := version[0] * 100 + version[1] * 10 + version[2]
    if err != nil { return subtemplate, err }

    for max:=300; max>0 && vers>0; max-=1 {
        str := strings.Join(strings.Split(strconv.Itoa(vers), ""), ".")
        if _, exists := available[str]; exists == true {
            selected_version = str
            break
        }
        vers-= 1
    }

    if selected_version == "" {
        err = errors.New("No best-fitting subtemplate version found")
    } else {
        path := filepath.Join(path_prefix, selected_version, filename)
        z.Log("Selected best-fitting subtemplate [%s]", path)
        subtemplate, err = share.ImportTemplate(path)
    }
    return subtemplate, err
}


func (z Zapi) ParseCounters(data *matrix.Matrix, elem *share.Element, path []string) {
    z.Log("Parsing [%s] with %d values and %d children", elem.Name, len(elem.Values), len(elem.Children))
    for _, value := range elem.Values {
        z.HandleCounter(data, path, value)
    }
    new_path := append(path, elem.Name)
    for _, child := range elem.Children {
        z.ParseCounters(data, child, new_path)
    }
}

func (z Zapi) HandleCounter(data *matrix.Matrix, path []string, value string) {
    var name, display, flat_path string
    var split_value, full_path []string

    split_value = strings.Split(value, "=>")
    if len(split_value) == 1 {
        name = value
    } else {
        name = split_value[0]
        display = strings.TrimLeft(split_value[1], " ")
    }

    name = strings.TrimLeft(name, "^")
    name = strings.TrimRight(name, " ")

    full_path = append(path[1:], name)
    flat_path = strings.Join(full_path, ".")

    if display == "" {
        display = ParseDisplay(data.Object, full_path)
    }

    if value[0] == '^' {
        data.AddLabel(flat_path, display)
            z.Log("Added as Label [%s] [%s]", display, flat_path)
        if value[1] == '^' {
            data.AddInstanceKey(full_path)
            z.Log("Added as Key [%s] [%s]", display, flat_path)
        }
    } else {
        data.AddCounter(flat_path, display, true)
            z.Log("Added as Metric [%s] [%s]", display, flat_path)
    }
}

func ParseDisplay(obj string, path []string) string {
    var ignore = map[string]int{"attributes" : 0, "info" : 0, "list" : 0, "details" : 0}
    var added = map[string]int{}
    var words []string

    for _, attribute := range path {
        split := strings.Split(attribute, "-")
        for _, word := range split {
            if word == obj { continue }
            if _, exists := ignore[word]; exists { continue }
            if _, exists := added[word]; exists { continue }
            words = append(words, word)
            added[word] = 0
        }
    }
    return strings.Join(words, "_")
}
