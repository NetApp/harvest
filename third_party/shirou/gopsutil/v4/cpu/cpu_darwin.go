// SPDX-License-Identifier: BSD-3-Clause
//go:build darwin

package cpu

// sys/resource.h
const (
	CPUser    = 0
	cpNice    = 1
	cpSys     = 2
	cpIntr    = 3
	cpIdle    = 4
	cpUStates = 5
)

// mach/machine.h
const (
	cpuStateUser   = 0
	cpuStateSystem = 1
	cpuStateIdle   = 2
	cpuStateNice   = 3
	cpuStateMax    = 4
)

// mach/processor_info.h
const (
	processorCpuLoadInfo = 2 //nolint:revive //FIXME
)

type hostCpuLoadInfoData struct { //nolint:revive //FIXME
	cpuTicks [cpuStateMax]uint32
}
