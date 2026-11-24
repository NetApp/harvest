package prometheus

import (
	"log/slog"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/exporters"
	"github.com/netapp/harvest/v2/cmd/poller/exporter"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/set"
)

func newTestDiskCache(t *testing.T) *diskCache {
	t.Helper()
	dc := newDiskCache(time.Minute, t.TempDir(), slog.Default())
	t.Cleanup(dc.Shutdown)
	return dc
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	assert.Nil(t, err)
	return string(b)
}

func streamLines(lines []string, names *set.Set) func(exporters.LineSink) *set.Set {
	return func(sink exporters.LineSink) *set.Set {
		for _, line := range lines {
			sink.Write([]byte(line))
		}
		return names
	}
}

func TestDiskCachePutWritesLines(t *testing.T) {
	dc := newTestDiskCache(t)
	lines := []string{"# HELP a_b Metric for a", "a_b{x=\"1\"} 1", "a_b{x=\"2\"} 2"}

	dc.put("k", streamLines(lines, set.NewFrom([]string{"a_b"})))

	assert.Equal(t, readFile(t, dc.files["k"]), strings.Join(lines, "\n")+"\n")
	assert.Equal(t, dc.metricCounts["k"], 3)
	assert.Equal(t, dc.metricNames["k"].Size(), 1)
	assert.Equal(t, dc.GetMetricCount(), 3)
	assertNoTmp(t, dc.generateFilepath("k"))
}

func assertNoTmp(t *testing.T, filePath string) {
	t.Helper()
	_, err := os.Stat(filePath + tmpSuffix)
	assert.True(t, os.IsNotExist(err))
}

func scrape(dc *diskCache) string {
	rec := httptest.NewRecorder()
	_, _ = dc.streamMetrics(rec, nil, nil)
	return rec.Body.String()
}

func TestDiskCacheScrapeNotBlockedByRender(t *testing.T) {
	dc := newTestDiskCache(t)
	dc.put("k", streamLines([]string{"old"}, nil))

	rendering := make(chan struct{})
	release := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		dc.put("k", func(sink exporters.LineSink) *set.Set {
			sink.Write([]byte("new"))
			close(rendering)
			<-release
			return nil
		})
	}()
	<-rendering

	scraped := make(chan string)
	go func() { scraped <- scrape(dc) }()
	select {
	case body := <-scraped:
		assert.Equal(t, body, "old\n")
	case <-time.After(5 * time.Second):
		t.Fatal("scrape blocked while put was rendering")
	}

	close(release)
	<-done
	assert.Equal(t, scrape(dc), "new\n")
	assertNoTmp(t, dc.generateFilepath("k"))
}

func TestDiskCachePutKeepsOldFileWhenCreateFails(t *testing.T) {
	dc := newTestDiskCache(t)
	dc.put("k", streamLines([]string{"old"}, nil))
	filePath := dc.files["k"]
	before := dc.timers["k"]

	// a directory at the temp file path makes os.Create fail
	assert.Nil(t, os.Mkdir(filePath+tmpSuffix, 0750))

	rendered := false
	dc.put("k", func(sink exporters.LineSink) *set.Set {
		rendered = true
		return streamLines([]string{"new1", "new2"}, nil)(sink)
	})

	assert.True(t, rendered)
	assert.Equal(t, readFile(t, filePath), "old\n")
	assert.Equal(t, dc.metricCounts["k"], 1)
	assert.Equal(t, dc.timers["k"], before)
}

func TestDiskCachePutRemovesTmpWhenRenderPanics(t *testing.T) {
	dc := newTestDiskCache(t)
	dc.put("k", streamLines([]string{"old"}, nil))
	filePath := dc.files["k"]

	func() {
		defer func() {
			assert.True(t, recover() != nil)
		}()
		dc.put("k", func(sink exporters.LineSink) *set.Set {
			sink.Write([]byte("new"))
			panic("render failed")
		})
	}()

	assert.Equal(t, readFile(t, filePath), "old\n")
	assert.Equal(t, dc.metricCounts["k"], 1)
	assertNoTmp(t, filePath)
}

func TestDiskCachePutRemovesTmpWhenRenameFails(t *testing.T) {
	dc := newTestDiskCache(t)
	filePath := dc.generateFilepath("k")

	// a non-empty directory at the target path makes the rename fail
	assert.Nil(t, os.MkdirAll(filepath.Join(filePath, "child"), 0750))

	dc.put("k", streamLines([]string{"new"}, nil))

	_, ok := dc.files["k"]
	assert.False(t, ok)
	assertNoTmp(t, filePath)
}

func TestDiskCacheCleanKeepsTmpFiles(t *testing.T) {
	dc := newTestDiskCache(t)
	inFlight := dc.generateFilepath("k") + tmpSuffix
	unknown := filepath.Join(dc.baseDir, "stale.metrics")
	assert.Nil(t, os.WriteFile(inFlight, []byte("partial"), 0600))
	assert.Nil(t, os.WriteFile(unknown, []byte("stale"), 0600))

	dc.Clean()

	_, err := os.Stat(inFlight)
	assert.Nil(t, err)
	_, err = os.Stat(unknown)
	assert.True(t, os.IsNotExist(err))
}

func TestExportStreamsToDiskCache(t *testing.T) {
	cacheDir := t.TempDir()
	sortLabels := true
	absExp := exporter.New(
		"Prometheus",
		"prom1",
		&options.Options{PromPort: 1},
		conf.Exporter{
			IsTest:     true,
			SortLabels: &sortLabels,
			DiskCache:  &conf.DiskCacheConfig{Path: cacheDir},
		},
		nil,
	)
	p := New(absExp)
	assert.Nil(t, p.Init())
	prom := p.(*Prometheus)
	dc, ok := prom.aCache.(*diskCache)
	assert.True(t, ok)
	t.Cleanup(dc.Shutdown)

	m := setUpChangeMatrix()
	stats, err := p.Export(m)
	assert.Nil(t, err)

	rendered, wantStats, _ := exporters.Render(m, prom.addMetaTags, prom.sortLabels, prom.globalPrefix, prom.Logger, "")
	want := make([]string, 0, len(rendered))
	for _, line := range rendered {
		want = append(want, string(line))
	}
	got := strings.Split(strings.TrimSuffix(readFile(t, dc.generateFilepath("change.change.change")), "\n"), "\n")
	slices.Sort(want)
	slices.Sort(got)

	assert.Equal(t, strings.Join(got, "\n"), strings.Join(want, "\n"))
	assert.Equal(t, stats, wantStats)
	assert.Equal(t, dc.metricCounts["change.change.change"], len(rendered))
	assert.Equal(t, dc.metricNames["change.change.change"].Size(), 1)
	assertNoTmp(t, dc.generateFilepath("change.change.change"))
}
