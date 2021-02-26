
# version, release and arch will be pasted here
%define name harvest

Name: %{name}
Summary: Storage System Monitoring
Version: %{version}
Release: %{release}
Source: %{name}_%{version}-%{version}.tgz
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
tar xvfz /tmp/build/rpm/SOURCES/%{name}_%{version}-%{version}.tgz
cd harvest
echo "executing install script"
export BUILD_ROOT=$RPM_BUILD_ROOT
./cmd/install.sh
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
harvest config welcome

%preun

%postun
./opt/harvest/cmd/install.sh uninstall

%clean
echo "clean up ..."

