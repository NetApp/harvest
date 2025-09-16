package volumesnaplock

import (
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"

	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
)

var (
	reDate = regexp.MustCompile(`^P(\d+)([YMD])$`) // years, months, days
	reTime = regexp.MustCompile(`^PT(\d+)([HM])$`) // hours, minutes
)

type VolumeSnaplock struct {
	*plugin.AbstractPlugin
	cfg Config
}

type Config struct {
	DaysPerYear  int64 // default 365
	DaysPerMonth int64 // default 30
}

func DefaultConfig() Config {
	return Config{
		DaysPerYear:  365,
		DaysPerMonth: 30,
	}
}

func plural(word string, n int64) string {
	if n == 1 {
		return word
	}
	return word + "s"
}

// DurationSecondsAndDisplay parses a single-unit ISO-8601-like duration string and
// returns (seconds, human-display, error).
//
//	PnY  (years)
//	PnM  (months)
//	PnD  (days)
//	PTnH (hours)
//	PTnM (minutes)
//
// Special values (case-insensitive):
//
//	"infinite" -> returns seconds = -1, display = "infinite"
//	"none"     -> returns seconds = 0,  display = "none"
//
// Combined durations like "P1Y10M" are rejected.
func DurationSecondsAndDisplay(s string, cfg Config) (int64, string, error) {
	u := strings.ToUpper(strings.TrimSpace(s))
	if u == "" {
		return 0, "", errors.New("empty duration")
	}
	if u == "INFINITE" {
		return -1, "infinite", nil
	}
	if u == "NONE" {
		return 0, "none", nil
	}

	const (
		minute = int64(60)
		hour   = 60 * minute
		day    = 24 * hour
	)

	year := cfg.DaysPerYear * day
	month := cfg.DaysPerMonth * day

	if m := reDate.FindStringSubmatch(u); m != nil {
		n, _ := strconv.ParseInt(m[1], 10, 64)
		switch m[2] {
		case "Y":
			secs := n * year
			return secs, fmt.Sprintf("%d %s", n, plural("year", n)), nil
		case "M":
			secs := n * month
			return secs, fmt.Sprintf("%d %s", n, plural("month", n)), nil
		case "D":
			secs := n * day
			return secs, fmt.Sprintf("%d %s", n, plural("day", n)), nil
		}
	}

	if m := reTime.FindStringSubmatch(u); m != nil {
		n, _ := strconv.ParseInt(m[1], 10, 64)
		switch m[2] {
		case "H":
			secs := n * hour
			return secs, fmt.Sprintf("%d %s", n, plural("hour", n)), nil
		case "M":
			secs := n * minute
			return secs, fmt.Sprintf("%d %s", n, plural("minute", n)), nil
		}
	}

	return 0, s, fmt.Errorf("unsupported duration (must be a single unit): %q", s)
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &VolumeSnaplock{AbstractPlugin: p}
}

func (v *VolumeSnaplock) Init(_ conf.Remote) error {
	if err := v.InitAbc(); err != nil {
		return err
	}
	v.cfg = DefaultConfig()
	return nil
}

func (v *VolumeSnaplock) processDurationField(instance *matrix.Instance, fieldName string) {
	fieldValue := instance.GetLabel(fieldName)
	if fieldValue == "" {
		return
	}

	if seconds, display, err := DurationSecondsAndDisplay(fieldValue, v.cfg); err == nil {
		instance.SetLabel(fieldName+"_seconds", strconv.FormatInt(seconds, 10))
		instance.SetLabel(fieldName+"_display", display)
	} else {
		v.SLogger.Debug("Failed to parse duration field",
			slog.String("field", fieldName),
			slog.String("value", fieldValue),
			slog.String("instance", instance.GetLabel("uuid")),
			slog.String("error", err.Error()))
	}
}

func (v *VolumeSnaplock) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[v.Object]

	durationFields := []string{
		"minimum_retention_period",
		"default_retention_period",
		"maximum_retention_period",
		"autocommit_period",
	}

	for _, instance := range data.GetInstances() {
		for _, field := range durationFields {
			v.processDurationField(instance, field)
		}
	}
	return nil, nil, nil
}
