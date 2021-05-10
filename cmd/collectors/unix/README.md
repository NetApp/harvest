

# Unix Collector

## Overview

This collector polls resource usage by Harvest pollers on the local system.

### Supported platforms
Collector requires an OS where the `/proc/` filesystem is available. This includes most Unix/Unix-like systems:

* Android / Termux
* DragonFly BSD
* FreeBSD
* IBM AIX
* Linux
* NetBSD
* Plan9
* Solaris

(On FreeBSD and NetBSD the `/proc/`-filesystem needs to be manually mounted)

### Collected metrics

* `start_time`
* `cpu`
* `cpu_percent`
* `memory`
* `memory_percent`
* `io`
* `net`
* `ctx`
* `threads`
* `fds`
  
Note that `cpu`, `memory`, `io`, `net`, `ctx` are histograms.