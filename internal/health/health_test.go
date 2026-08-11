package health

import "testing"

func TestIOFindingsReportsBusyDevice(t *testing.T) {
	before := DeviceStats{Name: "sda", Reads: 10, Writes: 20, ReadBytes: 1024, WriteBytes: 2048, IOTicks: 100}
	after := DeviceStats{Name: "sda", Reads: 110, Writes: 220, ReadBytes: 1024 * 1024, WriteBytes: 2 * 1024 * 1024, IOTicks: 1100}
	findings := ioFindings(before, after, 1)
	if len(findings) != 1 || findings[0].Severity != "warning" || findings[0].Kind != "device-io" {
		t.Fatalf("unexpected findings: %+v", findings)
	}
}

func TestDeviceKeyUsesLinuxMajorMinorEncoding(t *testing.T) {
	device := uint64(8<<8 | 1)
	if got := deviceKey(device); got != 8<<32|1 {
		t.Fatalf("unexpected device key: %d", got)
	}
}
