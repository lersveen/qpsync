# GoFwd

`gofwd` synchronizes the listening port in qBittorrent to the active port used by Gluetun. This is useful in cases

## Usage

```sh
gofwd [options]

Options:
  -f    Path to config (default "config.yaml")
  -j    Run as a job, updating once
  -u    Update frequency in seconds (default 600)
```

## Configuration

GoFwd can be configured using a YAML config file, see [config-example.yaml](config-example.yaml).

### Environment variables

- `QBITTORRENT_USER`: Username for qBittorrent web UI
- `QBITTORRENT_PASS`: Password for qBittorrent web UI
- `QBITTORRENT_SERVER`: Hostname or IP address for qBittorrent web UI
- `QBITTORRENT_PORT`: Port for qBittorrent web UI
- `GLUETUN_SERVER`: Hostname or IP address for Gluetun VPN container
- `GLUETUN_PORT`: Port for Gluetun VPN container
- `UPDATE_FREQ`: How often to update port
