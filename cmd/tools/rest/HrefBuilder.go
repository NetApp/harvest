package rest

import "strings"

type HrefBuilder struct {
	apiPath       string
	fields        string
	filter        []string
	queryFields   string
	queryValue    string
	maxRecords    string
	returnTimeout string
	endpoint      string
}

func NewHrefBuilder() *HrefBuilder {
	return &HrefBuilder{}
}

func (b *HrefBuilder) APIPath(apiPath string) *HrefBuilder {
	b.apiPath = apiPath
	return b
}

func (b *HrefBuilder) Fields(fields string) *HrefBuilder {
	b.fields = fields
	return b
}

func (b *HrefBuilder) Filter(filter []string) *HrefBuilder {
	b.filter = filter
	return b
}

func (b *HrefBuilder) QueryFields(queryFields string) *HrefBuilder {
	b.queryFields = queryFields
	return b
}

func (b *HrefBuilder) QueryValue(queryValue string) *HrefBuilder {
	b.queryValue = queryValue
	return b
}

func (b *HrefBuilder) MaxRecords(maxRecords string) *HrefBuilder {
	b.maxRecords = maxRecords
	return b
}

func (b *HrefBuilder) ReturnTimeout(returnTimeout string) *HrefBuilder {
	b.returnTimeout = returnTimeout
	return b
}

func (b *HrefBuilder) Endpoint(endpoint string) *HrefBuilder {
	b.endpoint = endpoint
	return b
}

func (b *HrefBuilder) Build() string {
	href := strings.Builder{}
	if b.endpoint == "" {
		if !strings.HasPrefix(b.apiPath, "api/") {
			href.WriteString("api/")
		}
		href.WriteString(b.apiPath)
	} else {
		href.WriteString(b.endpoint)
	}
	href.WriteString("?return_records=true")
	addArg(&href, "&fields=", b.fields)
	for _, f := range b.filter {
		addArg(&href, "&", f)
	}
	addArg(&href, "&query_fields=", b.queryFields)
	addArg(&href, "&query=", b.queryValue)
	addArg(&href, "&max_records=", b.maxRecords)
	addArg(&href, "&return_timeout=", b.returnTimeout)
	return href.String()
}

func addArg(href *strings.Builder, field string, value string) {
	if value == "" {
		return
	}
	href.WriteString(field)
	href.WriteString(value)
}
