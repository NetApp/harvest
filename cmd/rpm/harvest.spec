%define name harvest

Name: %{name}
Summary: Storage System Monitoring
Version: %{version}
Release: 1
Source: %{name}_%{version}.tgz
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
 provides default dashboards for visualization.

%prep
echo "unpacking tarball..."
mkdir -p $RPM_BUILD_ROOT/opt/
tar xvfz /tmp/build/rpm/SOURCES/%{name}_%{version}.tgz
mv "harvest2" $RPM_BUILD_ROOT/opt/
exit

%files
%attr(0744, root, root) /opt/harvest2

%pre

%post
echo "install..."
ln -s /opt/harvest2/bin/harvest /usr/local/bin/harvest
echo "complete!"
harvest config welcome

%preun

%postun
echo "uninstall..."
unlink /usr/local/bin/harvest
rm -Rf /opt/harvest2
echo "removed from system"

%clean
echo "clean up ..."

