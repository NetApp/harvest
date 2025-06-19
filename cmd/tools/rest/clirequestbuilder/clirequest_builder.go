package clirequestbuilder

import (
	"encoding/json"
	"errors"
	"strings"
)

type CLIRequestBuilder struct {
	apiPath   string
	baseSet   string
	query     string
	object    string
	filter    string
	fields    []string
	counters  []string
	instances []string
}

func New() *CLIRequestBuilder {
	return &CLIRequestBuilder{}
}

func (c *CLIRequestBuilder) BaseSet(baseSet string) *CLIRequestBuilder {
	c.baseSet = baseSet
	return c
}

func (c *CLIRequestBuilder) APIPath(apiPath string) *CLIRequestBuilder {
	c.apiPath = apiPath
	return c
}

func (c *CLIRequestBuilder) Query(query string) *CLIRequestBuilder {
	c.query = query
	return c
}

func (c *CLIRequestBuilder) Object(object string) *CLIRequestBuilder {
	c.object = object
	return c
}

func (c *CLIRequestBuilder) Fields(fields []string) *CLIRequestBuilder {
	c.fields = fields
	return c
}

func (c *CLIRequestBuilder) Filter(filter string) *CLIRequestBuilder {
	c.filter = filter
	return c
}

func (c *CLIRequestBuilder) Counters(counters []string) *CLIRequestBuilder {
	c.counters = counters
	return c
}

func (c *CLIRequestBuilder) Instances(instances []string) *CLIRequestBuilder {
	c.instances = instances
	return c
}

func (c *CLIRequestBuilder) Build() ([]byte, error) {
	var parts []string
	if c.query == "" {
		return nil, errors.New("query must be provided")
	}
	parts = append(parts, c.query)

	if c.object != "" {
		parts = append(parts, "-object", c.object)
	}

	if c.filter != "" {
		parts = append(parts, "-filter", c.filter)
	}

	if len(c.fields) > 0 {
		parts = append(parts, "-fields", strings.Join(c.fields, ","))
	}

	if len(c.instances) > 0 {
		parts = append(parts, "-instance", strings.Join(c.instances, "|"))
	}

	if len(c.counters) > 0 {
		parts = append(parts, "-counter", strings.Join(c.counters, "|"))
	}

	queryCmd := strings.Join(parts, " ")
	fullCommand := c.baseSet + " " + queryCmd
	fullCommand = strings.TrimSpace(fullCommand)

	payload := map[string]string{
		"input": fullCommand,
	}

	return json.Marshal(payload)
}
