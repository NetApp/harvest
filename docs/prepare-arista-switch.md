## Prepare Arista switch

NetApp Harvest requires login credentials to access Arista switches. Although, a generic admin account can be used, it
is better to create a dedicated monitoring user with read-only permissions.

If you want to create a dedicated monitoring user for Harvest, follow the steps below.

1. ssh into the switch with a user than can create new users. e.g. `ssh admin@switch-ip`
2. Create a new user with read-only permissions by running the following commands. Replace password with a strong password.

```bash
enable
configure
username ro_user privilege 1 role network-operator secret Netapp123
exit
```

## Enable eAPI on Arista switch

NetApp Harvest uses eAPI to collect metrics from Arista switches. You need to enable eAPI on the switch, follow the steps below.

1. ssh into the switch with a user than can enable eAPI. e.g. `ssh admin@switch-ip`
2. Enable eAPI by running the following commands:

```bash
enable
configure
management api http-commands
protocol https
no shutdown
exit
```

## Reference

See the [Arista EOS User Manual](https://www.arista.com/en/um-eos/eos-user-security), for more information on 
Arista EOS user accounts and roles.

See the [Arista EOS User Manual](https://www.arista.com/en/um-eos/eos-session-management-commands), for more
information on the Arista eAPI.
