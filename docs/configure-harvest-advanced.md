This chapter describes additional advanced configuration possibilities of NetApp Harvest. For a typical installation, 
this level of detail is likely not needed.

## Variable Expansion

The `harvest.yml` configuration file supports variable expansion.
This allows you to use environment variables in the configuration file.
Harvest will expand strings with the format `$__env{VAR}` or `${VAR}`,
replacing the variable `VAR` with the value of the environment variable.
If the environment variable is not set, the variable will be replaced with an empty string.

Here's an example snippet from `harvest.yml`:

```yaml
Pollers:
  netapp_frankfurt:
    addr: 10.0.1.2
    username: $__env{NETAPP_FRANKFURT_RO_USER}
  netapp_london:
    addr: uk-cluster
    username: ${NETAPP_LONDON_RO_USER}
  netapp_rtp:
    addr: 10.0.1.4
    username: $__env{NETAPP_RTP_RO_USER}
```

If you set the environment variable `NETAPP_FRANKFURT_RO_USER` to `harvest1` and `NETAPP_LONDON_RO_USER` to `harvest2`,
the configuration will be expanded to:

```yaml
Pollers:
  netapp_frankfurt:
    addr: 10.0.1.2
    username: harvest1
  netapp_london:
    addr: uk-cluster
    username: harvest2
  netapp_rtp:
    addr: 10.0.1.4
    username: 
```