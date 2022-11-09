NetApp Harvest requires login credentials to access monitored hosts. Although, a generic admin account can be used, it
is best practice to create a dedicated monitoring account with the least privilege access.

ONTAP 7-mode supports only username / password based authentication with NetApp Harvest.
Harvest communicates with monitored systems exclusively via HTTPS, which is not enabled by default in Data
ONTAP 7-mode. Login as a user with full administrative privileges and execute the following steps.

# Enabling HTTPS and TLS (ONTAP 7-mode only)

Verify SSL is configured

```
secureadmin status ssl
```

If ssl is ‘active’ continue. If not, setup SSL and be sure to choose a Key length (bits) of 2048:

```
secureadmin setup ssl
```

```
SSL Setup has already been done before. Do you want to proceed? [no] yes
Country Name (2 letter code) [US]: NL
State or Province Name (full name) [California]: Noord-Holland
Locality Name (city, town, etc.) [Santa Clara]: Schiphol
Organization Name (company) [Your Company]: NetApp
Organization Unit Name (division): SalesEngineering
Common Name (fully qualified domain name) [sdt-7dot1a.nltestlab.hq.netapp.com]:
Administrator email: noreply@netapp.com
Days until expires [5475] :5475 Key length (bits) [512] :2048
```

Enable management via SSL and enable TLS

```
options httpd.admin.ssl.enable on
options tls.enable on  
```

## Creating ONTAP user

### Create the role with required capabilities

```
role add netapp-harvest-role -c "Role for performance monitoring by NetApp Harvest" -a login-http-admin,api-system-get-version,api-system-get-info,api-perf-object-*,api-emsautosupport-log 
```

### Create a group for this role

```
useradmin group add netapp-harvest-group -c "Group for performance monitoring by NetApp Harvest" -r netapp-harvest-role 
```

### Create a user for the role and enter the password when prompted

```
useradmin user add netapp-harvest -c "User account for performance monitoring by NetApp Harvest" -n "NetApp Harvest" -g netapp-harvest-group
```

The user is now created and can be configured for use by NetApp Harvest.