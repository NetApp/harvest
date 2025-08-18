package cmds

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Netapp/harvest-automation/test/errs"
	"github.com/Netapp/harvest-automation/test/request"
	"github.com/carlmjohnson/requests"
	"github.com/netapp/harvest/v2/cmd/tools/grafana"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
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
	CheckMetrics      = "CHECK_METRICS"
	CopyDockerLogs    = "COPY_DOCKER_LOGS"
	InstallDocker     = "INSTALL_DOCKER"
	InstallNative     = "INSTALL_NATIVE"
	InstallRPM        = "INSTALL_RPM"
	Regression        = "REGRESSION"
	UpgradeRPM        = "UPGRADE_RPM"
	STOP              = "STOP"
	TestStatPerf      = "TEST_STAT_PERF"

	Fips = "fips140=on"
)

func Run(command string, arg ...string) (string, error) {
	return Exec("", command, nil, arg...)
}

func MkDir(dirname string) {
	if _, err := os.Stat(dirname); os.IsNotExist(err) {
		err := os.Mkdir(dirname, 0750)
		errs.PanicIfNotNil(err)
	}
}

func GetConfigDir() string {
	value := os.Getenv("TEST_CONFIG")
	if value != "" {
		return value
	}
	return "/home/harvestfiles"
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
	//goland:noinspection GoUnhandledErrorResult
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

	path := harvestHome + "/cert"
	slog.Info("Copy certificate files", slog.String("path", path))
	if FileExists(path) {
		err := RemoveDir(path)
		errs.PanicIfNotNil(err)
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
	err := requests.URL(url).Fetch(context.Background())
	return err == nil
}

func AddPrometheusToGrafana() {
	slog.Info("Add Prometheus into Grafana")
	url := GetGrafanaHTTPURL() + "/api/datasources"
	method := "POST"
	//goland:noinspection HttpUrlsUsage
	jsonValue := []byte(fmt.Sprintf(`{"name": "%s", "type": "prometheus", "access": "direct",
		"url": "%s", "isDefault": true, "basicAuth": false}`,
		grafana.DefaultDataSource,
		"http://"+GetOutboundIP()+":"+PrometheusPort))
	data := request.SendReqAndGetRes(url, method, jsonValue)
	key := fmt.Sprintf("%v", data["message"])
	if key == "Datasource added" {
		slog.Info("Prometheus has been added successfully into Grafana .")
		return
	}
	panic(errors.New("ERROR: unable to add Prometheus into grafana"))
}

func CreateGrafanaToken() string {
	slog.Info("Creating grafana API Key.")
	url := GetGrafanaHTTPURL() + "/api/auth/keys"
	method := "POST"
	name := strconv.FormatInt(time.Now().Unix(), 10)
	values := map[string]string{"name": name, "role": "Admin"}
	jsonValue, err := json.Marshal(values)
	if err != nil {
		panic(err)
	}
	data := request.SendReqAndGetRes(url, method, jsonValue)
	key := fmt.Sprintf("%v", data["key"])
	if key != "" {
		slog.Info("Grafana: Token has been created successfully.")
		return key
	}
	panic(errors.New("ERROR: unable to create grafana token"))
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
	errs.PanicIfNotNil(err)
	tools := conf.Config.Tools
	if tools != nil {
		if tools.GrafanaAPIToken != "" {
			slog.Error("Harvest.yml contains a grafana token", slog.String("path", abs))
			return
		}
	}
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		slog.Error("Failed to open file", slogx.Err(err))
		os.Exit(1)
	}
	defer func(f *os.File) { _ = f.Close() }(f)
	_, _ = fmt.Fprintf(f, "\n%s\n", "Tools:")
	_, _ = fmt.Fprintf(f, "  %s: %s\n", GrafanaTokeKey, token)
	slog.Info("Wrote Grafana token to harvest.yml", slog.String("path", abs))
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

func GetHarvestRootDir() string {
	path, err := os.Getwd()
	errs.PanicIfNotNil(err)
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
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.SourceKey {
				source := a.Value.Any().(*slog.Source)
				source.File = filepath.Base(source.File)
			}
			return a
		},
	}))
	slog.SetDefault(logger)
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

func SkipIfFipsSet(t *testing.T) {
	environ := os.Environ()
	for _, e := range environ {
		if strings.HasPrefix(e, "GODEBUG") && strings.Contains(e, Fips) {
			// FIPS 140-3 is only supported on ONTAP 9.
			t.Skipf("Skipping test because %s is set in the environment", Fips)
			return
		}
	}
}
