Get up and running with Harvest on your preferred platform.
We provide pre-compiled binaries for Linux, RPMs, Debs, as well 
as prebuilt container images for both [nightly](https://github.com/NetApp/harvest/releases/tag/nightly) 
and stable [releases](https://github.com/NetApp/harvest/releases).

- [Binaries for Linux](native.md)
- [RPM and Debs](package-managers.md)
- [Containers](containers.md)

## Nabox

Instructions on how to install Harvest via [NAbox](https://nabox.org/documentation/installation/).

## Source

To build Harvest from source code, first make sure you have a working Go environment 
with [version 1.19 or greater installed](https://golang.org/doc/install).

Clone the repo and build everything.

```
git clone https://github.com/NetApp/harvest.git
cd harvest
make build
bin/harvest version
```

If you're building on a Mac use `GOOS=darwin make build`

Checkout the `Makefile` for other targets of interest.
