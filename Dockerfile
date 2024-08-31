FROM fedora:latest as builder
LABEL authors="regis"

# Install necessary tools
RUN dnf install -y wget tar rpm-build make rpmlint

# Install Go 1.23
RUN wget https://go.dev/dl/go1.23.0.linux-amd64.tar.gz \
    && tar -C /usr/local -xzf go1.23.0.linux-amd64.tar.gz \
    && rm go1.23.0.linux-amd64.tar.gz

RUN dnf install -y gccgo

# Set up Go environment
ENV PATH="/usr/local/go/bin:${PATH}"

# Verify Go installation
RUN go version

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files first to leverage Docker caching
COPY go.mod go.sum ./

# Download dependencies only when go.mod or go.sum change
RUN go mod tidy && go mod download

# Copy the project files into the container
COPY . .

# Create necessary directories for RPM build
RUN mkdir -p rpmbuild/{BUILD,RPMS,SOURCES,SPECS,SRPMS} \
    && mkdir -p rpmbuild/BUILDROOT/drive-sync-1.0-1.x86_64/usr/local/bin \
    && mkdir -p rpmbuild/BUILDROOT/drive-sync-1.0-1.x86_64/etc/systemd/system \
    && mkdir -p rpmbuild/BUILDROOT/drive-sync-1.0-1.x86_64/var/lib/dsync

RUN tar -cvzf rpmbuild/SOURCES/drive-sync-1.0.tar.gz .

# Build the CLI binary
RUN cd cli/dsync \
    && go build -o /app/rpmbuild/BUILDROOT/drive-sync-1.0-1.x86_64/usr/local/bin/dsync

# Build the daemon binary
RUN cd /app/daemon \
    && go build -o /app/rpmbuild/BUILDROOT/drive-sync-1.0-1.x86_64/usr/local/bin/dsync-daemon

# Initialize the database file
RUN touch /app/rpmbuild/BUILDROOT/drive-sync-1.0-1.x86_64/var/lib/dsync/database.sqlite \
    && chmod 0660 /app/rpmbuild/BUILDROOT/drive-sync-1.0-1.x86_64/var/lib/dsync/database.sqlite

# Copy the systemd service file into the RPM structure
RUN cp /app/service/dsync-daemon.service /app/rpmbuild/BUILDROOT/drive-sync-1.0-1.x86_64/etc/systemd/system/

# Create the RPM spec file
RUN cat <<EOL > /app/rpmbuild/SPECS/drive-sync.spec
Name:           drive-sync
Version:        1.0
Release:        1%{?dist}
Summary:        Drive Sync CLI and Daemon

License:        Your License
Source0:        %{name}-%{version}.tar.gz

%description
This package installs the Drive Sync CLI and a daemon for background synchronization with Google Drive.

%prep

%build

%install
echo "Directory tree of the build root:"
find %{buildroot} -type d -print

# Make sure this path exists in the build directory
mkdir -p %{buildroot}/usr/local/bin
mkdir -p %{buildroot}/etc/systemd/system
mkdir -p %{buildroot}/var/lib/dsync

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
* Fri Aug 30 2024 Inshal Khan <khanmf@rknec.edu> 1.0-1
- Initial package
EOL

# Build the RPM package
RUN rpmbuild -ba /app/rpmbuild/SPECS/drive-sync.spec --define "_topdir /app/rpmbuild"

# Final stage to extract the RPM from the build image
FROM alpine
WORKDIR /app
COPY --from=builder /app/rpmbuild/RPMS/x86_64/ .
