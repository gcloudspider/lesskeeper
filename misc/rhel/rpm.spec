%define app_home /opt/less/keeper

Name: lesskeeper
Version: x.y.z
Release: 1%{?dist}
Vendor: LessCompute.com
Summary: Distributed coordination service
License: Apache 2
Group: Applications
Source0: lesskeeper-x.y.z.tar.gz
BuildRoot:  %{_tmppath}/%{name}-%{version}-%{release}

%description
%prep
%setup  -q -n %{name}-%{version}
%build

%install
rm -rf %{buildroot}
install -d %{buildroot}%{app_home}/var
install -d %{buildroot}%{app_home}/bin
install -d %{buildroot}%{app_home}/etc
install -d %{buildroot}%{_initrddir}

install -m 0755 -p bin/lesskeeper-store %{buildroot}%{app_home}/bin/lesskeeper-store
install -m 0755 -p bin/lesskeeper %{buildroot}%{app_home}/bin/lesskeeper
cp -rp etc/keeper.json %{buildroot}%{app_home}/etc/keeper.json
cp -rp etc/redis.conf %{buildroot}%{app_home}/etc/redis.conf

install -m 0755 -p misc/rhel/init.d-scripts %{buildroot}%{_initrddir}/%{name}

%clean

%pre
if [ $1 == 2 ]; then
    service lesskeeper stop
fi

%post

if [ $1 == 2 ]; then
    service lesskeeper start
fi

%preun
if [ $1 = 0 ]; then
    service lesskeeper stop
    chkconfig --del lesskeeper
fi

%postun

%files
%defattr(-,root,root,-)
%dir %{app_home}
%{_initrddir}/%{name}
%config(noreplace) %{app_home}/etc/keeper.json
%config(noreplace) %{app_home}/etc/redis.conf
%{app_home}/
