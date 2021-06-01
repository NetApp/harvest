

# Unix

This collector polls resource usage by Harvest pollers on the local system. Collector might be extended in the future to monitor any local or remote process.

### Table of Contents
- [Target System](#target-system)
- [Requirements](#requirements)
- [Parameters](#parameters)
- [Metrics](#metrics)

## Target System
The machine where Harvest is running ("localhost").

## Requirements
Collector requires any OS where the  proc-filesystem is available. If you are a developer, you are welcome to add support for other platforms. Currently, supported platforms includes most Unix/Unix-like systems:

* Android / Termux
* DragonFly BSD
* FreeBSD
* IBM AIX
* Linux
* NetBSD
* Plan9
* Solaris

(On FreeBSD and NetBSD the proc-filesystem needs to be manually mounted).

## Parameters

| parameter              | type         | description                                      | default                |
|------------------------|--------------|--------------------------------------------------|------------------------|
| `mount_point`      | string, optional | path to the `proc` filesystem                    | `/proc                 |

## Metrics

The Collector follows [the Linux proc(5) manual](https://man7.org/linux/man-pages/man5/procfs.5.html) to parse a static set of metrics. Unless otherwise stated, the metric has a scalar value:

| metric             | type                       | unit          | description                                              |
|--------------------|----------------------------|---------------|----------------------------------------------------------|
| `start_time`       | counter, `float64`         | seconds       | process uptime                                           |
| `cpu_percent`      | gauge, `float64`           | percent       | CPU used since last poll                                 |
| `memory_percent`   | gauge, `float64`           | percent       | Memory used (RSS) since last poll                        |
| `cpu`              | histogram, `float64`       | seconds       | CPU used since last poll (`system`, `user`, `iowait`)    |
| `memory`           | histogram, `uint64`        | kB            | Memory used since last poll (`rss`, `vms`, `swap`, etc)  |
| `io` | histogram, `uint64`  | <br>byte<br>count  | IOs performed by process:<br>`rchar`, `wchar`, `read_bytes`, `write_bytes` - read/write IOs<br>`syscr`, `syscw` - syscalls for IO operations  |
| `net`              | histogram, `uint64`        | count/byte    | Different IO operations over network devices  |
| `ctx`              | histogram, `uint64`        | count         | Number of context switched (`voluntary`, `involuntary`)  |
| `threads`          | counter, `uint64`          | count         | Number of threads                                        |
| `fds`              | counter, `uint64`          | count         | Number of file descriptors                               |
  

Additionally, the collector provides the following instance labels:

| label             | description                                              |
|-------------------|----------------------------------------------------------|
| poller            | name of the poller                                       |
| pid               | PID of the poller                                        |