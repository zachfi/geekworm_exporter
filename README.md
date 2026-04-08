# geekworm_exporter

A [Prometheus](https://prometheus.io/) exporter for the [Geekworm X728](https://wiki.geekworm.com/X728) UPS HAT for Raspberry Pi.

Exposes battery capacity, cell voltage, and AC power state as Prometheus metrics via an HTTP endpoint.

## Hardware

The X728 uses a MAX17040 fuel gauge IC at I2C address `0x36` and a GPIO Power Loss Detection (PLD) pin.

| Sensor | Interface | Details |
|--------|-----------|---------|
| Battery capacity | I2C (SMBus) | MAX17040 SOC register (0x04/0x05) |
| Battery voltage | I2C (SMBus) | MAX17040 VCELL register (0x02/0x03) |
| AC power state | GPIO pin 6 | High = power lost |

## Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `geekworm_battery_capacity_percent` | Gauge | State of charge (0–100%) |
| `geekworm_battery_voltage_volts` | Gauge | Cell voltage in volts |
| `geekworm_power_state` | Gauge | `1` = AC powered, `0` = on battery |
| `geekworm_last_collect_timestamp_seconds` | Gauge | Unix timestamp of last successful read |

## Usage

```
geekworm_exporter [flags]

Flags:
  -web.listen-address string   Address on which to expose metrics (default ":9990")
  -interval duration           Interval between sensor reads (default 15s)
```

Metrics are available at `http://<host>:9990/metrics`.

## Building

Requires Go 1.22+ and access to `/dev/i2c-1` and `/dev/gpiomem` (or run as root).

```sh
make build           # build for current platform
make compile-only    # cross-compile for ARMv7 (Raspberry Pi)
make test            # run unit tests
make lint            # go vet
```

## Installation

Build for ARM and copy the binary to your Raspberry Pi:

```sh
GOARCH=arm GOARM=7 GOOS=linux CGO_ENABLED=0 go build -o geekworm_exporter ./cmd/geekworm_exporter/
scp geekworm_exporter pi@<host>:/usr/local/bin/
```

### systemd service

```ini
[Unit]
Description=Geekworm X728 UPS Prometheus exporter
After=network.target

[Service]
ExecStart=/usr/local/bin/geekworm_exporter -web.listen-address=:9990
Restart=always
RestartSec=5s
User=root

[Install]
WantedBy=multi-user.target
```

```sh
sudo systemctl enable --now geekworm_exporter
```

## Prometheus scrape config

```yaml
scrape_configs:
  - job_name: geekworm
    static_configs:
      - targets: ['<host>:9990']
```
