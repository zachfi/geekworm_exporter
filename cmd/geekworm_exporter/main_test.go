package main

import (
	"math"
	"testing"
)

func TestCalcVoltage(t *testing.T) {
	tests := []struct {
		name string
		hi   byte
		lo   byte
		want float64
	}{
		// MAX17040 VCELL: 12-bit value in upper 12 bits, LSB = 1.25mV.
		// To derive hi/lo for a target voltage V:
		//   raw12 = V / 0.00125
		//   raw16 = raw12 << 4
		//   hi = raw16 >> 8, lo = raw16 & 0xFF
		{"zero volts", 0x00, 0x00, 0.0},
		{"3.6V (typical mid charge)", 0xB4, 0x00, 3.6},
		{"4.2V (fully charged)", 0xD2, 0x00, 4.2},
		{"3.0V (nearly empty)", 0x96, 0x00, 3.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calcVoltage(tt.hi, tt.lo)
			if math.Abs(got-tt.want) > 0.01 {
				t.Errorf("calcVoltage(%#02x, %#02x) = %.4fV, want %.4fV", tt.hi, tt.lo, got, tt.want)
			}
		})
	}
}

func TestCalcCapacity(t *testing.T) {
	tests := []struct {
		name string
		hi   byte
		lo   byte
		want float64
	}{
		{"empty", 0, 0, 0.0},
		{"half", 50, 0, 50.0},
		{"full", 100, 0, 100.0},
		{"fractional 75.5%", 75, 128, 75.5},
		{"fractional 99.9%", 99, 230, 99.0 + 230.0/256.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calcCapacity(tt.hi, tt.lo)
			if math.Abs(got-tt.want) > 0.01 {
				t.Errorf("calcCapacity(%d, %d) = %.4f%%, want %.4f%%", tt.hi, tt.lo, got, tt.want)
			}
		})
	}
}
