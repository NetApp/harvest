// SPDX-License-Identifier: BSD-3-Clause
//go:build linux

package cpu

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/netapp/harvest/v2/third_party/shirou/gopsutil/v4/internal/common"
)

var armModelToModelName = map[uint64]string{
	0x810: "ARM810",
	0x920: "ARM920",
	0x922: "ARM922",
	0x926: "ARM926",
	0x940: "ARM940",
	0x946: "ARM946",
	0x966: "ARM966",
	0xa20: "ARM1020",
	0xa22: "ARM1022",
	0xa26: "ARM1026",
	0xb02: "ARM11 MPCore",
	0xb36: "ARM1136",
	0xb56: "ARM1156",
	0xb76: "ARM1176",
	0xc05: "Cortex-A5",
	0xc07: "Cortex-A7",
	0xc08: "Cortex-A8",
	0xc09: "Cortex-A9",
	0xc0d: "Cortex-A17",
	0xc0f: "Cortex-A15",
	0xc0e: "Cortex-A17",
	0xc14: "Cortex-R4",
	0xc15: "Cortex-R5",
	0xc17: "Cortex-R7",
	0xc18: "Cortex-R8",
	0xc20: "Cortex-M0",
	0xc21: "Cortex-M1",
	0xc23: "Cortex-M3",
	0xc24: "Cortex-M4",
	0xc27: "Cortex-M7",
	0xc60: "Cortex-M0+",
	0xd01: "Cortex-A32",
	0xd02: "Cortex-A34",
	0xd03: "Cortex-A53",
	0xd04: "Cortex-A35",
	0xd05: "Cortex-A55",
	0xd06: "Cortex-A65",
	0xd07: "Cortex-A57",
	0xd08: "Cortex-A72",
	0xd09: "Cortex-A73",
	0xd0a: "Cortex-A75",
	0xd0b: "Cortex-A76",
	0xd0c: "Neoverse-N1",
	0xd0d: "Cortex-A77",
	0xd0e: "Cortex-A76AE",
	0xd13: "Cortex-R52",
	0xd20: "Cortex-M23",
	0xd21: "Cortex-M33",
	0xd40: "Neoverse-V1",
	0xd41: "Cortex-A78",
	0xd42: "Cortex-A78AE",
	0xd43: "Cortex-A65AE",
	0xd44: "Cortex-X1",
	0xd46: "Cortex-A510",
	0xd47: "Cortex-A710",
	0xd48: "Cortex-X2",
	0xd49: "Neoverse-N2",
	0xd4a: "Neoverse-E1",
	0xd4b: "Cortex-A78C",
	0xd4c: "Cortex-X1C",
	0xd4d: "Cortex-A715",
	0xd4e: "Cortex-X3",
	0xd4f: "Neoverse-V2",
	0xd81: "Cortex-A720",
	0xd82: "Cortex-X4",
	0xd84: "Neoverse-V3",
	0xd85: "Cortex-X925",
	0xd87: "Cortex-A725",
	0xd8e: "Neoverse-N3",
}

func sysCPUPath(ctx context.Context, cpu int32, relPath string) string {
	return common.HostSysWithContext(ctx, fmt.Sprintf("devices/system/cpu/cpu%d", cpu), relPath)
}

func finishCPUInfo(ctx context.Context, c *InfoStat) {
	var lines []string
	var err error
	var value float64

	if c.CoreID == "" {
		lines, err = common.ReadLines(sysCPUPath(ctx, c.CPU, "topology/core_id"))
		if err == nil {
			c.CoreID = lines[0]
		}
	}

	// override the value of c.Mhz with cpufreq/cpuinfo_max_freq regardless
	// of the value from /proc/cpuinfo because we want to report the maximum
	// clock-speed of the CPU for c.Mhz, matching the behaviour of Windows
	lines, err = common.ReadLines(sysCPUPath(ctx, c.CPU, "cpufreq/cpuinfo_max_freq"))
	// if we encounter errors below such as there are no cpuinfo_max_freq file,
	// we just ignore. so let Mhz is 0.
	if err != nil || len(lines) == 0 {
		return
	}
	value, err = strconv.ParseFloat(lines[0], 64)
	if err != nil {
		return
	}
	c.Mhz = value / 1000.0 // value is in kHz
	if c.Mhz > 9999 {
		c.Mhz /= 1000.0 // value in Hz
	}
}

// CPUInfo on linux will return 1 item per physical thread.
//
// CPUs have three levels of counting: sockets, cores, threads.
// Cores with HyperThreading count as having 2 threads per core.
// Sockets often come with many physical CPU cores.
// For example a single socket board with two cores each with HT will
// return 4 CPUInfoStat structs on Linux and the "Cores" field set to 1.
func Info() ([]InfoStat, error) {
	return InfoWithContext(context.Background())
}

