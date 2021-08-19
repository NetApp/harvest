# Harvest Support and Getting Help

Harvest is an open-source project developed and published by NetApp to collect performance, capacity and hardware metrics from ONTAP clusters. These metrics can then be delivered to a range of databases and displayed in Grafana dashboards. Harvest is not an officially supported NetApp product. NetApp maintains and updates Harvest with bug fixes, security updates, and feature development. For assistance refer [Getting Help](#getting-help)

This document describes Harvest's release and support lifecycle as well as places you can get help.

## Harvest's Release and Support Lifecycle

Harvest's current release schedule is quarterly in January, April, July, and October, but it may change at our discretion. 

Each release of Harvest supports the most recently released version of ONTAP. We try our best to also support earlier versions of ONTAP. When that's not possible, breaking changes will be outlined in the changelog.

Harvest is constantly being improved with new features and bug fixes. Customers are encouraged to upgrade frequently to take advantage of these improvements.

`We intend to support each Harvest release for 12 months.`

For example, when `YY.MM` (ex: 21.04) is released, we intend to support it until `YY+1.MM` (ex: 22.04) is released. At the same time, `YY-1.MM` (ex: 20.04) and associated minor releases (ex: 20.04.1) move to limited or no support.

If you are running a version of Harvest that’s more than 12 months old, you must upgrade to a newer version to receive any support then available from NetApp. We always recommend running the latest version.

We use GitHub for tracking bugs and feature requests.

# Getting Help

There is a vibrant community of Harvest users on the `#harvest` channel in [NetApp's Slack team](https://netapppub.slack.com/archives/C02072M1UCD). Slack is a great place to ask general questions about the project and discuss related topics with like-minded peers.

## Slack

Join [thePub workspace](https://www.netapp.io/slack). After joining, click the `+` sign next to `Channels` and then click the `Browse Channels` button. Search for `harvest` from the Channel Browser and click `Join`.

![Join channel image](/docs/slack.png)

## Documentation

* [Harvest Documentation](README.md)
* [Harvest Architecture](ARCHITECTURE.md)
* [Contributing](CONTRIBUTING.md)
* [Wiki](https://github.com/NetApp/harvest/wiki)
