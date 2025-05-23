// SPDX-License-Identifier: BSD-3-Clause
//go:build darwin

package process

import (
	"context"
	"golang.org/x/sys/unix"
	"sort"
	"strconv"
	"strings"

	"github.com/netapp/harvest/v2/third_party/shirou/gopsutil/v4/internal/common"
	"github.com/netapp/harvest/v2/third_party/shirou/gopsutil/v4/net"
)

type _Ctype_struct___0 struct { //nolint:revive //FIXME
	Pad uint64
}

func pidsWithContext(_ context.Context) ([]int32, error) {
	var ret []int32

	kprocs, err := unix.SysctlKinfoProcSlice("kern.proc.all")
	if err != nil {
		return ret, err
	}

	for _, proc := range kprocs {
		ret = append(ret, int32(proc.Proc.P_pid))
	}

	return ret, nil
}

func (p *Process) PpidWithContext(_ context.Context) (int32, error) {
	k, err := p.getKProc()
	if err != nil {
		return 0, err
	}

	return k.Eproc.Ppid, nil
}

func (p *Process) createTimeWithContext(_ context.Context) (int64, error) {
	k, err := p.getKProc()
	if err != nil {
		return 0, err
	}

	return k.Proc.P_starttime.Sec*1000 + int64(k.Proc.P_starttime.Usec)/1000, nil
}

func (p *Process) ForegroundWithContext(ctx context.Context) (bool, error) {
	// see https://github.com/shirou/gopsutil/issues/596#issuecomment-432707831 for implementation details
	pid := p.Pid
	out, err := invoke.CommandWithContext(ctx, "ps", "-o", "stat=", "-p", strconv.Itoa(int(pid)))
	if err != nil {
		return false, err
	}
	return strings.IndexByte(string(out), '+') != -1, nil
}

func (p *Process) UidsWithContext(_ context.Context) ([]uint32, error) {
	k, err := p.getKProc()
	if err != nil {
		return nil, err
	}

	// See: http://unix.superglobalmegacorp.com/Net2/newsrc/sys/ucred.h.html
	userEffectiveUID := uint32(k.Eproc.Ucred.Uid)

	return []uint32{userEffectiveUID}, nil
}

func (p *Process) GidsWithContext(_ context.Context) ([]uint32, error) {
	k, err := p.getKProc()
	if err != nil {
		return nil, err
	}

	gids := make([]uint32, 0, 3)
	gids = append(gids, uint32(k.Eproc.Pcred.P_rgid), uint32(k.Eproc.Pcred.P_rgid), uint32(k.Eproc.Pcred.P_svgid))

	return gids, nil
}

func (p *Process) GroupsWithContext(_ context.Context) ([]uint32, error) {
	return nil, common.ErrNotImplementedError
	// k, err := p.getKProc()
	// if err != nil {
	// 	return nil, err
	// }

	// groups := make([]int32, k.Eproc.Ucred.Ngroups)
	// for i := int16(0); i < k.Eproc.Ucred.Ngroups; i++ {
	// 	groups[i] = int32(k.Eproc.Ucred.Groups[i])
	// }

	// return groups, nil
}

func (p *Process) TerminalWithContext(_ context.Context) (string, error) {
	return "", common.ErrNotImplementedError
	/*
		k, err := p.getKProc()
		if err != nil {
			return "", err
		}

		ttyNr := uint64(k.Eproc.Tdev)
		termmap, err := getTerminalMap()
		if err != nil {
			return "", err
		}

		return termmap[ttyNr], nil
	*/
}

func (p *Process) NiceWithContext(_ context.Context) (int32, error) {
	k, err := p.getKProc()
	if err != nil {
		return 0, err
	}
	return int32(k.Proc.P_nice), nil
}

func (p *Process) IOCountersWithContext(_ context.Context) (*IOCountersStat, error) {
	return nil, common.ErrNotImplementedError
}

func (p *Process) ChildrenWithContext(ctx context.Context) ([]*Process, error) {
	procs, err := ProcessesWithContext(ctx)
	if err != nil {
		return nil, nil
	}
	ret := make([]*Process, 0, len(procs))
	for _, proc := range procs {
		ppid, err := proc.PpidWithContext(ctx)
		if err != nil {
			continue
		}
		if ppid == p.Pid {
			ret = append(ret, proc)
		}
	}
	sort.Slice(ret, func(i, j int) bool { return ret[i].Pid < ret[j].Pid })
	return ret, nil
}

func (p *Process) ConnectionsWithContext(ctx context.Context) ([]net.ConnectionStat, error) {
	return net.ConnectionsPidWithContext(ctx, "all", p.Pid)
}

func (p *Process) ConnectionsMaxWithContext(ctx context.Context, maxConn int) ([]net.ConnectionStat, error) {
	return net.ConnectionsPidMaxWithContext(ctx, "all", p.Pid, maxConn)
}

func ProcessesWithContext(ctx context.Context) ([]*Process, error) {
	out := []*Process{}

	pids, err := PidsWithContext(ctx)
	if err != nil {
		return out, err
	}

	for _, pid := range pids {
		p, err := NewProcessWithContext(ctx, pid)
		if err != nil {
			continue
		}
		out = append(out, p)
	}

	return out, nil
}

// Returns a proc as defined here:
// http://unix.superglobalmegacorp.com/Net2/newsrc/sys/kinfo_proc.h.html
func (p *Process) getKProc() (*unix.KinfoProc, error) {
	return unix.SysctlKinfoProc("kern.proc.pid", int(p.Pid))
}

func (p *Process) MemoryInfoWithContext(ctx context.Context) (*MemoryInfoStat, error) {

	// Harvest added to eliminate the need for ebitengine/purego

	mem := &MemoryInfoStat{}
	out, err := invoke.CommandWithContext(ctx, "ps", "-o", "rss=,vsz=,nswap=", "-p", strconv.Itoa(int(p.Pid)))
	if err != nil {
		return nil, err
	}

	fields := strings.Fields(strings.TrimRight(string(out), "\n"))
	if len(fields) != 3 {
		return nil, common.ErrNotImplementedError
	}
	rss, err := strconv.ParseUint(fields[0], 10, 64)
	if err != nil {
		return nil, err
	}
	vsz, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return nil, err
	}
	if fields[2] == "-" {
		mem.Swap = 0
	} else {
		swap, err := strconv.ParseUint(fields[2], 10, 64)
		if err != nil {
			return nil, err
		}
		mem.Swap = swap
	}

	// convert to bytes
	mem.RSS = rss * 1024
	mem.VMS = vsz * 1024

	return mem, nil
}
