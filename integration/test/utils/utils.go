package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"goharvest2/pkg/conf"
	"goharvest2/pkg/tree/node"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	GrafanaPort    = "3000"
	PrometheusPort = "9090"
	GrafanaTokeKey = "grafana_api_token"
)

func Run(command string, arg ...string) string {
	fmt.Print("CMD :" + command + " ")
	for _, param := range arg {
		fmt.Print(param)
		fmt.Print(" ")
	}
	fmt.Print("\n")
	cmd := exec.Command(command, arg...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Println(stderr.String())
		panic(err)
	}
	return out.String()
}

func Exec(dir string, command string, arg ...string) string {
	cmd := exec.Command(command, arg...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		panic(err)
	}
	output := string(out[:])
	return output
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer func(out *os.File) {
		err := out.Close()
		if err != nil {
			panic(err)
		}
	}(out)

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func RemoveSafely(filename string) bool {
	exist := FileExists(filename)
	if exist {
		err := os.Remove(filename)
		if err != nil {
			fmt.Println(err)
			return false
		}
		fmt.Println("File " + filename + " has been deleted.")
	}
	return true
}

func CopyFile(src, dst string) (err error) {
	sfi, err := os.Stat(src)
	if err != nil {
		return
	}
	if !sfi.Mode().IsRegular() {
		// cannot copy non-regular files (e.g., directories,
		// symlinks, devices, etc.)
		return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	}
	dfi, err := os.Stat(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return
		}
	} else {
		if !(dfi.Mode().IsRegular()) {
			return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
		}
		if os.SameFile(sfi, dfi) {
			return
		}
	}
	if err = os.Link(src, dst); err == nil {
		return
	}
	err = copyFileContents(src, dst)
	return
}

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer func(in *os.File) {
		err := in.Close()
		if err != nil {
			panic(err)
		}
	}(in)
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}

func IsUrlReachable(url string) bool {
	response, errors := http.Get(url)
	if errors == nil && response.StatusCode == 200 {
		return true
	}
	return false
}

func AddPrometheusToGrafana() {
	log.Println("Add Prometheus into Grafana")
	url := GetGrafanaHttpUrl() + "/api/datasources"
	method := "POST"
	jsonValue := []byte(fmt.Sprintf(`{"name": "Prometheus", "type": "prometheus", "access": "direct",
		"url": "%s", "isDefault": true, "basicAuth": false}`, GetPrometheusUrl()))
	var data map[string]interface{}
	data = SendReqAndGetRes(url, method, jsonValue)
	key := fmt.Sprintf("%v", data["message"])
	if key == "Datasource added" {
		log.Println("Prometheus has been added successfully into Grafana .")
		return
	}
	panic(fmt.Errorf("ERROR: unable to add Prometheus into grafana"))
}

func CreateGrafanaToken() string {
	log.Println("Creating grafana API Key.")
	url := GetGrafanaHttpUrl() + "/api/auth/keys"
	method := "POST"
	name := fmt.Sprint(time.Now().Unix())
	values := map[string]string{"name": name, "role": "Admin"}
	jsonValue, _ := json.Marshal(values)
	var data map[string]interface{}
	data = SendReqAndGetRes(url, method, jsonValue)
	key := fmt.Sprintf("%v", data["key"])
	if len(key) > 0 {
		log.Println("Grafana: Token has been created successfully.")
		return key
	}
	panic(fmt.Errorf("ERROR: unable to create grafana token"))
}

func PanicIfNotNil(err error) {
	if err != nil {
		panic(err)
	}
}

func GetOutboundIP() string {
	if interfaces, err := net.Interfaces(); err == nil {
		for _, interfac := range interfaces {
			if interfac.HardwareAddr.String() != "" {
				if strings.Index(interfac.Name, "en") == 0 ||
					strings.Index(interfac.Name, "eth") == 0 {
					if addrs, err := interfac.Addrs(); err == nil {
						for _, addr := range addrs {
							if addr.Network() == "ip+net" {
								pr := strings.Split(addr.String(), "/")
								if len(pr) == 2 && len(strings.Split(pr[0], ".")) == 4 {
									return pr[0]
								}
							}
						}
					}
				}
			}
		}
	}
	panic(fmt.Errorf("ERROR : Failed to get ip address of this system"))
}

func WriteToken(token string) {
	var (
		params, tools *node.Node
		err           error
	)
	filename := "harvest.yml"
	if params, err = conf.LoadConfig(filename); err != nil {
		PanicIfNotNil(err)
	} else if params == nil {
		PanicIfNotNil(fmt.Errorf("config [%s] not found", filename))
	}
	if tools = params.GetChildS("Tools"); tools != nil {
		token = tools.GetChildContentS("grafana_api_token")
		if len(token) > 0 {
			log.Println(filename + "  has an entry for grafana token")
			return
		}
	}
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	defer func(f *os.File) {
		err := f.Close()
		PanicIfNotNil(err)
	}(f)
	fmt.Fprintf(f, "\n%s\n", "Tools:")
	fmt.Fprintf(f, "  %s: %s\n", GrafanaTokeKey, token)
}

func GetGrafanaHttpUrl() string {
	return "http://admin:admin@" + GetGrafanaUrl()
}

func GetGrafanaUrl() string {
	return "localhost:" + GrafanaPort
}

func GetPrometheusUrl() string {
	return "http://localhost:" + PrometheusPort
}

func Contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}
