## Version Names
Harvest releases use a calendar-based numbering scheme. Example: `YY.MM.bb`, where `YY` is the last two digits of the year of the release, `MM` is the month of release, and `bb` is the bugfix release.

There are four release types:
* Major release: `YY.MM` is updated to match year and month of release
* Bugfix, patch, or point-release: `bb` is incremented for each point release
* Release candidates (aka pre-releases): These are similar to point-releases, but with `-rc.num` suffixed to the point-release pattern
* Nightly: `YY.MM.ddHH-nightly`, where `dd` is day-of-month and `HH` is the hour-of-day (0-23)

Examples:

| Type    | Version              | Notes                                                            |
|:--------|:---------------------|:-----------------------------------------------------------------|
| Release | `21.05.0`            | a release in May 2021                                            |
| Release | `21.11.0`            | a release in Nov 2021                                            |
| Bugfix  | `21.11.1`            | `21.11.0 had a defect, this version fixes it                     |
| RC      | `21.05.0-rc1`        | First release candidate of `21.05.0`, 2nd would be `21.05.0-rc2` |
| Nightly | `21.05.1813-nightly` | Build of `21.05.0` release created on the 18th at 1 pm           |


### RPM and Debian Names
The RPM and Debian file formats require `Version` and `Release` fields. The [RPM packaging guide](https://rpm-packaging-guide.github.io/) says `Release` is
> The number of times this version of the software was released. Reset to 1 when a new Version of the software is built.

When the Harvest version changes, per the rules above, the `RPM Version` will follow in lock-step and in most cases the `RPM Release` will be 1. The exceptions are release candidates and nightly builds which will include that information in the `RPM Release` field as shown below.

Examples of RPM version/release and filename values based on the Harvest's version names.

| Harvest            | RPM Version | RPM Release | File name                             |
|:-------------------|:------------|:------------|:--------------------------------------|
| `21.05.0`          | `21.05.0`   | 1           | `harvest-21.05.0-1.x86_64.rpm`        |
| `21.11.1`          | `21.11.1`   | 1           | `harvest-21.11.1-1.x86_64.rpm`        |
| `21.05.0-rc1`      | `21.05.0`   | rc1         | `harvest-21.05.0-rc1.x86_64.rpm`      |
| `21.05.13-nightly` | `21.05.13`  | nightly     | `harvest-21.05.13-nightly.x86_64.rpm` |

Every release includes the whole project.

### Build RPM

Refer [RPM and Debian Names](#rpm-and-debian-names) for $VERSION $RELEASE

`./package.sh rpm x86_64 $VERSION $RELEASE`

### Build DEB

`./package.sh deb x86_64 $VERSION $RELEASE`

### Tar gz

`make package VERSION=$VERSION RELEASE=$RELEASE`