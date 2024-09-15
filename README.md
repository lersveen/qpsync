# qpsync – qBittorrent Port Sync

`qpsync` is a utility to dynamically map the listening port in [qBittorrent](https://www.qbittorrent.org/) to the port forwarded by [Gluetun VPN client](https://github.com/qdm12/gluetun).

Rewritten in Golang from [this shell script gist](https://gist.github.com/socketbox/12be539ba0e26b76529e082c97bff53c) by [socketbox](https://github.com/socketbox).

## Usage

```sh
qpsync [options]

Options:
  -f    Path to config (default "config.yaml")
  -j    Run as a job, updating once
  -u    Update frequency in seconds (default 600)
```

## Configuration

`qpsync` can be configured using a YAML config file, see [config-example.yaml](config-example.yaml).

### Environment variables

- `QBITTORRENT_USER`: Username for qBittorrent web UI
- `QBITTORRENT_PASS`: Password for qBittorrent web UI
- `QBITTORRENT_SERVER`: Hostname or IP address for qBittorrent web UI
- `QBITTORRENT_PORT`: Port for qBittorrent web UI
- `GLUETUN_SERVER`: Hostname or IP address for Gluetun VPN container
- `GLUETUN_PORT`: Port for Gluetun VPN container
- `UPDATE_FREQ`: How often to update port
