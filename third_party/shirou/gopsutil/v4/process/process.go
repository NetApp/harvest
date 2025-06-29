// SPDX-License-Identifier: BSD-3-Clause
package process

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/netapp/harvest/v2/third_party/shirou/gopsutil/v4/cpu"
	"github.com/netapp/harvest/v2/third_party/shirou/gopsutil/v4/internal/common"
	"github.com/netapp/harvest/v2/third_party/shirou/gopsutil/v4/net"
)

var (
	invoke                 common.Invoker = common.Invoke{}
	ErrorNoChildren                       = errors.New("process does not have children") // Deprecated: ErrorNoChildren is never returned by process.Children(), check its returned []*Process slice length instead
	ErrorProcessNotRunning                = errors.New("process does not exist")
	ErrorNotPermitted                     = errors.New("operation not permitted")
)

type Process struct {
	Pid            int32 `json:"pid"`
	name           string
	status         string
	parent         int32
	parentMutex    sync.RWMutex // for windows ppid cache
	numCtxSwitches *NumCtxSwitchesStat
	uids           []uint32
	gids           []uint32
	groups         []uint32
	numThreads     int32
	memInfo        *MemoryInfoStat
	sigInfo        *SignalInfoStat
	createTime     int64

	lastCPUTimes *cpu.TimesStat
	lastCPUTime  time.Time

	tgid int32
}

// Process status
const (
	// Running marks a task a running or runnable (on the run queue)
	Running = "running"
	// Blocked marks a task waiting on a short, uninterruptible operation (usually I/O)
	Blocked = "blocked"
	// Idle marks a task sleeping for more than about 20 seconds
	Idle = "idle"
	// Lock marks a task waiting to acquire a lock
	Lock = "lock"
	// Sleep marks task waiting for short, interruptible operation
	Sleep = "sleep"
	// Stop marks a stopped process
	Stop = "stop"
	// Wait marks an idle interrupt thread (or paging in pre 2.6.xx Linux)
	Wait = "wait"
	// Zombie marks a defunct process, terminated but not reaped by its parent
	Zombie = "zombie"

	// Solaris states. See https://github.com/collectd/collectd/blob/1da3305c10c8ff9a63081284cf3d4bb0f6daffd8/src/processes.c#L2115
	Daemon   = "daemon"
	Detached = "detached"
	System   = "system"
	Orphan   = "orphan"

	UnknownState = ""
)

type OpenFilesStat struct {
	Path string `json:"path"`
	Fd   uint64 `json:"fd"`
}

type MemoryInfoStat struct {
	RSS    uint64 `json:"rss"`    // bytes
	VMS    uint64 `json:"vms"`    // bytes
	HWM    uint64 `json:"hwm"`    // bytes
	Data   uint64 `json:"data"`   // bytes
	Stack  uint64 `json:"stack"`  // bytes
	Locked uint64 `json:"locked"` // bytes
	Swap   uint64 `json:"swap"`   // bytes
}

type SignalInfoStat struct {
	PendingProcess uint64 `json:"pending_process"`
	PendingThread  uint64 `json:"pending_thread"`
	Blocked        uint64 `json:"blocked"`
	Ignored        uint64 `json:"ignored"`
	Caught         uint64 `json:"caught"`
}

type RlimitStat struct {
	Resource int32  `json:"resource"`
	Soft     uint64 `json:"soft"`
	Hard     uint64 `json:"hard"`
	Used     uint64 `json:"used"`
}

type IOCountersStat struct {
	// ReadCount is a number of read I/O operations such as syscalls.
	ReadCount uint64 `json:"readCount"`
	// WriteCount is a number of read I/O operations such as syscalls.
	WriteCount uint64 `json:"writeCount"`
	// ReadBytes is a number of all I/O read in bytes. This includes disk I/O on Linux and Windows.
	ReadBytes uint64 `json:"readBytes"`
	// WriteBytes is a number of all I/O write in bytes. This includes disk I/O on Linux and Windows.
	WriteBytes uint64 `json:"writeBytes"`
	// DiskReadBytes is a number of disk I/O write in bytes. Currently only Linux has this value.
	DiskReadBytes uint64 `json:"diskReadBytes"`
	// DiskWriteBytes is a number of disk I/O read in bytes.  Currently only Linux has this value.
	DiskWriteBytes uint64 `json:"diskWriteBytes"`
}

