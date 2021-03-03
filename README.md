# Coriolis OVM exporter

This service aims to augment Oracle VM with efficient incremental backup exports of virtual machine disks.

## Compiling the binary

The simplest way to build the binary is to have docker installed and simply run:

```bash
make
```

This will create a build image called ```coriolis-ovm-exporter-builder``` from the alpine golang image. This new image will have libdb recompiled using ```--enable-static```. It will then build ```coriolis-ovm-exporter``` as a statically linked binary. The resulting binary will be copied to your current working directory.

## Configuration

The configuration file is quite simple:

```toml
# Path on disk to the database file the exporter will use.
db_file = "/etc/coriolis-ovm-exporter/exporter.db"
# This is the base URL to your OVM manager
ovm_endpoint = "https://your-ovm-api-server.example:7002"

[jwt]
# Obviously, this needs to be changed :-)
secret = "yoidthOcBauphFeykCotdidNorjAnAtGhonsabShegAtfexbavlyakPak4SletEd"

[api]
bind = "0.0.0.0"
port = 5544
    [api.tls]
    # These settings are required
    certificate = "/tmp/certs/srv-pub.pem"
    key = "/tmp/certs/srv-key.pem"
    ca_certificate = "/tmp/certs/ca-pub.pem"
```