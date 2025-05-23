# Harvest Support and Getting Help

Harvest is an open-source project developed and published by NetApp to collect performance, capacity and hardware metrics from ONTAP clusters. These metrics can then be delivered to a range of databases and displayed in Grafana dashboards. Harvest is not an officially supported NetApp product. NetApp maintains and updates Harvest with bug fixes, security updates, and feature development. For assistance, refer to [Getting Help](#getting-help)

This document describes Harvest's release and support lifecycle as well as places you can get help.

## Harvest's Release and Support Lifecycle

Harvest's current release schedule is quarterly in February, May, August, and November, but it may change at our discretion. 

Each release of Harvest supports the most recently released version of ONTAP. We try our best to also support earlier versions of ONTAP. When that's not possible, breaking changes will be outlined in the changelog.

Harvest is constantly being improved with new features and bug fixes. Customers are encouraged to upgrade frequently to take advantage of these improvements.

`We intend to support each Harvest release for 12 months.`

For example, when `YY.MM` (ex: 21.04) is released, we intend to support it until `YY+1.MM` (ex: 22.04) is released. At the same time, `YY-1.MM` (ex: 20.04) and associated minor releases (ex: 20.04.1) move to limited or no support.

If you are running a version of Harvest that’s more than 12 months old, you must upgrade to a newer version to receive support from NetApp. We always recommend running the latest version.

We use GitHub for tracking bugs and feature requests.

# Getting Help

The fastest way to ask a question or discuss Harvest related topics is via NetApp's Discord server or 
[GitHub discussions](https://github.com/NetApp/harvest/discussions).

## Discord

1. [Join and review the community guidelines](https://discord.gg/NetApp)
2. After you've joined the server, ask your question in the 
   [#harvest](https://discord.com/channels/855068651522490400/1062050414146625536) channel.

## Documentation

* [Harvest Documentation](README.md)
* [Harvest Architecture](ARCHITECTURE.md)
* [Contributing](CONTRIBUTING.md)
* [Wiki](https://github.com/NetApp/harvest/wiki)
