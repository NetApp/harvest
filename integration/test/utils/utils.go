package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/netapp/harvest/v2/cmd/tools/grafana"
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
	"strconv"
	"strings"
	"testing"
	"time"
)

const (
	GrafanaPort    = "3000"
	PrometheusPort = "9090"
	GrafanaTokeKey = "grafana_api_token"

	AnalyzeDockerLogs = "ANALYZE_DOCKER_LOGS"
	CopyDockerLogs    = "COPY_DOCKER_LOGS"
	InstallDocker     = "INSTALL_DOCKER"
	InstallNative     = "INSTALL_NATIVE"
	InstallRPM        = "INSTALL_RPM"
	Regression        = "REGRESSION"
	UpgradeRPM        = "UPGRADE_RPM"
	STOP              = "STOP"
)

func Run(command string, arg ...string) (string, error) {
	return Exec("", command, nil, arg...)
}

func MkDir(dirname string) {
	if _, err := os.Stat(dirname); os.IsNotExist(err) {
		err := os.Mkdir(dirname, 0750)
		PanicIfNotNil(err)
	}
}

func GetConfigDir() string {
	value := os.Getenv("TEST_CONFIG")
	if value != "" {
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
	cmd.Env = append(cmd.Env, env...)
	if dir != "" {
		cmd.Dir = dir
	}
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	fmt.Println("----------Output---------")
	if out.String() != "" {
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
func DownloadFile(aPath string, url string) error {

	// Get the data
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(aPath)
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
		if time.Since(now) > waitFor {
			return false
		}
		time.Sleep(time.Second * 1)
	}
}

func IsURLReachable(url string) bool {
	response, err := http.Get(url) //nolint:gosec
	if err != nil {
		return false
	}
	defer response.Body.Close()
	return response.StatusCode == http.StatusOK
}

func AddPrometheusToGrafana() {
	log.Info().Msg("Add Prometheus into Grafana")
	url := GetGrafanaHTTPURL() + "/api/datasources"
	method := "POST"
	//goland:noinspection HttpUrlsUsage
	jsonValue := []byte(fmt.Sprintf(`{"name": "%s", "type": "prometheus", "access": "direct",
		"url": "%s", "isDefault": true, "basicAuth": false}`,
		grafana.DefaultDataSource,
		"http://"+GetOutboundIP()+":"+PrometheusPort))
	data := SendReqAndGetRes(url, method, jsonValue)
	key := fmt.Sprintf("%v", data["message"])
	if key == "Datasource added" {
		log.Info().Msg("Prometheus has been added successfully into Grafana .")
		return
	}
	panic(errors.New("ERROR: unable to add Prometheus into grafana"))
}

func CreateGrafanaToken() string {
	log.Info().Msg("Creating grafana API Key.")
	url := GetGrafanaHTTPURL() + "/api/auth/keys"
	method := "POST"
	name := strconv.FormatInt(time.Now().Unix(), 10)
	values := map[string]string{"name": name, "role": "Admin"}
	jsonValue, err := json.Marshal(values)
	if err != nil {
		panic(err)
	}
	data := SendReqAndGetRes(url, method, jsonValue)
	key := fmt.Sprintf("%v", data["key"])
	if key != "" {
		log.Info().Msg("Grafana: Token has been created successfully.")
		return key
	}
	panic(errors.New("ERROR: unable to create grafana token"))
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
	panic(errors.New("ERROR : Failed to get ip address of this system"))
}

func WriteToken(token string) {
	var err error
	filename := "harvest.yml"
	abs, _ := filepath.Abs(filename)
	_, err = conf.LoadHarvestConfig(filename)
	PanicIfNotNil(err)
	tools := conf.Config.Tools
	if tools != nil {
		if tools.GrafanaAPIToken != "" {
			log.Error().Str("path", abs).Msg("Harvest.yml contains a grafana token")
			return
		}
	}
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to open file")
	}
	defer func(f *os.File) { _ = f.Close() }(f)
	_, _ = fmt.Fprintf(f, "\n%s\n", "Tools:")
	_, _ = fmt.Fprintf(f, "  %s: %s\n", GrafanaTokeKey, token)
	log.Info().Str("path", abs).Msg("Wrote Grafana token to harvest.yml")
}

func GetGrafanaHTTPURL() string {
	//goland:noinspection HttpUrlsUsage
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
			if item != "" {
				list = append(list, item)
			}
		}
	}
	return list
}

func SetupLogging() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.ErrorStackMarshaler = MarshalStack //nolint:reassign
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
	for range 5 {
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
