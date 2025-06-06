/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package servicecontrol

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/api/servicecontrol/v1"
)

type Measurement struct {
	measurement string
	tagSet      []string
	fieldSet    []string
	timestamp   string
}

func NewMeasurement(name string, size int) *Measurement {
	m := Measurement{measurement: name}
	m.tagSet = make([]string, size)
	m.fieldSet = make([]string, 0)
	return &m
}

func (m *Measurement) String() string {
	format := "\nmeasurement=%s \ntag_set=%v \nfield_set=%v \ntimestamp=%s"
	return fmt.Sprintf(format, m.measurement, m.tagSet, m.fieldSet, m.timestamp)
}

func (m *Measurement) AddTag(key, value string) {
	m.tagSet = append(m.tagSet, escape(key)+"="+escape(value))
}

func (m *Measurement) AddField(key, value string) {
	m.fieldSet = append(m.fieldSet, escape(key)+"="+value)
}

func (m *Measurement) AddFieldString(key, value string) {
	value = strings.ReplaceAll(value, "\\", `\\`)
	value = strings.ReplaceAll(value, `"`, `\"`)
	m.AddField(key, `"`+value+`"`)
}

func (m *Measurement) SetTimestamp(t string) {
	m.timestamp = t
}

func (m *Measurement) Render() (*servicecontrol.Operation, error) {

	value := int64(423432)

	//labels["location"] = dataCenter
	//labels["resource_container"] = "projects/" + consumerId
	//labels["name"], _ = googleMetric.GetResourceName()

	return &servicecontrol.Operation{
		OperationId:   uuid.NewString(),
		OperationName: "metric",
		StartTime:     time.Now().Format(time.RFC3339),
		EndTime:       time.Now().Add(3 * time.Second).Format(time.RFC3339),
		Labels: map[string]string{
			"location":           "us-east1",
			"resource_container": "projects/sridhar-yalla",
			"name":               "test-volume",
		},
		ConsumerId: "project:sridhar-yalla",
		MetricValueSets: []*servicecontrol.MetricValueSet{
			{
				MetricName: "netapp.googleapis.com/volume/allocated_bytes",
				MetricValues: []*servicecontrol.MetricValue{
					{
						Int64Value: &value,
					},
				},
			},
		},
	}, nil
}

func generateSql(metricType string, tagSet []string, fieldSet []string) string {
	var columns, values []string
	for _, tag := range tagSet {
		parts := strings.SplitN(tag, "=", 2)
		if len(parts) == 2 {
			columns = append(columns, parts[0])
			values = append(values, "'"+parts[1]+"'")
		}
	}
	for _, tag := range fieldSet {
		parts := strings.SplitN(tag, "=", 2)
		if len(parts) == 2 {
			columns = append(columns, parts[0])
			values = append(values, "'"+parts[1]+"'")
		}
	}
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);", metricType, strings.Join(columns, ", "), strings.Join(values, ", "))
}

func joinAll(slices ...string) string {
	return strings.Join(slices, "")
}

func escape(s string) string {
	s = strings.ReplaceAll(s, "\\", `\\`)
	s = strings.ReplaceAll(s, " ", `\ `)
	s = strings.ReplaceAll(s, ",", `\,`)
	s = strings.ReplaceAll(s, "=", `\=`)
	return strings.ReplaceAll(s, "\n", `\n`)
}
