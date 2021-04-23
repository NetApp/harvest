
# version, release and arch will be pasted here
%define name harvest

Name: %{name}
Summary: Storage System Monitoring
Version: %{version}
Release: %{release}
Source: %{name}_%{version}-%{release}.tgz
License: (c) NetApp
URL: https://www.support.netapp.com
AutoReq: no
Provides: harvest
BuildRoot: %{buildroot}
BuildArch: %{arch}

%description
Description: NetApp Harvest 2 - Monitoring of Storage Systems
 This application collects data from NetApp systems,
 forwards it to the Prometheus time series db, and
 provides Grafana dashboards for visualization.

%prep

%pre

%install
echo " --> unpacking package..."
mkdir -p $RPM_BUILD_ROOT/opt/
tar xvfz /tmp/build/rpm/SOURCES/%{name}_%{version}-%{release}.tgz -C $RPM_BUILD_ROOT/opt/
cd $RPM_BUILD_ROOT/opt/harvest
ls -halt $RPM_BUILD_ROOT/opt/
ls -halt $RPM_BUILD_ROOT/opt/harvest
echo " --> installing harvest..."
cd $RPM_BUILD_ROOT/opt/harvest
mkdir -p $RPM_BUILD_ROOT/etc/harvest
mkdir -p $RPM_BUILD_ROOT/var/log/harvest
mkdir -p $RPM_BUILD_ROOT/var/run/harvest
mv conf/ grafana/ harvest.yml $RPM_BUILD_ROOT/etc/harvest/

%post
echo " --> create harvest group & user"
getent group harvest > /dev/null 2>&1 || groupadd -r harvest && echo " group created"
getent passwd harvest > /dev/null 2>&1 || useradd -r -M -g harvest --shell=/sbin/nologin harvest && echo " user created"
chown -R harvest:harvest $RPM_BUILD_ROOT/etc/harvest/
chown -R harvest:harvest $RPM_BUILD_ROOT/opt/harvest/
chown -R harvest:harvest $RPM_BUILD_ROOT/var/log/harvest/
chown -R harvest:harvest $RPM_BUILD_ROOT/var/run/harvest/
if [ -e /usr/local/bin/harvest ]; then
    unlink /usr/local/bin/harvest
fi
ln -s /opt/harvest/bin/harvest /usr/local/bin/harvest
echo " link created/updated"
echo " --> install complete!"

%preun

%postun
userdel harvest
groupdel harvest
echo "uninstall complete"

%clean

%files
/opt/harvest
/etc/harvest
/var/log/harvest
/var/run/harvest

