// Copyright © 2016 Zlatko Čalušić
//
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file.

package sysinfo

import (
	"fmt"
	"strings"
	"testing"
)

// From an Orange Pi Prime (Allwinner H5).
const cpuinfoCortexA53 = `processor	: 0
BogoMIPS	: 48.00
Features	: fp asimd evtstrm aes pmull sha1 sha2 crc32 cpuid
CPU implementer	: 0x41
CPU architecture: 8
CPU variant	: 0x0
CPU part	: 0xd03
CPU revision	: 4

`

// bigLittle is laid out like an RK3588: four Cortex-A55, then four Cortex-A76.
func bigLittle() string {
	var b strings.Builder
	for i := range 8 {
		part := "0xd05"
		if i >= 4 {
			part = "0xd0b"
		}
		fmt.Fprintf(&b, "processor\t: %d\nCPU implementer\t: 0x41\nCPU architecture: 8\nCPU part\t: %s\n\n", i, part)
	}
	return b.String()
}

func TestParseCPUInfoARM(t *testing.T) {
	for _, tc := range []struct {
		name, cpuinfo, model string
		threads              uint
	}{
		{"cortex-a53", cpuinfoCortexA53, "Cortex-A53", 1},
		{"big.LITTLE", bigLittle(), "Cortex-A55/Cortex-A76", 8},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var si SysInfo
			si.parseCPUInfo(strings.NewReader(tc.cpuinfo))
			if si.CPU.Vendor != "ARM" || si.CPU.Model != tc.model || si.CPU.Threads != tc.threads {
				t.Fatalf("got vendor %q model %q threads %d, want ARM %q %d", si.CPU.Vendor, si.CPU.Model, si.CPU.Threads, tc.model, tc.threads)
			}
		})
	}
}
