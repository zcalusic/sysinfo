// Copyright © 2016 Zlatko Čalušić
//
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file.

package sysinfo

import (
	"strings"
	"unsafe"

	"github.com/zcalusic/sysinfo/cpuid"
)

// https://en.wikipedia.org/wiki/CPUID#EAX.3D0:_Get_vendor_ID
var hvmap = map[string]string{
	"bhyve bhyve ": "bhyve",
	"KVMKVMKVM":    "kvm",
	"Microsoft Hv": "hyperv",
	" lrpepyh vr":  "parallels",
	"VMwareVMware": "vmware",
	"XenVMMXenVMM": "xenhvm",
}

func isHypervisorActive() bool {
	var info [4]uint32
	cpuid.CPUID(&info, 0x1)
	return info[2]&(1<<31) != 0
}

func getHypervisorCpuid(ax uint32) string {
	var info [4]uint32
	cpuid.CPUID(&info, ax)
	return hvmap[strings.TrimRight(string((*[12]byte)(unsafe.Pointer(&info[1]))[:]), "\000")]
}

func (n *Node) getHypervisor(bios *BIOS) {
	if !isHypervisorActive() {
		if hypervisorType := slurpFile("/sys/hypervisor/type"); hypervisorType != "" {
			if hypervisorType == "xen" {
				n.Hypervisor = "xenpv"
			}
		}
		return
	}

	// KVM has been caught to move its real signature to this leaf, and put something completely different in the
	// standard location. So this leaf must be checked first.
	if hv := getHypervisorCpuid(0x40000100); hv != "" {
		n.Hypervisor = hv
		return
	}

	if hv := getHypervisorCpuid(0x40000000); hv != "" {
		n.Hypervisor = hv
		return
	}

	// getBIOSInfo() must have run first, to detect BIOS vendor
	if bios.Vendor == "Bochs" {
		n.Hypervisor = "bochs"
		return
	}

	n.Hypervisor = "unknown"
}