func InfoWithContext(ctx context.Context) ([]InfoStat, error) {
	filename := common.HostProcWithContext(ctx, "cpuinfo")
	lines, err := common.ReadLines(filename)
	if err != nil {
		return nil, fmt.Errorf("could not read %s: %w", filename, err)
	}

	var ret []InfoStat
	var processorName string

	c := InfoStat{CPU: -1, Cores: 1}
	for _, line := range lines {
		fields := strings.SplitN(line, ":", 2)
		if len(fields) < 2 {
			continue
		}
		key := strings.TrimSpace(fields[0])
		value := strings.TrimSpace(fields[1])

		switch key {
		case "Processor":
			processorName = value
		case "processor", "cpu number":
			if c.CPU >= 0 {
				finishCPUInfo(ctx, &c)
				ret = append(ret, c)
			}
			c = InfoStat{Cores: 1, ModelName: processorName}
			t, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return ret, err
			}
			c.CPU = int32(t)
		case "vendorId", "vendor_id":
			c.VendorID = value
			if strings.Contains(value, "S390") {
				processorName = "S390"
			}
		case "mvendorid":
			if !strings.HasPrefix(value, "0x") {
				continue
			}

			if v, err := strconv.ParseUint(value[2:], 16, 32); err == nil {
				switch v {
				case 0x31e:
					c.VendorID = "Andes"
				case 0x029:
					c.VendorID = "Microchip"
				case 0x127:
					c.VendorID = "MIPS"
				case 0x489:
					c.VendorID = "SiFive"
				case 0x5b7:
					c.VendorID = "T-Head"
				}
			}
		case "CPU implementer":
			if v, err := strconv.ParseUint(value, 0, 8); err == nil {
				switch v {
				case 0x41:
					c.VendorID = "ARM"
				case 0x42:
					c.VendorID = "Broadcom"
				case 0x43:
					c.VendorID = "Cavium"
				case 0x44:
					c.VendorID = "DEC"
				case 0x46:
					c.VendorID = "Fujitsu"
				case 0x48:
					c.VendorID = "HiSilicon"
				case 0x49:
					c.VendorID = "Infineon"
				case 0x4d:
					c.VendorID = "Motorola/Freescale"
				case 0x4e:
					c.VendorID = "NVIDIA"
				case 0x50:
					c.VendorID = "APM"
				case 0x51:
					c.VendorID = "Qualcomm"
				case 0x56:
					c.VendorID = "Marvell"
				case 0x61:
					c.VendorID = "Apple"
				case 0x69:
					c.VendorID = "Intel"
				case 0xc0:
					c.VendorID = "Ampere"
				}
			}
		case "cpu family", "marchid":
			c.Family = value
		case "model", "CPU part", "mimpid":
			c.Model = value
			// if CPU is arm based, model name is found via model number. refer to: arch/arm64/kernel/cpuinfo.c
			if c.VendorID == "ARM" {
				if v, err := strconv.ParseUint(c.Model, 0, 16); err == nil {
					modelName, exist := armModelToModelName[v]
					if exist {
						c.ModelName = modelName
					} else {
						c.ModelName = "Undefined"
					}
				}
			}
		case "Model Name", "model name", "cpu", "uarch":
			c.ModelName = value
			if strings.Contains(value, "POWER") {
				c.Model = strings.Split(value, " ")[0]
				c.Family = "POWER"
				c.VendorID = "IBM"
			}
		case "stepping", "revision", "CPU revision":
			val := value

			if key == "revision" {
				val = strings.Split(value, ".")[0]
			}

			if strings.EqualFold(val, "unknown") {
				continue
			}

			t, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return ret, err
			}
			c.Stepping = int32(t)
		case "cpu MHz", "clock", "cpu MHz dynamic":
			// treat this as the fallback value, thus we ignore error
			if t, err := strconv.ParseFloat(strings.Replace(value, "MHz", "", 1), 64); err == nil {
				c.Mhz = t
			}
		case "cache size":
			t, err := strconv.ParseInt(strings.Replace(value, " KB", "", 1), 10, 64)
			if err != nil {
				return ret, err
			}
			c.CacheSize = int32(t)
		case "physical id", "hart":
			c.PhysicalID = value
		case "core id":
			c.CoreID = value
		case "flags", "Features":
			c.Flags = strings.FieldsFunc(value, func(r rune) bool {
				return r == ',' || r == ' '
			})
		case "isa", "hart isa":
			if len(c.Flags) != 0 || !strings.HasPrefix(value, "rv64") {
				continue
			}
			c.Flags = riscvISAParse(value)
		case "microcode":
			c.Microcode = value
		}
	}
	if c.CPU >= 0 {
		finishCPUInfo(ctx, &c)
		ret = append(ret, c)
	}
	return ret, nil
}

func riscvISAParse(s string) []string {
	ext := strings.Split(s, "_")
	if len(ext[0]) <= 4 {
		return nil
	}
	// the base extensions must "rv64" prefix
	base := strings.Split(ext[0][4:], "")
	return append(base, ext[1:]...)
}
