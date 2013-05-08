%define app_home /opt/%{name}

Name: h5keeper
Version: x.y.z
Release: 1%{?dist}
Vendor: Hooto
Summary: Distributed coordination service
License: Apache 2
Group: Applications
BuildRequires: gcc
Source0: h5keeper-x.y.z.tar.gz
BuildRoot:  %{_tmppath}/%{name}-%{version}-%{release}

%description
%prep
%setup  -q -n %{name}-%{version}
%build

%install
rm -rf %{buildroot}
install -d %{buildroot}%{app_home}/data
install -d %{buildroot}%{app_home}/bin
install -d %{buildroot}%{app_home}/etc

#cp -rp ./* %{buildroot}%{app_home}/

install -m 0755 -p bin/redis-server %{buildroot}%{app_home}/bin/redis-server
install -m 0755 -p bin/h5keeper %{buildroot}%{app_home}/bin/h5keeper

%clean

%pre
if [ $1 == 2 ]; then
    service h5keeper stop
fi

%post

if [ $1 == 2 ]; then
    service h5keeper start
fi

%preun
if [ $1 = 0 ]; then
    service h5keeper stop
    chkconfig --del h5keeper
fi

%postun

%files
%defattr(-,root,root,-)
%dir %{app_home}
%{app_home}/
