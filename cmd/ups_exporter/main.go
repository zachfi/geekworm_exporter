package main

import (
	"os"

	rpio "github.com/stianeikeland/go-rpio/v4"

	"github.com/xaque208/znet/pkg/util"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"

	"github.com/go-daq/smbus"
)

type PowerState int64

const (
	Powered   PowerState = 0
	Unpowered            = 1
)

func main() {

	logger := util.NewLogger()

	err := rpio.Open()
	if err != nil {
		_ = level.Error(logger).Log("msg", "failed to open gpio", "err", err)
		os.Exit(1)
	}

	_ = level.Info(logger).Log("msg", "battery", "percent", batteryPercent(logger))
	_ = level.Info(logger).Log("msg", "power", "state", powerState(logger))
}

func batteryPercent(logger log.Logger) float64 {
	var max float64 = 255

	c, err := smbus.Open(1, 0x36)
	if err != nil {
		_ = level.Error(logger).Log("msg", "failed to open smbus", "err", err)
		os.Exit(1)
	}
	defer c.Close()

	v, err := c.ReadReg(0x36, 0x1)
	if err != nil {
		_ = level.Error(logger).Log("msg", "failed to read register", "err", err)
		os.Exit(1)
	}

	return (float64(v) / max) * 100
}

func powerState(logger log.Logger) PowerState {
	// Power Loss Detection (PLD) pin 6
	pin := rpio.Pin(6)
	pin.Input()
	s := pin.Read()

	if s == 0 {
		return Powered
	}

	// Make a noise when there is no power.
	// buzzer := rpio.Pin(20)
	// buzzer.Output()
	// buzzer.Write(rpio.High)
	// time.Sleep(time.Millisecond * 500)
	// buzzer.Write(rpio.Low)

	return Unpowered
}
