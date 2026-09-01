/*
 * Copyright NetApp Inc, 2025 All rights reserved
 */

package rest

import (
	"log/slog"

	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
)

// Array identifies the storage system a collector polls.
type Array struct {
	ID   string
	Name string
	// System is the raw storage-system document, so callers can read the
	// fields only they care about (chassisSerialNumber, model, ...).
	System gjson.Result
}

// DiscoverArray returns the storage system this poller collects from. A web
// services proxy can front more than one array; Harvest polls the first and
// warns about the rest. Name falls back to ID so callers always have a
// non-empty display value.
func (c *Client) DiscoverArray(logger *slog.Logger) (Array, error) {
	systems, err := c.GetStorageSystems()
	if err != nil {
		return Array{}, err
	}

	if len(systems) == 0 {
		return Array{}, errs.New(errs.ErrNoInstance, "no storage system found")
	}

	if len(systems) > 1 {
		logger.Warn("multiple systems found, using first one", slog.Int("count", len(systems)))
	}

	system := systems[0]
	array := Array{
		ID:     system.Get("id").ClonedString(),
		Name:   system.Get("name").ClonedString(),
		System: system,
	}

	if array.ID == "" {
		return Array{}, errs.New(errs.ErrNoInstance, "system missing id")
	}

	if array.Name == "" {
		array.Name = array.ID
	}

	return array, nil
}
