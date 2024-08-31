Name:           drive-sync
Version:        1.43
Release:        1%{?dist}
Summary:        Drive Sync CLI and Daemon

%define __debug_install_post %{nil}
%global debug_package %{nil}

License:        MIT
URL:            https://github.com/Regis-Caelum/drive-sync
Source0:        %{name}-%{version}.tar.gz

BuildRequires:  gccgo
BuildRequires:  golang >= 1.23

%description
This package installs the Drive Sync CLI and a daemon for background synchronization with Google Drive.

%prep
%setup -q

echo "Directory tree of the source:"
find . -type d -print
go version

%build
#export PATH=$PATH:/usr/local/go/bin
# Build CLI
ls -a

#export GO111MODULE=on

# Build CLI with build-id
#go build -ldflags="-X main.buildid=%{version} -w -s" -o %{_builddir}/dsync ./cli/dsync
#go build -ldflags="-buildid=$(uuidgen)" -o %{_builddir}/dsync ./cli/dsync

echo "Built CLI binary:"
ls -l %{_builddir}/dsync

# Build Daemon with build-id
#go build -ldflags="-X main.buildid=%{version} -w -s" -o %{_builddir}/dsync-daemon ./daemon
#go build -ldflags="-buildid=$(uuidgen)" -o %{_builddir}/dsync-daemon ./daemon

echo "Built Daemon binary:"
ls -l %{_builddir}/dsync-daemon

%install
echo "Directory tree of the build root:"
find %{buildroot} -type d -print

# Make sure this path exists in the build directory
mkdir -p %{buildroot}/usr/local/bin
mkdir -p %{buildroot}/etc/systemd/system
mkdir -p %{buildroot}/var/lib/dsync

# Copy the built binaries to the build root
install -m 0755 ./builds/dsync %{buildroot}/usr/local/bin/dsync
install -m 0755 ./builds/dsync-daemon %{buildroot}/usr/local/bin/dsync-daemon

touch %{buildroot}/var/lib/dsync/database.sqlite
chmod 0660 %{buildroot}/var/lib/dsync/database.sqlite

# Copy binaries to the build root
#cp -a /app/rpmbuild/BUILDROOT/drive-sync-1.0-1.x86_64/usr/local/bin/* %{buildroot}/usr/local/bin/
cp ./service/dsync-daemon.service %{buildroot}/etc/systemd/system/
#cp /app/rpmbuild/BUILDROOT/drive-sync-1.0-1.x86_64/var/lib/dsync/database.sqlite %{buildroot}/var/lib/dsync/

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
* Sat Aug 31 2024 Inshal Khan <khanmf@rknec.edu> 1.43-1
- 

* Sat Aug 31 2024 Inshal Khan <khanmf@rknec.edu> 1.42-1
- 

* Sat Aug 31 2024 Inshal Khan <khanmf@rknec.edu> 1.41-1
- 

* Sat Aug 31 2024 Inshal Khan <khanmf@rknec.edu> 1.40-1
- 

* Sat Aug 31 2024 Inshal Khan <khanmf@rknec.edu> 1.39-1
- 

* Sat Aug 31 2024 Inshal Khan <khanmf@rknec.edu> 1.38-1
- 

* Sat Aug 31 2024 Inshal Khan <khanmf@rknec.edu> 1.37-1
- 

* Sat Aug 31 2024 Inshal Khan <khanmf@rknec.edu> 1.36-1
- 

* Sat Aug 31 2024 Inshal Khan <khanmf@rknec.edu> 1.35-1
- 

* Sat Aug 31 2024 Inshal Khan <khanmf@rknec.edu> 1.34-1
- 

* Sat Aug 31 2024 Inshal Khan <khanmf@rknec.edu> 1.33-1
- 

* Sat Aug 31 2024 Inshal Khan <khanmf@rknec.edu> 1.32-1
- 

* Sat Aug 31 2024 Inshal Khan <khanmf@rknec.edu> 1.31-1
- 

* Sat Aug 31 2024 Inshal Khan <khanmf@rknec.edu> 1.30-1
- 

* Sat Aug 31 2024 Inshal Khan <khanmf@rknec.edu> 1.29-1
- 

* Sat Aug 31 2024 Inshal Khan <khanmf@rknec.edu> 1.28-1
- 

* Sat Aug 31 2024 Inshal Khan <khanmf@rknec.edu> 1.27-1
- 

* Sat Aug 31 2024 Inshal Khan <khanmf@rknec.edu> 1.26-1
- 

* Sat Aug 31 2024 Inshal Khan <khanmf@rknec.edu> 1.25-1
- 

* Sat Aug 31 2024 Inshal Khan <khanmf@rknec.edu> 1.24-1
- 

* Sat Aug 31 2024 Inshal Khan <khanmf@rknec.edu> 1.23-1
- 

* Sat Aug 31 2024 Inshal Khan <khanmf@rknec.edu> 1.22-1
- 

* Sat Aug 31 2024 Inshal Khan <khanmf@rknec.edu> 1.21-1
- 

* Sat Aug 31 2024 Inshal Khan <khanmf@rknec.edu> 1.20-1
- 

* Sat Aug 31 2024 Inshal Khan <khanmf@rknec.edu> 1.19-1
- 

* Sat Aug 31 2024 Inshal Khan <khanmf@rknec.edu> 1.18-1
- Add generated files (khanmf@rknec.edu)

* Sat Aug 31 2024 Inshal Khan <khanmf@rknec.edu> 1.17-1
- 

* Sat Aug 31 2024 Inshal Khan <khanmf@rknec.edu> 1.16-1
- 

* Sat Aug 31 2024 Inshal Khan <khanmf@rknec.edu> 1.15-1
- 

* Sat Aug 31 2024 Inshal Khan <khanmf@rknec.edu> 1.14-1
- 

* Sat Aug 31 2024 Inshal Khan <khanmf@rknec.edu> 1.13-1
- Updated mod file (khanmf@rknec.edu)

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