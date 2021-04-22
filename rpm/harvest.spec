
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
export INSTALL_ROOT=$RPM_BUILD_ROOT
export INSTALL_TARGET=rpm
make install
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

%preun

%postun
make uninstall

%clean
echo "clean up ..."

