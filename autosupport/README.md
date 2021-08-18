# Harvest Autosupport

By default, Harvest sends basic poller information to NetApp on a daily
cadence. This behavior can be disabled by adding `autosupport_disabled: true` to
the `Tools` section of your `harvest.yml` file. e.g.

```
Tools:
  autosupport_disabled: true
```

This information is collected and sent via the `bin/asup` executable. Harvest
autosupport does not gather or transmit Personally Identifiable Information
(PII) or Personal Information. `bin/asup` comes with a
[EULA](https://www.netapp.com/us/media/enduser-license-agreement-worldwide.pdf)
that is not applicable to the other Harvest binaries in `./bin`. The EULA only applies to `bin/asup`. 

You can learn more about NetApp's commitment to data security and trust
[here](https://www.netapp.com/us/company/trust-center/index.aspx).

## Example Autosupport Information

An example payload sent by Harvest looks like this. You can see exactly what Harvest sends by checking the `./asup/payload/` directory.

```
{
  "Target": {
    "Version": "9.10.1",
    "Model": "cdot",
    "Serial": "1-80-000011",
    "Ping": 0,
    "ClusterUuid": "2c3877ad-7523-11eb-aafa-d039ea1fbe60"
  },
  "Nodes": {
    "Count": 4,
    "DataPoints": 68,
    "PollTime": 106202,
    "ApiTime": 104031,
    "ParseTime": 1127,
    "PluginTime": 20,
    "NodeUuid": [
      "8ac0a65e-7522-11eb-aafa-d039ea1fbe60",
      "88ef3daf-7522-11eb-bbf2-d039ea1febb4",
      "88f1555f-7522-11eb-8d8c-d039ea1fa41d",
      "88f7a00c-7522-11eb-9ca6-d039ea1fc4e0"
    ]
  },
  "Volumes": {
    "Count": 100,
    "DataPoints": 69,
    "PollTime": 1230,
    "ApiTime": 2380,
    "ParseTime": 10,
    "PluginTime": 10
  },
  "Platform": {
    "OS": "darwin",
    "Arch": "darwin",
    "MemoryKb": 1106248,
    "CPUs": 16
  },
  "Harvest": {
    "UUID": "32799326289865b63dfe03689909333437491b4e",
    "Version": "21.08.1809",
    "Release": "v21.05.4",
    "Commit": "06ef539",
    "BuildDate": "2021-08-18T09:46:40-0400",
    "NumClusters": 1
  }
}
```