type NumCtxSwitchesStat struct {
	Voluntary   int64 `json:"voluntary"`
	Involuntary int64 `json:"involuntary"`
}

type PageFaultsStat struct {
	MinorFaults      uint64 `json:"minorFaults"`
	MajorFaults      uint64 `json:"majorFaults"`
	ChildMinorFaults uint64 `json:"childMinorFaults"`
	ChildMajorFaults uint64 `json:"childMajorFaults"`
}

// Resource limit constants are from /usr/include/x86_64-linux-gnu/bits/resource.h
// from libc6-dev package in Ubuntu 16.10
const (
	RLIMIT_CPU        int32 = 0
	RLIMIT_FSIZE      int32 = 1
	RLIMIT_DATA       int32 = 2
	RLIMIT_STACK      int32 = 3
	RLIMIT_CORE       int32 = 4
	RLIMIT_RSS        int32 = 5
	RLIMIT_NPROC      int32 = 6
	RLIMIT_NOFILE     int32 = 7
	RLIMIT_MEMLOCK    int32 = 8
	RLIMIT_AS         int32 = 9
	RLIMIT_LOCKS      int32 = 10
	RLIMIT_SIGPENDING int32 = 11
	RLIMIT_MSGQUEUE   int32 = 12
	RLIMIT_NICE       int32 = 13
	RLIMIT_RTPRIO     int32 = 14
	RLIMIT_RTTIME     int32 = 15
)

func (p Process) String() string {
	s, _ := json.Marshal(p)
	return string(s)
}

func (o OpenFilesStat) String() string {
	s, _ := json.Marshal(o)
	return string(s)
}

func (m MemoryInfoStat) String() string {
	s, _ := json.Marshal(m)
	return string(s)
}

func (r RlimitStat) String() string {
	s, _ := json.Marshal(r)
	return string(s)
}

func (i IOCountersStat) String() string {
	s, _ := json.Marshal(i)
	return string(s)
}

func (p NumCtxSwitchesStat) String() string {
	s, _ := json.Marshal(p)
	return string(s)
}

func PidsWithContext(ctx context.Context) ([]int32, error) {
	pids, err := pidsWithContext(ctx)
	sort.Slice(pids, func(i, j int) bool { return pids[i] < pids[j] })
	return pids, err
}

// NewProcess creates a new Process instance, it only stores the pid and
// checks that the process exists. Other method on Process can be used
// to get more information about the process. An error will be returned
// if the process does not exist.
func NewProcess(pid int32) (*Process, error) {
	return NewProcessWithContext(context.Background(), pid)
}

func NewProcessWithContext(ctx context.Context, pid int32) (*Process, error) {
	p := &Process{
		Pid: pid,
	}

	exists, err := PidExistsWithContext(ctx, pid)
	if err != nil {
		return p, err
	}
	if !exists {
		return p, ErrorProcessNotRunning
	}
	p.CreateTimeWithContext(ctx)
	return p, nil
}

func (p *Process) CreateTimeWithContext(ctx context.Context) (int64, error) {
	if p.createTime != 0 {
		return p.createTime, nil
	}
	createTime, err := p.createTimeWithContext(ctx)
	p.createTime = createTime
	return p.createTime, err
}

// Groups returns all group IDs(include supplementary groups) of the process as a slice of the int
func (p *Process) Groups() ([]uint32, error) {
	return p.GroupsWithContext(context.Background())
}

// Ppid returns Parent Process ID of the process.
func (p *Process) Ppid() (int32, error) {
	return p.PpidWithContext(context.Background())
}

// Parent returns parent Process of the process.
func (p *Process) Parent() (*Process, error) {
	return p.ParentWithContext(context.Background())
}

