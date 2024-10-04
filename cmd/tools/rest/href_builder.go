package rest

import (
	"log/slog"
	"slices"
	"strconv"
	"strings"
)

const URLMaxLimit = 8 * 1024

type HrefBuilder struct {
	apiPath                      string
	fields                       []string
	hiddenFields                 []string
	counterSchema                string
	filter                       []string
	queryFields                  string
	queryValue                   string
	maxRecords                   *int
	returnTimeout                *int
	isIgnoreUnknownFieldsEnabled bool
}

func NewHrefBuilder() *HrefBuilder {
	return &HrefBuilder{}
}

func (b *HrefBuilder) APIPath(apiPath string) *HrefBuilder {
	b.apiPath = apiPath
	return b
}

func (b *HrefBuilder) Fields(fields []string) *HrefBuilder {
	b.fields = fields
	return b
}

func (b *HrefBuilder) HiddenFields(hiddenFields []string) *HrefBuilder {
	b.hiddenFields = hiddenFields
	return b
}

func (b *HrefBuilder) CounterSchema(counterSchema []string) *HrefBuilder {
	b.counterSchema = strings.Join(counterSchema, ",")
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

func (b *HrefBuilder) MaxRecords(maxRecords *int) *HrefBuilder {
	b.maxRecords = maxRecords
	return b
}

func (b *HrefBuilder) ReturnTimeout(returnTimeout *int) *HrefBuilder {
	b.returnTimeout = returnTimeout
	return b
}

func (b *HrefBuilder) IsIgnoreUnknownFieldsEnabled(isIgnoreUnknownFieldsEnabled bool) *HrefBuilder {
	b.isIgnoreUnknownFieldsEnabled = isIgnoreUnknownFieldsEnabled
	return b
}

func (b *HrefBuilder) Build() string {
	href := strings.Builder{}
	if !strings.HasPrefix(b.apiPath, "api/") {
		href.WriteString("api/")
	}
	href.WriteString(b.apiPath)

	href.WriteString("?return_records=true")

	if len(b.hiddenFields) > 0 {
		fieldsMap := make(map[string]bool)
		for _, field := range b.fields {
			fieldsMap[field] = true
		}

		// append hidden fields
		for _, hiddenField := range b.hiddenFields {
			if _, exists := fieldsMap[hiddenField]; !exists {
				b.fields = append(b.fields, hiddenField)
				fieldsMap[hiddenField] = true
			}
		}
	}

	if len(strings.Join(b.fields, ",")) > URLMaxLimit {
		b.fields = append([]string{"*"}, b.hiddenFields...)
		if len(strings.Join(b.fields, ",")) > URLMaxLimit {
			slog.Info("converting to * due to URL max limit")
			b.fields = []string{"*"}
		}
	}

	// Sort fields so that the href is deterministic
	slices.Sort(b.fields)

	addArg(&href, "&fields=", strings.Join(b.fields, ","))
	addArg(&href, "&counter_schemas=", b.counterSchema)

	// Sort filters so that the href is deterministic
	slices.Sort(b.filter)

	for _, f := range b.filter {
		addArg(&href, "&", f)
	}
	addArg(&href, "&query_fields=", b.queryFields)
	addArg(&href, "&query=", b.queryValue)
	if b.maxRecords != nil {
		addArg(&href, "&max_records=", strconv.Itoa(*b.maxRecords))
	}
	if b.returnTimeout != nil {
		addArg(&href, "&return_timeout=", strconv.Itoa(*b.returnTimeout))
	}
	if b.isIgnoreUnknownFieldsEnabled {
		addArg(&href, "&ignore_unknown_fields=", "true")
	}
	return href.String()
}

func addArg(href *strings.Builder, field string, value string) {
	if value == "" {
		return
	}
	href.WriteString(field)
	href.WriteString(value)
}
