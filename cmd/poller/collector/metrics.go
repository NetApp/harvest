package collector

import (
	"github.com/netapp/harvest/v2/pkg/num"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/third_party/shirou/gopsutil/v4/mem"
	"github.com/netapp/harvest/v2/third_party/shirou/gopsutil/v4/process"
	"log/slog"
	"os"
	"runtime/metrics"
)

type MemMetrics struct {
	RSSBytes          uint64
	VMSBytes          uint64
	SwapBytes         uint64
	PercentageRssUsed float64
	LiveHeapBytes     uint64
	HeapSizeBytes     uint64
	HeapGoalBytes     uint64
}

func MemoryMetrics() MemMetrics {
	var memStats MemMetrics

	// Get runtime metrics
	// See https://github.com/golang/go/blob/master/src/runtime/metrics/doc.go
	keys := []string{
		// Heap memory occupied by live objects that were marked by the previous GC.
		"/gc/heap/live:bytes",
		// Memory occupied by live objects and dead objects that have not
		// yet been marked free by the garbage collector.
		"/memory/classes/heap/objects:bytes",
		// Heap size target for the end of the GC cycle.
		"/gc/heap/goal:bytes",
	}
	sample := make([]metrics.Sample, len(keys))
	for i := range keys {
		sample[i].Name = keys[i]
	}
	metrics.Read(sample)

	memStats.LiveHeapBytes = uint64SafeMetric(sample[0])
	memStats.HeapSizeBytes = uint64SafeMetric(sample[1])
	memStats.HeapGoalBytes = uint64SafeMetric(sample[2])

	// Get OS memory metrics
	pid := os.Getpid()
	pid32, err := num.SafeConvertToInt32(pid)
	if err != nil {
		slog.Warn(err.Error(), slog.Int("pid", pid))
		return memStats
	}

	proc, err := process.NewProcess(pid32)
	if err != nil {
		slog.Error("Failed to lookup process for poller", slogx.Err(err), slog.Int("pid", pid))
		return memStats
	}
	memInfo, err := proc.MemoryInfo()
	if err != nil {
		slog.Error("Failed to get memory info for poller", slogx.Err(err), slog.Int("pid", pid))
		return memStats
	}

	// The unix poller used KB for memory so use the same here
	memStats.RSSBytes = memInfo.RSS
	memStats.VMSBytes = memInfo.VMS
	memStats.SwapBytes = memInfo.Swap

	// Calculate memory percentage
	memory, err := mem.VirtualMemory()
	if err != nil {
		slog.Error("Failed to get memory for machine", slogx.Err(err), slog.Int("pid", pid))
		return memStats
	}

	memStats.PercentageRssUsed = float64(memInfo.RSS) / float64(memory.Total) * 100

	return memStats
}

// Return the uint64 value of a metric or zero
func uint64SafeMetric(sample metrics.Sample) uint64 {
	if sample.Value.Kind() == metrics.KindBad {
		return 0
	}
	return sample.Value.Uint64()
}