// ParentWithContext returns parent Process of the process.
func (p *Process) ParentWithContext(ctx context.Context) (*Process, error) {
	ppid, err := p.PpidWithContext(ctx)
	if err != nil {
		return nil, err
	}
	return NewProcessWithContext(ctx, ppid)
}

// Foreground returns true if the process is in foreground, false otherwise.
func (p *Process) Foreground() (bool, error) {
	return p.ForegroundWithContext(context.Background())
}

// Uids returns user ids of the process as a slice of the int
func (p *Process) Uids() ([]uint32, error) {
	return p.UidsWithContext(context.Background())
}

// Gids returns group ids of the process as a slice of the int
func (p *Process) Gids() ([]uint32, error) {
	return p.GidsWithContext(context.Background())
}

// Terminal returns a terminal which is associated with the process.
func (p *Process) Terminal() (string, error) {
	return p.TerminalWithContext(context.Background())
}

// Nice returns a nice value (priority).
func (p *Process) Nice() (int32, error) {
	return p.NiceWithContext(context.Background())
}

// IOCounters returns IO Counters.
func (p *Process) IOCounters() (*IOCountersStat, error) {
	return p.IOCountersWithContext(context.Background())
}

// MemoryInfo returns generic process memory information,
// such as RSS and VMS.
func (p *Process) MemoryInfo() (*MemoryInfoStat, error) {
	return p.MemoryInfoWithContext(context.Background())
}

// Children returns the children of the process represented as a slice
// of pointers to Process type.
func (p *Process) Children() ([]*Process, error) {
	return p.ChildrenWithContext(context.Background())
}

// Connections returns a slice of net.ConnectionStat used by the process.
// This returns all kind of the connection. This means TCP, UDP or UNIX.
func (p *Process) Connections() ([]net.ConnectionStat, error) {
	return p.ConnectionsWithContext(context.Background())
}

// ConnectionsMax returns a slice of net.ConnectionStat used by the process at most `max`.
func (p *Process) ConnectionsMax(maxConn int) ([]net.ConnectionStat, error) {
	return p.ConnectionsMaxWithContext(context.Background(), maxConn)
}

// SendSignal sends a unix.Signal to the process.
func (p *Process) SendSignal(sig Signal) error {
	return p.SendSignalWithContext(context.Background(), sig)
}

// Suspend sends SIGSTOP to the process.
func (p *Process) Suspend() error {
	return p.SuspendWithContext(context.Background())
}

// Resume sends SIGCONT to the process.
func (p *Process) Resume() error {
	return p.ResumeWithContext(context.Background())
}

// Terminate sends SIGTERM to the process.
func (p *Process) Terminate() error {
	return p.TerminateWithContext(context.Background())
}

// Kill sends SIGKILL to the process.
func (p *Process) Kill() error {
	return p.KillWithContext(context.Background())
}

// Username returns a username of the process.
func (p *Process) Username() (string, error) {
	return p.UsernameWithContext(context.Background())
}

// convertStatusChar as reported by the ps command across different platforms.
func convertStatusChar(letter string) string {
	// Sources
	// Darwin: http://www.mywebuniversity.com/Man_Pages/Darwin/man_ps.html
	// FreeBSD: https://www.freebsd.org/cgi/man.cgi?ps
	// Linux https://man7.org/linux/man-pages/man1/ps.1.html
	// OpenBSD: https://man.openbsd.org/ps.1#state
	// Solaris: https://github.com/collectd/collectd/blob/1da3305c10c8ff9a63081284cf3d4bb0f6daffd8/src/processes.c#L2115
	switch letter {
	case "A":
		return Daemon
	case "D", "U":
		return Blocked
	case "E":
		return Detached
	case "I":
		return Idle
	case "L":
		return Lock
	case "O":
		return Orphan
	case "R":
		return Running
	case "S":
		return Sleep
	case "T", "t":
		// "t" is used by Linux to signal stopped by the debugger during tracing
		return Stop
	case "W":
		return Wait
	case "Y":
		return System
	case "Z":
		return Zombie
	default:
		return UnknownState
	}
}
