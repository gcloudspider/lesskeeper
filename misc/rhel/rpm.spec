%define app_home /opt/less/keeper

Name: less-keeper
Version: x.y.z
Release: 1%{?dist}
Vendor: LessCompute.com
Summary: Distributed coordination service
License: Apache 2
Group: Applications
Source0: less-keeper-x.y.z.tar.gz
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

install -m 0755 -p bin/less-keeper-store %{buildroot}%{app_home}/bin/less-keeper-store
install -m 0755 -p bin/less-keeper %{buildroot}%{app_home}/bin/less-keeper
cp -rp etc/keeper.json %{buildroot}%{app_home}/etc/keeper.json
cp -rp etc/redis.conf %{buildroot}%{app_home}/etc/redis.conf

install -m 0755 -p misc/rhel/init.d-scripts %{buildroot}%{_initrddir}/%{name}

%clean

%pre
if [ $1 == 2 ]; then
    service less-keeper stop
fi

%post

if [ $1 == 2 ]; then
    service less-keeper start
fi

%preun
if [ $1 = 0 ]; then
    service less-keeper stop
    chkconfig --del less-keeper
fi

%postun

%files
%defattr(-,root,root,-)
%dir %{app_home}
%{_initrddir}/%{name}
%config(noreplace) %{app_home}/etc/keeper.json
%config(noreplace) %{app_home}/etc/redis.conf
%{app_home}/
