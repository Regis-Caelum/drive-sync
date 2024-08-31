podman build -t drive-sync-rpm-builder .
podman run --name drive-sync-temp localhost/drive-sync-rpm-builder
podman cp drive-sync-temp:/app ./rpms
podman rm drive-sync-temp
