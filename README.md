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

# This is the base URL to your OVM manager. We will use this to
# authenticate login requests. Make sure this matches the manager
# that this node belongs to.
#
# NOTE: Setting this is *OPTIONAL*. If this setting is omitted, the
# exporter will attempt to fetch the manager IP from the ovs-agent
# database.
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

## API usage

### Authentication

Coriolis OVM exporter validates access credentials against the management API of the OVM deployment. You must have a valid username and password that can access the usual OVM console, to authenticate against the exporter.

To access the API, you must first fetch a [JWT token](https://jwt.io/) by logging into the service.

```
GET /api/v1/auth/login
```

Example usage:

```bash
curl -s -k -X POST -d \
    '{"username": "admin", "password": "SuperSecret"}' \
    https://10.107.8.20:5544/api/v1/auth/login/ | jq
{
  "token": "eyJhbGciOiJI ... Bn-KpcBo82IFnU"
}
```

### Fetch all VMs

```
GET /api/v1/vms
```

Example usage:

```bash
curl -s -k -X GET -H 'Accept: application/json' \
    -H "Authorization: Bearer TOKEN_GOES_HERE" \
    https://10.107.8.20:5544/api/v1/vms | jq
[
  {
    "friendly_name": "example-vm",
    "name": "0004fb0000060000ccaf98a0baa2c186",
    "uuid": "0004fb00-0006-0000-ccaf-98a0baa2c186",
    "disks": [
      {
        "name": "0004fb000012000005bcf25e906ce843.img",
        "path": "/OVS/Repositories/0004fb00000300006a09d4e1065041cb/VirtualDisks/0004fb000012000005bcf25e906ce843.img",
        "device_name": "xvda",
        "mode": "w"
      }
    ],
    "snapshots": [
      "9633d114-8270-41eb-bffc-67cc342957d9",
      "381ba26b-4a62-4baf-a450-af120ceddcbf"
    ]
  },
  .... truncated for clarity ...
  {
    "friendly_name": "another-vm",
    "name": "0004fb00000600000595d9e2d42334ec",
    "uuid": "0004fb00-0006-0000-0595-d9e2d42334ec",
    "disks": [
      {
        "name": "0004fb0000120000510f357e96bd5290.img",
        "path": "/OVS/Repositories/0004fb00000300009d120c269f2d8e1e/VirtualDisks/0004fb0000120000510f357e96bd5290.img",
        "device_name": "xvda",
        "mode": "w"
      },
      {
        "name": "0004fb0000150000aae1c593a64762c2.iso",
        "path": "/OVS/Repositories/0004fb0000030000d42b204197e41e75/ISOs/0004fb0000150000aae1c593a64762c2.iso",
        "device_name": "xvdb:cdrom",
        "mode": "r"
      }
    ],
    "snapshots": []
  }
]
```

### Fetch info about one VM

```
GET /api/v1/vms/{vmID}/
```

Example usage:

```bash
curl -s -k -X GET -H 'Accept: application/json' \
    -H "Authorization: Bearer TOKEN_GOES_HERE" \
    https://10.107.8.20:5544/api/v1/vms/0004fb0000060000ccaf98a0baa2c186/ | jq
{
  "friendly_name": "example-vm",
  "name": "0004fb0000060000ccaf98a0baa2c186",
  "uuid": "0004fb00-0006-0000-ccaf-98a0baa2c186",
  "disks": [
    {
      "name": "0004fb000012000005bcf25e906ce843.img",
      "path": "/OVS/Repositories/0004fb00000300006a09d4e1065041cb/VirtualDisks/0004fb000012000005bcf25e906ce843.img",
      "device_name": "xvda",
      "mode": "w"
    }
  ],
  "snapshots": [
    "9633d114-8270-41eb-bffc-67cc342957d9",
    "381ba26b-4a62-4baf-a450-af120ceddcbf"
  ]
}
```

### List snapshots

```
GET /api/v1/vms/{vmID}/snapshots/
```

Example usage:

```bash
curl -s -k -X GET -H 'Accept: application/json' \
    -H "Authorization: Bearer TOKEN_GOES_HERE" \
    https://10.107.8.20:5544/api/v1/vms/0004fb0000060000ccaf98a0baa2c186/snapshots/ | jq
[
  {
    "id": "9633d114-8270-41eb-bffc-67cc342957d9",
    "vmID": "0004fb0000060000ccaf98a0baa2c186",
    "disks": [
      {
        "parent_path": "/OVS/Repositories/0004fb0000030000d42b204197e41e75/VirtualDisks/0004fb0000120000763ac50c4e345d0a.img",
        "path": "/OVS/Repositories/0004fb0000030000d42b204197e41e75/CoriolisSnapshots/9633d114-8270-41eb-bffc-67cc342957d9/0004fb0000120000763ac50c4e345d0a.img",
        "snapshotID": "9633d114-8270-41eb-bffc-67cc342957d9",
        "chunks": [
          {
            "start": 0,
            "length": 643825664,
            "physical_start": 103910735872
          },
          ... truncated ...
          {
            "start": 21464350720,
            "length": 10485760,
            "physical_start": 155250065408
          }
        ],
        "name": "0004fb0000120000763ac50c4e345d0a.img",
        "repo_mountpoint": "/OVS/Repositories/0004fb0000030000d42b204197e41e75"
      }
    ]
  },
  {
    "id": "381ba26b-4a62-4baf-a450-af120ceddcbf",
    "vmID": "0004fb0000060000ccaf98a0baa2c186",
    "disks": [
      {
        "parent_path": "/OVS/Repositories/0004fb0000030000d42b204197e41e75/VirtualDisks/0004fb0000120000763ac50c4e345d0a.img",
        "path": "/OVS/Repositories/0004fb0000030000d42b204197e41e75/CoriolisSnapshots/381ba26b-4a62-4baf-a450-af120ceddcbf/0004fb0000120000763ac50c4e345d0a.img",
        "snapshotID": "381ba26b-4a62-4baf-a450-af120ceddcbf",
        "chunks": [
          {
            "start": 0,
            "length": 643825664,
            "physical_start": 103910735872
          },
          ... truncated ...
          {
            "start": 21464350720,
            "length": 10485760,
            "physical_start": 155250065408
          }
        ],
        "name": "0004fb0000120000763ac50c4e345d0a.img",
        "repo_mountpoint": "/OVS/Repositories/0004fb0000030000d42b204197e41e75"
      }
    ]
  }
]

```

### Get one snapshot

```
GET /api/v1/vms/{vmID}/snapshots/{snapshotID}/
```

Query parameters:

| Name | Type | Optional | Description |
| --- | --- | --- | --- |
| squashChunks | bool | true | If true, continuous chunks will be squashed into a larger chunk, that the client can read in one go.  |
| compareTo | string | true | A snapshot ID previous to this one. If set, the "chunks" field will only hold extents that have changed from the snapshot specified in "compareTo". |

### Create snapshot

```
POST /api/v1/vms/{vmID}/snapshots/
```

Currently, no POST body is required. On success, the API returns snapshot details, including chunks.

### Delete all snapshots of a VM

```
DELETE /api/v1/vms/{vmID}/snapshots/
```

### Delete single snapshot

```
DELETE /api/v1/vms/{vmID}/snapshots/{snapshotID}/
```

### Get disk data

Each snapshot will have associated disks. These disks can be downloaded as a file, or you can choose to download specific ranges of bytes from these disks. Combined with the knowledge we have about written extents exposed by the "chunks" field, we can download the disks as sparse files, or we can do incremental downloads.

```
GET /vms/{vmID}/snapshots/{snapshotID}/disks/{diskID}
```

This handler allows the standard [Range](https://tools.ietf.org/html/rfc7233) header, which can be used to download chunks, selectively.

### Get disk size

```
HEAD /vms/{vmID}/snapshots/{snapshotID}/disks/{diskID}
```

The ```Content-Length``` header will hold the size of the disk.

## Examples

There is an example python client in the [examples](examples) folder of this repository. A CLI app and client [can ge found here](https://github.com/cloudbase/python-ovmexporterclient).
