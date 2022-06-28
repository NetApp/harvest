package grafana

import (
	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
)

func TestDatasource(t *testing.T) {
	dir := "../../../grafana/dashboards"
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal("failed to read directory:", err)
		}
		ext := filepath.Ext(path)
		if ext != ".json" {
			return nil
		}
		checkDashboardForDatasource(t, path)
		return nil
	})
	if err != nil {
		log.Fatal("failed to read dashboards:", err)
	}
}

func checkDashboardForDatasource(t *testing.T, path string) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("failed to read dashboards path=%s err=%v", path, err)
	}
	gjson.GetBytes(data, "panels").ForEach(func(key, value gjson.Result) bool {
		dsResult := value.Get("datasource")
		if dsResult.Type == gjson.Null {
			// if the panel is a row, it is OK if there is no datasource
			if value.Get("type").String() == "row" {
				return true
			}
			t.Errorf("dashboard=%s panel=%s has a null datasource", path, key.String())
		}
		return true
	})
}
