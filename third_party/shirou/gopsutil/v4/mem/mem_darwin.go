// SPDX-License-Identifier: BSD-3-Clause
//go:build darwin

package mem

import (
	"context"
	"golang.org/x/sys/unix"
)

func getHwMemsize() (uint64, error) {
	total, err := unix.SysctlUint64("hw.memsize")
	if err != nil {
		return 0, err
	}
	return total, nil
}

// VirtualMemory returns VirtualmemoryStat.
func VirtualMemory() (*VirtualMemoryStat, error) {
	return VirtualMemoryWithContext(context.Background())
}

func VirtualMemoryWithContext(_ context.Context) (*VirtualMemoryStat, error) {
	total, err := getHwMemsize()
	if err != nil {
		return nil, err
	}

	return &VirtualMemoryStat{
		Total: total,
	}, nil
}
