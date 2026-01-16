package rest

import (
	"strings"
)

type URLBuilder struct {
	apiPath  string
	systemID string
	filters  []string
}

func NewURLBuilder() *URLBuilder {
	return &URLBuilder{}
}

func (b *URLBuilder) APIPath(apiPath string) *URLBuilder {
	b.apiPath = apiPath
	return b
}

func (b *URLBuilder) SystemID(systemID string) *URLBuilder {
	b.systemID = systemID
	return b
}

func (b *URLBuilder) Filter(filters []string) *URLBuilder {
	b.filters = filters
	return b
}

func (b *URLBuilder) Build() string {
	url := b.apiPath

	// Replace {system_id} placeholder if systemID is set
	if b.systemID != "" {
		url = strings.ReplaceAll(url, "{system_id}", b.systemID)
	}

	if len(b.filters) > 0 {
		separator := "?"
		if strings.Contains(url, "?") {
			separator = "&"
		}

		var sb strings.Builder
		sb.WriteString(url)
		for i, filter := range b.filters {
			if i == 0 {
				sb.WriteString(separator)
				sb.WriteString(filter)
			} else {
				sb.WriteString("&")
				sb.WriteString(filter)
			}
		}
		url = sb.String()
	}

	return url
}
