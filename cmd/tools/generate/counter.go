package generate

import (
	"encoding/json"
	"fmt"
	"github.com/netapp/harvest/v2/cmd/tools"
	"github.com/netapp/harvest/v2/cmd/tools/grafana"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	SgVersion      = "11.6.0"
	CiscoVersion   = "9.3.12"
	ESeriesVersion = "11.80.0"
)

var panelKeyMap = make(map[string]bool)
var opts = &tools.Options{
	Loglevel: 2,
	Image:    "harvest:latest",
}

func generateCounterTemplate(metricsPanelMap map[string]tools.PanelData) (map[string]tools.Counter, map[string]tools.Counter, map[string]tools.Counter) {
	sgCounters := tools.GenerateCounters("", make(map[string]tools.Counter), "storagegrid", metricsPanelMap)
	tools.GenerateStorageGridCounterTemplate(sgCounters, SgVersion)
	ciscoCounters := tools.GenerateCounters("", make(map[string]tools.Counter), "cisco", metricsPanelMap)
	tools.GenerateCiscoSwitchCounterTemplate(ciscoCounters, CiscoVersion)
	eseriesCounters := tools.GenerateCounters("", make(map[string]tools.Counter), "eseries", metricsPanelMap)
	tools.GenerateESeriesCounterTemplate(eseriesCounters, ESeriesVersion)
	return sgCounters, ciscoCounters, eseriesCounters
}

// generateMetadataFiles generates JSON metadata files for MCP server consumption
func generateMetadataFiles(ontapCounters, sgCounters, ciscoCounters, eseriesCounters map[string]tools.Counter) {
	metadataDir := "mcp/metadata"
	if err := os.MkdirAll(metadataDir, 0750); err != nil {
		fmt.Printf("Error creating metadata directory: %v\n", err)
		return
	}

	// Generate ONTAP metadata
	ontapMetadata := extractMetricDescriptions(ontapCounters)
	ontapPath := filepath.Join(metadataDir, "ontap_metrics.json")
	if err := writeMetadataFile(ontapPath, ontapMetadata); err != nil {
		fmt.Printf("Error writing ONTAP metadata: %v\n", err)
	} else {
		fmt.Printf("ONTAP metadata file generated at %s with %d metrics\n", ontapPath, len(ontapMetadata))
	}

	// Generate StorageGrid metadata
	sgMetadata := extractMetricDescriptions(sgCounters)
	sgPath := filepath.Join(metadataDir, "storagegrid_metrics.json")
	if err := writeMetadataFile(sgPath, sgMetadata); err != nil {
		fmt.Printf("Error writing StorageGrid metadata: %v\n", err)
	} else {
		fmt.Printf("StorageGrid metadata file generated at %s with %d metrics\n", sgPath, len(sgMetadata))
	}

	// Generate Cisco metadata
	ciscoMetadata := extractMetricDescriptions(ciscoCounters)
	ciscoPath := filepath.Join(metadataDir, "cisco_metrics.json")
	if err := writeMetadataFile(ciscoPath, ciscoMetadata); err != nil {
		fmt.Printf("Error writing Cisco metadata: %v\n", err)
	} else {
		fmt.Printf("Cisco metadata file generated at %s with %d metrics\n", ciscoPath, len(ciscoMetadata))
	}

	// Generate ESeries metadata
	eseriesMetadata := extractMetricDescriptions(eseriesCounters)
	eseriesPath := filepath.Join(metadataDir, "eseries_metrics.json")
	if err := writeMetadataFile(eseriesPath, eseriesMetadata); err != nil {
		fmt.Printf("Error writing ESeries metadata: %v\n", err)
	} else {
		fmt.Printf("ESeries metadata file generated at %s with %d metrics\n", eseriesPath, len(eseriesMetadata))
	}
}

// extractMetricDescriptions extracts just the name->description mapping
func extractMetricDescriptions(counters map[string]tools.Counter) map[string]string {
	metadata := make(map[string]string)
	for _, counter := range counters {
		// Only include counters with descriptions
		if counter.Description != "" {
			metadata[counter.Name] = counter.Description
		}
	}
	return metadata
}

// writeMetadataFile writes metadata to a JSON file
func writeMetadataFile(path string, metadata map[string]string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(metadata)
}

func visitDashboard(dirs []string, metricsPanelMap map[string]tools.PanelData, eachDash func(data []byte, metricsPanelMap map[string]tools.PanelData)) {
	for _, dir := range dirs {
		err := filepath.Walk(dir, func(path string, _ os.FileInfo, err error) error {
			if err != nil {
				log.Fatal("failed to read directory:", err)
			}
			ext := filepath.Ext(path)
			if ext != ".json" {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				log.Fatalf("failed to read dashboards path=%s err=%v", path, err)
			}
			eachDash(data, metricsPanelMap)
			return nil
		})
		if err != nil {
			log.Fatal("failed to read dashboards:", err)
		}
	}
}

func visitExpressions(data []byte, metricsPanelMap map[string]tools.PanelData) {
	// collect all expressions
	expressions := make([]grafana.ExprP, 0)
	dashboard := gjson.GetBytes(data, "title").String()
	uid := gjson.GetBytes(data, "uid").String()
	link := "d/" + uid + "/" + strings.ToLower(strings.Replace(dashboard, ": ", "3a-", 1)) + "?orgId=1&viewPanel="
	gjson.GetBytes(data, "panels").ForEach(func(key, value gjson.Result) bool {
		grafana.DoTarget("", "", key, value, func(path string, expr string, kind string, id string, title string, rowTitle string) {
			expressions = append(expressions, grafana.NewExpr(path, expr, kind, id, title, rowTitle))
		})
		tp := value.Get("type").String()
		rowTitle := ""
		if tp == "row" {
			rowTitle = value.Get("title").String()
		}
		value.Get("panels").ForEach(func(key2, value2 gjson.Result) bool {
			pathPrefix := fmt.Sprintf("panels[%d].", key.Int())
			grafana.DoTarget(pathPrefix, rowTitle, key2, value2, func(path string, expr string, kind string, id string, title string, rowTitle string) {
				expressions = append(expressions, grafana.NewExpr(path, expr, kind, id, title, rowTitle))
			})
			return true
		})
		return true
	})

	for _, expr := range expressions {
		allMatches := metricRe.FindAllStringSubmatch(expr.Expr, -1)
		for _, match := range allMatches {
			m := match[1]
			if m == "" {
				continue
			}

			key := dashboard + expr.RowTitle + expr.Kind + expr.PanelTitle + link + expr.PanelID
			if !panelKeyMap[m+key] {
				panelKeyMap[m+key] = true
				metricsPanelMap[m] = tools.PanelData{Panels: append(metricsPanelMap[m].Panels, tools.PanelDef{Dashboard: dashboard, Row: expr.RowTitle, Type: expr.Kind, Panel: expr.PanelTitle, PanelLink: link + expr.PanelID})}
			}
		}
	}
}
