Name:           drive-sync
Version:        1.12
Release:        1%{?dist}
Summary:        Drive Sync CLI and Daemon

License:        MIT
URL:            https://github.com/Regis-Caelum/drive-sync
Source0:        %{name}-%{version}.tar.gz

%description
This package installs the Drive Sync CLI and a daemon for background synchronization with Google Drive.

%prep
%setup -q
cat go.mod

%build
# Build CLI
ls -a

export GO111MODULE=on

# Build CLI
go build -o %{_builddir}/dsync cli/dsync

# Build Daemon
go build -o %{_builddir}/dsync-daemon daemon

%install
echo "Directory tree of the build root:"
find %{buildroot} -type d -print

# Make sure this path exists in the build directory
mkdir -p %{buildroot}/usr/local/bin
mkdir -p %{buildroot}/etc/systemd/system
mkdir -p %{buildroot}/var/lib/dsync

# Copy the built binaries to the build root
install -m 0755 cli/dsync/dsync %{buildroot}/usr/local/bin/dsync
install -m 0755 daemon/dsync-daemon %{buildroot}/usr/local/bin/dsync-daemon

touch %{buildroot}/var/lib/dsync/database.sqlite
chmod 0660 %{buildroot}/var/lib/dsync/database.sqlite

# Copy binaries to the build root
cp -a /app/rpmbuild/BUILDROOT/drive-sync-1.0-1.x86_64/usr/local/bin/* %{buildroot}/usr/local/bin/
cp /app/rpmbuild/BUILDROOT/drive-sync-1.0-1.x86_64/etc/systemd/system/dsync-daemon.service %{buildroot}/etc/systemd/system/
cp /app/rpmbuild/BUILDROOT/drive-sync-1.0-1.x86_64/var/lib/dsync/database.sqlite %{buildroot}/var/lib/dsync/

%files
/usr/local/bin/dsync
/usr/local/bin/dsync-daemon
/etc/systemd/system/dsync-daemon.service
/var/lib/dsync/database.sqlite

%post
systemctl daemon-reload
systemctl enable dsync-daemon.service
systemctl start dsync-daemon.service


%preun
if [ $1 -eq 0 ]; then
    systemctl stop dsync-daemon.service
    systemctl disable dsync-daemon.service
    rm -f /etc/systemd/system/dsync-daemon.service
    rm -rf %{buildroot}/var/lib/dsync
fi

%postun
systemctl daemon-reload
pkill dsync-daemon

%changelog
* Sat Aug 31 2024 Inshal Khan <khanmf@rknec.edu> 1.12-1
- 

* Sat Aug 31 2024 Inshal Khan <khanmf@rknec.edu> 1.11-1
- 

* Sat Aug 31 2024 Inshal Khan <khanmf@rknec.edu> 1.10-1
- 

* Sat Aug 31 2024 Inshal Khan <khanmf@rknec.edu> 1.9-1
- 

* Sat Aug 31 2024 Inshal Khan <khanmf@rknec.edu> 1.8-1
- 

* Sat Aug 31 2024 Inshal Khan <khanmf@rknec.edu> 1.7-1
- 

* Sat Aug 31 2024 Inshal Khan <khanmf@rknec.edu> 1.6-1
- 

* Sat Aug 31 2024 Inshal Khan <khanmf@rknec.edu> 1.5-1
- 

* Sat Aug 31 2024 Inshal Khan <khanmf@rknec.edu> 1.4-1
- 

* Sat Aug 31 2024 Inshal Khan <khanmf@rknec.edu> 1.3-1
- 

* Sat Aug 31 2024 Inshal Khan <khanmf@rknec.edu> 1.2-1
- 

* Sat Aug 31 2024 Inshal Khan <khanmf@rknec.edu> 1.1-1
- 

* Sat Aug 31 2024 Inshal Khan <khanmf@rknec.edu>
- 

* Sat Aug 31 2024 Inshal Khan <khanmf@rknec.edu>
- 

* Sat Aug 31 2024 Inshal Khan <khanmf@rknec.edu>
- 

* Sat Aug 31 2024 Inshal Khan <khanmf@rknec.edu>
- 

* Sat Aug 31 2024 Inshal Khan <khanmf@rknec.edu> 1.0-1
- 

* Sat Aug 31 2024 Inshal Khan <khanmf@rknec.edu> 1.0.0-1
- new package built with tito

* Fri Aug 30 2024 Inshal Khan <khanmf@rknec.edu> 1.0-1
- Initial package