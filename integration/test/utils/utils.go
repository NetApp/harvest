package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

const (
	GrafanaPort    = "3000"
	PrometheusPort = "9090"
	GrafanaTokeKey = "grafana_api_token"

	AnalyzeDockerLogs = "ANALYZE_DOCKER_LOGS"
	BookendEms        = "BOOKEND_EMS"
	CopyDockerLogs    = "COPY_DOCKER_LOGS"
	InstallDocker     = "INSTALL_DOCKER"
	InstallNative     = "INSTALL_NATIVE"
	InstallRPM        = "INSTALL_RPM"
	NonBookendEms     = "NON_BOOKEND_EMS"
	Regression        = "REGRESSION"
	UpgradeRPM        = "UPGRADE_RPM"
)

func Run(command string, arg ...string) (string, error) {
	return Exec("", command, nil, arg...)
}

func MkDir(dirname string) {
	if _, err := os.Stat(dirname); os.IsNotExist(err) {
		err := os.Mkdir(dirname, os.ModePerm)
		PanicIfNotNil(err)
	}
}

func GetConfigDir() string {
	value := os.Getenv("TEST_CONFIG")
	if len(value) > 0 {
		return value
	}
	return "/u/mpeg/harvest"
}

func Exec(dir string, command string, env []string, arg ...string) (string, error) {
	cmdString := command + " "
	for _, param := range arg {
		cmdString = cmdString + param + " "
	}
	fmt.Println("CMD : " + cmdString)
	cmd := exec.Command(command, arg...)
	cmd.Env = os.Environ()
	for _, v := range env {
		cmd.Env = append(cmd.Env, v)
	}
	if len(dir) > 0 {
		cmd.Dir = dir
	}
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	fmt.Println("----------Output---------")
	if len(out.String()) > 0 {
		fmt.Println(out.String())
	}
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("-------------------------")
	return out.String(), err
}

// DownloadFile will download the url to a local file.
// It's efficient because it will write as it downloads and not load the whole file into memory.
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

func RemoveDir(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer func(d *os.File) { _ = d.Close() }(d)
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

func UseCertFile(harvestHome string) {
	harvestCertFile := "harvest_cert.yml"
	harvestFile := "harvest.yml"
	_, _ = Run("cp", "-p", GetConfigDir()+"/"+harvestCertFile, harvestHome+"/"+harvestFile)
	_, _ = Run("certer", "-ip", "10.193.48.11")

	path := harvestHome + "/cert"
	log.Info().Str("path", path).Msg("Copy certificate files")
	if FileExists(path) {
		err := RemoveDir(path)
		PanicIfNotNil(err)
	}
	_, _ = Run("mkdir", "-p", path)
	_, _ = Run("cp", "-R", GetConfigDir()+"/cert", harvestHome)
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

func WaitForGrafana() bool {
	now := time.Now()
	const waitFor = time.Second * 30
	for {
		if IsURLReachable(GetGrafanaHTTPURL()) {
			return true
		}
		if time.Now().Sub(now) > waitFor {
			return false
		}
		time.Sleep(time.Second * 1)
	}
}

func IsURLReachable(url string) bool {
	response, err := http.Get(url)
	if err == nil && response.StatusCode == 200 {
		return true
	}
	return false
}

func AddPrometheusToGrafana() {
	log.Info().Msg("Add Prometheus into Grafana")
	url := GetGrafanaHTTPURL() + "/api/datasources"
	method := "POST"
	jsonValue := []byte(fmt.Sprintf(`{"name": "Prometheus", "type": "prometheus", "access": "direct",
		"url": "%s", "isDefault": true, "basicAuth": false}`, "http://"+GetOutboundIP()+":"+PrometheusPort))
	var data map[string]interface{}
	data = SendReqAndGetRes(url, method, jsonValue)
	key := fmt.Sprintf("%v", data["message"])
	if key == "Datasource added" {
		log.Info().Msg("Prometheus has been added successfully into Grafana .")
		return
	}
	panic(fmt.Errorf("ERROR: unable to add Prometheus into grafana"))
}

func CreateGrafanaToken() string {
	log.Info().Msg("Creating grafana API Key.")
	url := GetGrafanaHTTPURL() + "/api/auth/keys"
	method := "POST"
	name := fmt.Sprint(time.Now().Unix())
	values := map[string]string{"name": name, "role": "Admin"}
	jsonValue, _ := json.Marshal(values)
	var data map[string]interface{}
	data = SendReqAndGetRes(url, method, jsonValue)
	key := fmt.Sprintf("%v", data["key"])
	if len(key) > 0 {
		log.Info().Msg("Grafana: Token has been created successfully.")
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
	var err error
	filename := "harvest.yml"
	abs, _ := filepath.Abs(filename)
	err = conf.LoadHarvestConfig(filename)
	PanicIfNotNil(err)
	tools := conf.Config.Tools
	if tools != nil {
		if len(tools.GrafanaAPIToken) > 0 {
			log.Error().Str("path", abs).Msg("Harvest.yml contains a grafana token")
			return
		}
	}
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to open file")
	}
	defer func(f *os.File) { _ = f.Close() }(f)
	_, _ = fmt.Fprintf(f, "\n%s\n", "Tools:")
	_, _ = fmt.Fprintf(f, "  %s: %s\n", GrafanaTokeKey, token)
	log.Info().Str("path", abs).Msg("Wrote Grafana token to harvest.yml")
}

func GetGrafanaHTTPURL() string {
	return "http://admin:admin@" + GetGrafanaURL()
}

func GetGrafanaURL() string {
	return "localhost:" + GrafanaPort
}

func GetPrometheusURL() string {
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

func GetHarvestRootDir() string {
	path, err := os.Getwd()
	PanicIfNotNil(err)
	return filepath.Dir(filepath.Dir(path))
}

func RemoveDuplicateStr(strSlice []string) []string {
	allKeys := make(map[string]bool)
	var list []string
	for _, item := range strSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			if len(item) > 0 {
				list = append(list, item)
			}
		}
	}
	return list
}

func SetupLogging() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.ErrorStackMarshaler = MarshalStack
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, NoColor: true}).
		With().Caller().Stack().Timestamp().Logger()
}

func MarshalStack(err error) interface{} {
	if err == nil {
		return nil
	}
	// We don't know how big the stack trace will be, so start with 10K and double a few times if needed
	n := 10_000
	var trace []byte
	for i := 0; i < 5; i++ {
		trace = make([]byte, n)
		bytesWritten := runtime.Stack(trace, false)
		if bytesWritten < len(trace) {
			trace = trace[:bytesWritten]
			break
		}
		n *= 2
	}
	return string(trace)
}

func SkipIfMissing(t *testing.T, vars ...string) {
	t.Helper()
	anyMatches := false
	for _, v := range vars {
		e := os.Getenv(v)
		if e != "" {
			anyMatches = true
			break
		}
	}
	if !anyMatches {
		t.Skipf("Set one of %s envvars to run this test", strings.Join(vars, ", "))
	}
}
