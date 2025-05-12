## Prepare Cisco switch

NetApp Harvest requires login credentials to access Cisco switches. Although, a generic admin account can be used, it
is better to create a dedicated monitoring user with read-only permissions.

If you want to create a dedicated monitoring user for Harvest, follow the steps below.

1. ssh into the switch with a user than can create new users. e.g. `ssh admin@switch-ip`
2. Create a new user with read-only permissions by running the following commands. Replace password with a strong password.

```bash
configure terminal
username ro_user role network-operator password Netapp123
exit
```

## Enable NX-API on Cisco switch

NetApp Harvest uses NX-API to collect metrics from Cisco switches. You need to enable NX-API on the switch, follow the steps below.

1. ssh into the switch with a user than can enable NX-API. e.g. `ssh admin@switch-ip`
2. Enable NX-API by running the following commands:

```bash
configure terminal
feature nxapi
exit
```

## Reference

See [Configuring User Accounts and RBAC](https://www.cisco.com/c/en/us/td/docs/switches/datacenter/nexus9000/sw/93x/security/configuration/guide/b-cisco-nexus-9000-nx-os-security-configuration-guide-93x/b-cisco-nexus-9000-nx-os-security-configuration-guide-93x_chapter_01000.html)
for more information on Cisco NX-OS user accounts and RBAC.

See [NX-OS Programmability Guide](https://www.cisco.com/c/en/us/td/docs/switches/datacenter/nexus9000/sw/93x/progammability/guide/b-cisco-nexus-9000-series-nx-os-programmability-guide-93x/b-cisco-nexus-9000-series-nx-os-programmability-guide-93x_chapter_010011.html) for more information on the Cisco NX-API.