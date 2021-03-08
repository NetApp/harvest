
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
echo "unpacking tarball..."
tar xvfz /tmp/build/rpm/SOURCES/%{name}_%{version}-%{release}.tgz
cd harvest
echo "executing install script"
export BUILD_ROOT=$RPM_BUILD_ROOT
sh cmd/install.sh
echo "cleaning up..."
cd ..
rm -Rf harvest
exit

%files
/opt/harvest
/etc/harvest
/var/log/harvest
/var/run/harvest

%pre

%post
#!/bin/sh
ln -s /opt/harvest/bin/harvest /usr/local/bin/harvest
cd /opt/harvest

%preun

%postun
unlink /opt/harvest/bin/harvest
echo "uninstall complete"

%clean
echo "clean up ..."

