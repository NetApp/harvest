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

To build Harvest from source code follow these steps.

1. `git clone https://github.com/NetApp/harvest.git`
2. cd `harvest`
3. check the version of go required in the `go.mod` file
4. ensure you have a working Go environment at that version or newer. Go installs found [here](https://golang.org/doc/install). 
5. `make build` (if you want to run Harvest from a Mac use `GOOS=darwin make build`) 
6. `bin/harvest version`

Checkout the `Makefile` for other targets of interest.
