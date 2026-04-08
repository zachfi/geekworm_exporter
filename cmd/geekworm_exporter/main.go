package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-daq/smbus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	rpio "github.com/stianeikeland/go-rpio/v4"
)

// Geekworm X728 UPS hardware constants.
const (
	smbusChannel = 1    // I2C bus number (bus 1 on Raspberry Pi)
	smbusAddr    = 0x36 // MAX17040 fuel gauge I2C address

	// MAX17040 register addresses (single-byte reads, big-endian pairs).
	regVCellHi = 0x02 // VCELL high byte
	regVCellLo = 0x03 // VCELL low byte
	regSOCHi   = 0x04 // State of Charge high byte (integer %)
	regSOCLo   = 0x05 // State of Charge low byte (fractional %, units of 1/256)

	pldPin = 6 // GPIO pin for Power Loss Detection (active high = power lost)
)

var (
	batteryCapacity = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "geekworm_battery_capacity_percent",
		Help: "Battery state of charge as a percentage (0–100).",
	})
	batteryVoltage = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "geekworm_battery_voltage_volts",
		Help: "Battery voltage in volts.",
	})
	powerState = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "geekworm_power_state",
		Help: "Mains power state: 1 = AC powered, 0 = running on battery.",
	})
	lastCollectTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "geekworm_last_collect_timestamp_seconds",
		Help: "Unix timestamp of the last successful sensor read.",
	})
)

func main() {
	listenAddr := flag.String("web.listen-address", ":9990", "Address on which to expose metrics")
	interval := flag.Duration("interval", 15*time.Second, "Interval between sensor reads")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	prometheus.MustRegister(batteryCapacity, batteryVoltage, powerState, lastCollectTime)

	if err := rpio.Open(); err != nil {
		logger.Error("failed to open GPIO", "err", err)
		os.Exit(1)
	}
	defer rpio.Close()

	go func() {
		for {
			collect(logger)
			time.Sleep(*interval)
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html><body><p><a href="/metrics">Metrics</a></p></body></html>`))
	})

	logger.Info("starting metrics server", "addr", *listenAddr)
	if err := http.ListenAndServe(*listenAddr, nil); err != nil {
		logger.Error("server failed", "err", err)
		os.Exit(1)
	}
}

func collect(logger *slog.Logger) {
	if pct, err := readCapacity(); err != nil {
		logger.Warn("failed to read battery capacity", "err", err)
	} else {
		batteryCapacity.Set(pct)
	}

	if v, err := readVoltage(); err != nil {
		logger.Warn("failed to read battery voltage", "err", err)
	} else {
		batteryVoltage.Set(v)
	}

	powerState.Set(float64(readPowerState()))
	lastCollectTime.SetToCurrentTime()
}

// readCapacity returns the battery state of charge as a percentage (0.0–100.0).
// Reads the MAX17040 SOC register pair (0x04/0x05) via SMBus.
func readCapacity() (float64, error) {
	c, err := smbus.Open(smbusChannel, smbusAddr)
	if err != nil {
		return 0, err
	}
	defer c.Close()

	hi, err := c.ReadReg(smbusAddr, regSOCHi)
	if err != nil {
		return 0, err
	}

	lo, err := c.ReadReg(smbusAddr, regSOCLo)
	if err != nil {
		return 0, err
	}

	return calcCapacity(hi, lo), nil
}

// readVoltage returns the battery cell voltage in volts.
// Reads the MAX17040 VCELL register pair (0x02/0x03) via SMBus.
func readVoltage() (float64, error) {
	c, err := smbus.Open(smbusChannel, smbusAddr)
	if err != nil {
		return 0, err
	}
	defer c.Close()

	hi, err := c.ReadReg(smbusAddr, regVCellHi)
	if err != nil {
		return 0, err
	}

	lo, err := c.ReadReg(smbusAddr, regVCellLo)
	if err != nil {
		return 0, err
	}

	return calcVoltage(hi, lo), nil
}

// calcVoltage converts two MAX17040 VCELL register bytes to volts.
// The 12-bit ADC value is stored in the upper 12 bits of the 16-bit pair; LSB = 1.25mV.
func calcVoltage(hi, lo byte) float64 {
	raw := (uint16(hi) << 8) | uint16(lo)
	return float64(raw>>4) * 0.00125
}

// calcCapacity converts two MAX17040 SOC register bytes to a percentage.
// The high byte is the integer portion; the low byte is the fractional portion in units of 1/256.
func calcCapacity(hi, lo byte) float64 {
	return float64(hi) + float64(lo)/256.0
}

// readPowerState returns 1 if AC power is present, 0 if running on battery.
// GPIO pin 6 (PLD) reads high when power is lost.
func readPowerState() int {
	pin := rpio.Pin(pldPin)
	pin.Input()
	if pin.Read() == 0 {
		return 1
	}
	return 0
}
