%define app_home /opt/less/keeper
%define app_user lesskeeper
%define app_grp  lesskeeper

Name: lesskeeper
Version: x.y.z
Release: 1%{?dist}
Vendor: LessCompute.com
Summary: Distributed coordination service
License: Apache 2
Group: Applications
Source0: lesskeeper-x.y.z.tar.gz
BuildRoot:  %{_tmppath}/%{name}-%{version}-%{release}

Requires(pre):  shadow-utils
Requires(post): chkconfig

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
rm -rf %{buildroot}

%pre
# Add the "lesskeeper" user
getent group %{app_grp} >/dev/null || groupadd -r %{app_grp}
getent passwd %{app_user} >/dev/null || \
    useradd -r -g %{app_grp} -s /sbin/nologin \
    -d %{app_home} -c "lesskeeper user"  %{app_user}

if [ $1 == 2 ]; then
    service lesskeeper stop
fi

%post
# Register the lesskeeper service
if [ $1 -eq 1 ]; then
    /sbin/chkconfig --add lesskeeper
fi

if [ $1 -ge 1 ]; then
    service lesskeeper start
fi

%preun
if [ $1 = 0 ]; then
    /sbin/service lesskeeper stop  > /dev/null 2>&1
    /sbin/chkconfig --del lesskeeper
fi

%postun

%files
%defattr(-,root,root,-)
%dir %{app_home}

%{_initrddir}/%{name}
%config(noreplace) %{app_home}/etc/keeper.json
%config(noreplace) %{app_home}/etc/redis.conf
%{app_home}/
