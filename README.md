# Spectre Stratum Adapter

[![GitHub release](https://img.shields.io/github/v/release/spectre-project/spectre-stratum-bridge.svg)](https://github.com/spectre-project/spectre-stratum-bridge/releases)
[![LICENSE](https://img.shields.io/badge/License-MIT-yellow.svg)](https://github.com/spectre-project/spectre-stratum-bridge/blob/main/LICENSE)
[![GitHub downloads](https://img.shields.io/github/downloads/spectre-project/spectre-stratum-bridge/total.svg)](https://github.com/spectre-project/spectre-stratum-bridge/releases)
[![Join the Spectre Discord Server](https://img.shields.io/discord/1233113243741061240.svg?label=&logo=discord&logoColor=ffffff&color=5865F2)](https://discord.com/invite/FZPYpwszcF)

This is a lightweight daemon that allows mining to a local (or remote)
spectre node using stratum-base miners. It is up to the community to
build a stratum based miner, the original built-in miner is using gRPC
interface.

## Features

Shares-based work allocation with miner-like periodic stat output:

```
===============================================================================
  worker name   |  avg hashrate  |   acc/stl/inv  |    blocks    |    uptime
-------------------------------------------------------------------------------
 ghostface      |      12.62KH/s |          6/0/0 |            6 |   17h36m38s
-------------------------------------------------------------------------------
 Total          |      12.62KH/s |          6/0/0 |            6 |   17h36m46s
-------------------------------------------------------------------------------
 Network Hashrate: 138.40MH/s
======================================================== spr_bridge_v0.3.16 ===
```

## Variable difficulty engine (vardiff)

Multiple miners with significantly different hashrates can be connected
to the same stratum bridge instance, and the appropriate difficulty
will automatically be decided for each one. Default settings target
15 shares/min, resulting in high confidence decisions regarding
difficulty adjustments, and stable measured hashrates (1hr avg
hashrates within +/- 10% of actual). The minimum share difficulty is 12
and optimized for CPUs mining SpectreX.

## Solo mining mode

`solo_mining`: If set to `true`, stratum's variable difficulty (vardiff)
calculations are bypassed, and network difficulty is used instead.
Over several days, the hashrate will essentially reflect your raw luck, as only 
successfully mined blocks impact the hashrate estimates. This makes it a useful
way to gauge luck. `:)`

If you prefer more frequent hashrate updates based on submitted shares, you'll
need to leave `solo_mining` set to `false`.

## Grafana UI

The grafana monitoring UI is an optional component but included for
convenience. It will help to visualize collected statistics.

[detailed instructions here](docs/monitoring-setup.md)

## Prometheus API

If the app is run with the `-prom={port}` flag the application will host
stats on the port specified by `{port}`, these stats are documented in
the file [prom.go](src/spectrestratum/prom.go). This is intended to be use
by prometheus but the stats can be fetched and used independently if
desired. `curl http://localhost:2114/metrics | grep spr_` will get a
listing of current stats. All published stats have a `spr_` prefix for
ease of use.

# Install

## Build from source (native executable)

Install go 1.19 or later using whatever package manager is approprate
for your system, or from [https://go.dev/doc/install](https://go.dev/doc/install).

```
cd cmd/spectrebridge
go build .
```

Modify the config file in `./cmd/spectrebridge/config.yaml` with your setup,
the file comments explain the various flags.

```
./spectrebridge
```

## Docker (all-in-one)

Best option for users who want access to reporting, and aren't already
using Grafana/Prometheus. Requires a local copy of this repository, and
docker installation.

[Install Docker](https://docs.docker.com/engine/install/) using the
appropriate method for your OS. The docker commands below are assuming a
server type installation - details may be different for a desktop
installation.

The following will run the bridge assuming a local spectred node with
default port settings, and listen on port 5555 for incoming stratum
connections.

```
git clone https://github.com/spectre-project/spectre-stratum-bridge.git
cd spectre-stratum-bridge
docker compose -f docker-compose-all-src.yml up -d --build
```

These settings can be updated in the [config.yaml](cmd/spectrebridge/config.yaml)
file, or overridden by modifying, adding or deleting the parameters in the
`command` section of the `docker-compose-all-src.yml` file. Additionally,
Prometheus (the stats database) and Grafana (the dashboard) will be
started and accessible on ports 9090 and 3000 respectively. Once all
services are running, the dashboard should be reachable at
`http://127.0.0.1:3000/d/z73gHk89e1/sprb-monitoring` with default
username and password `admin`.

These commands builds the bridge component from source, rather than
the previous behavior of pulling down a pre-built image. You may still
use the pre-built image by replacing `docker-compose-all-src.yml` with
`docker-compose-all.yml`, but it is not guaranteed to be up to date, so
compiling from source is the better alternative.

## Docker (bridge only)

Best option for users who want docker encapsulation, and don't need
reporting, or are already using Grafana/Prometheus. Requires a local
copy of this repository, and docker installation.

[Install Docker](https://docs.docker.com/engine/install/) using the
appropriate method for your OS. The docker commands below are assuming a
server type installation - details may be different for a desktop
installation.

The following will run the bridge assuming a local spectred node with
default port settings, and listen on port 5555 for incoming stratum
connections.

```
git clone https://github.com/spectre-project/spectre-stratum-bridge.git
cd spectre-stratum-bridge
docker compose -f docker-compose-bridge-src.yml up -d --build
```

These settings can be updated in the [config.yaml](cmd/spectrebridge/config.yaml)
file, or overridden by modifying, adding or deleting the parameters in the
`command` section of the `docker-compose-bridge-src.yml`

These commands builds the bridge component from source, rather than the
previous behavior of pulling down a pre-built image. You may still use
the pre-built image by issuing the command `docker run -p 5555:5555 spectrenetwork/spectre_bridge:latest`,
but it is not guaranteed to be up to date, so compiling from source is
the better alternative.

## Kudos

- https://github.com/KaffinPX/KStratum
- https://github.com/onemorebsmith/kaspa-stratum-bridge
- https://github.com/rdugan/kaspa-stratum-bridge
