package prometheus

import (
	"bufio"
	"context"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type diskCache struct {
	*sync.Mutex
	files        map[string]string    // key -> filepath
	timers       map[string]time.Time // key -> timestamp
	metricNames  map[string]*set.Set  // key -> metric names
	metricCounts map[string]int       // key -> number of metric lines
	expire       time.Duration
	baseDir      string
	logger       *slog.Logger
	ctx          context.Context
	cancel       context.CancelFunc
	writerPool   *sync.Pool
	readerPool   *sync.Pool
	keyReplacer  *strings.Replacer
}

var _ diskCacher = (*diskCache)(nil)

func newDiskCache(d time.Duration, baseDir string, logger *slog.Logger) *diskCache {
	if d <= 0 {
		logger.Warn("invalid expire duration, using default 5 minutes", slog.Duration("provided", d))
		d = 5 * time.Minute
	}
	if baseDir == "" {
		logger.Warn("empty base directory provided")
		return nil
	}

	_ = os.RemoveAll(baseDir)
	if err := os.MkdirAll(baseDir, 0750); err != nil {
		logger.Warn("failed to create cache directory", slogx.Err(err), slog.String("dir", baseDir))
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	dc := &diskCache{
		Mutex:        &sync.Mutex{},
		files:        make(map[string]string),
		timers:       make(map[string]time.Time),
		metricNames:  make(map[string]*set.Set),
		metricCounts: make(map[string]int),
		expire:       d,
		baseDir:      baseDir,
		logger:       logger,
		ctx:          ctx,
		cancel:       cancel,
		writerPool: &sync.Pool{
			New: func() any {
				return bufio.NewWriterSize(nil, 64*1024)
			},
		},
		readerPool: &sync.Pool{
			New: func() any {
				return bufio.NewReaderSize(nil, 64*1024)
			},
		},
		keyReplacer: strings.NewReplacer("/", "_", "\\", "_", ":", "_"),
	}

	go dc.cleanup()
	return dc
}

// GetStats returns cache statistics.
func (dc *diskCache) GetStats() (*CacheStats, error) {
	stats := &CacheStats{
		UniqueData: make(map[string]map[string][]string),
	}

	seenCollectors := make(map[string]struct{})
	seenObjects := make(map[string]struct{})

	for key := range dc.files {
		if dc.isExpired(key) {
			continue
		}

		parts := strings.Split(key, ".")
		if len(parts) < 2 {
			continue
		}

		collector := parts[0]
		object := parts[1]

		if strings.HasPrefix(object, "metadata_") {
			continue
		}

		metricNames, exists := dc.metricNames[key]
		if !exists || metricNames == nil || metricNames.Size() == 0 {
			continue
		}

		stats.NumMetrics += metricNames.Size()

		if _, exists := stats.UniqueData[collector]; !exists {
			stats.UniqueData[collector] = make(map[string][]string)
			seenCollectors[collector] = struct{}{}
		}

		objectKey := collector + "." + object
		if _, exists := stats.UniqueData[collector][object]; !exists {
			seenObjects[objectKey] = struct{}{}
		}

		stats.UniqueData[collector][object] = metricNames.Values()
	}

	stats.NumCollectors = len(seenCollectors)
	stats.NumObjects = len(seenObjects)

	return stats, nil
}

// GetMetricCount returns the total number of cached metrics.
func (dc *diskCache) GetMetricCount() int {
	count := 0
	for key := range dc.files {
		if dc.isExpired(key) {
			continue
		}
		if metricCount, exists := dc.metricCounts[key]; exists {
			count += metricCount
		}
	}
	return count
}

// Put stores metrics to disk and updates cache metadata.
func (dc *diskCache) Put(key string, data [][]byte, metricNames *set.Set) {
	filePath := dc.generateFilepath(key)

	if err := dc.writeToDisk(filePath, data); err != nil {
		dc.logger.Warn("failed to write cache file",
			slogx.Err(err),
			slog.String("key", key),
			slog.String("file", filePath))
		return
	}

	dc.files[key] = filePath
	dc.timers[key] = time.Now()
	if metricNames != nil && metricNames.Size() > 0 {
		dc.metricNames[key] = metricNames
	} else {
		dc.metricNames[key] = nil
	}
	dc.metricCounts[key] = len(data)

	dc.logger.Debug("cached metrics to disk",
		slog.String("key", key),
		slog.String("file", filePath),
		slog.Int("metrics_count", len(data)))
}

// StreamToWriter streams all non-expired cache files to the writer.
func (dc *diskCache) StreamToWriter(w io.Writer) error {
	var resultErr error
	errorCount := 0
	totalCount := 0

	for key, path := range dc.files {
		if dc.isExpired(key) {
			continue
		}
		totalCount++

		if err := dc.streamFile(path, w); err != nil {
			errorCount++
			if resultErr == nil {
				resultErr = err
			}
			dc.logger.Debug("failed to stream cache file",
				slogx.Err(err), slog.String("file", path))
		}
	}

	if resultErr != nil {
		dc.logger.Warn("failed to stream some cache files",
			slog.Int("failed_count", errorCount),
			slog.Int("total_count", totalCount))
	}
	return resultErr
}

func (dc *diskCache) openFile(filePath string) (*os.File, error) {
	file, err := os.Open(filePath)
	if os.IsNotExist(err) {
		return nil, nil
	}
	return file, err
}

func (dc *diskCache) closeFile(file *os.File) {
	if err := file.Close(); err != nil {
		dc.logger.Debug("failed to close file", slogx.Err(err))
	}
}

func (dc *diskCache) streamFile(filePath string, w io.Writer) error {
	file, err := dc.openFile(filePath)
	if err != nil {
		return err
	}
	if file == nil {
		dc.logger.Debug("file is nil", slog.String("filePath", filePath))
		return nil
	}
	defer dc.closeFile(file)

	reader := dc.readerPool.Get().(*bufio.Reader)
	reader.Reset(file)
	defer dc.readerPool.Put(reader)

	_, err = io.Copy(w, reader)
	return err
}

func (dc *diskCache) Clean() {
	dc.Lock()
	defer dc.Unlock()

	for key, timestamp := range dc.timers {
		if time.Since(timestamp) <= dc.expire {
			continue
		}
		filePath := dc.files[key]

		delete(dc.files, key)
		delete(dc.timers, key)
		delete(dc.metricNames, key)
		delete(dc.metricCounts, key)

		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			dc.logger.Debug("failed to remove expired cache file",
				slogx.Err(err),
				slog.String("file", filePath))
		}

		dc.logger.Debug("expired cache entry", slog.String("key", key))
	}

	entries, err := os.ReadDir(dc.baseDir)
	if err != nil {
		dc.logger.Debug("failed to read cache directory", slogx.Err(err), slog.String("baseDir", dc.baseDir))
		return
	}

	knownFiles := make(map[string]struct{}, len(dc.files))
	for _, path := range dc.files {
		knownFiles[path] = struct{}{}
	}

	for _, entry := range entries {
		fullPath := filepath.Join(dc.baseDir, entry.Name())

		if _, found := knownFiles[fullPath]; !found {
			_ = os.Remove(fullPath)
		}
	}
}

func (dc *diskCache) generateFilepath(key string) string {
	safeKey := dc.keyReplacer.Replace(key)
	return filepath.Join(dc.baseDir, safeKey+".metrics")
}

func (dc *diskCache) writeToDisk(filePath string, data [][]byte) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer dc.closeFile(file)

	writer := dc.writerPool.Get().(*bufio.Writer)
	writer.Reset(file)
	defer dc.writerPool.Put(writer)

	for _, line := range data {
		if _, err := writer.Write(line); err != nil {
			return err
		}
		if err := writer.WriteByte('\n'); err != nil {
			return err
		}
	}

	return writer.Flush()
}

// isExpired checks if a key is expired.
func (dc *diskCache) isExpired(key string) bool {
	if timer, exists := dc.timers[key]; exists {
		return time.Since(timer) >= dc.expire
	}
	return true
}

func (dc *diskCache) cleanup() {
	ticker := time.NewTicker(dc.expire / 2) // Clean twice per expiry period
	defer ticker.Stop()

	for {
		select {
		case <-dc.ctx.Done():
			return
		case <-ticker.C:
			dc.Clean()
		}
	}
}

func (dc *diskCache) Shutdown() {
	if dc.cancel != nil {
		dc.cancel()
	}
}
