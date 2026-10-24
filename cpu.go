// Copyright © 2016 Zlatko Čalušić
//
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file.

package sysinfo

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"runtime"
	"slices"
	"strconv"
	"strings"
)

// CPU information.
type CPU struct {
	Vendor  string `json:"vendor,omitempty"`
	Model   string `json:"model,omitempty"`
	Speed   uint   `json:"speed,omitempty"`   // CPU clock rate in MHz
	Cache   uint   `json:"cache,omitempty"`   // CPU cache size in KB
	Cpus    uint   `json:"cpus,omitempty"`    // number of physical CPUs
	Cores   uint   `json:"cores,omitempty"`   // number of physical CPU cores
	Threads uint   `json:"threads,omitempty"` // number of logical (HT) CPU cores
}

var (
	reTwoColumns = regexp.MustCompile("\t+: ")
	reExtraSpace = regexp.MustCompile(" +")
	reCacheSize  = regexp.MustCompile(`^(\d+) KB$`)
)

// ARM vendors and cores named by the "CPU implementer" and "CPU part" lines,
// from util-linux sys-utils/lscpu-arm.c.
var (
	armImplementers = map[string]string{
		"0x41": "ARM",
		"0x42": "Broadcom",
		"0x43": "Cavium",
		"0x48": "HiSilicon",
		"0x4e": "NVIDIA",
		"0x51": "Qualcomm",
		"0x53": "Samsung",
		"0x61": "Apple",
		"0xc0": "Ampere",
	}
	armParts = map[string]string{
		"0xc05": "Cortex-A5",
		"0xc07": "Cortex-A7",
		"0xc08": "Cortex-A8",
		"0xc09": "Cortex-A9",
		"0xc0d": "Cortex-A17",
		"0xc0e": "Cortex-A17",
		"0xc0f": "Cortex-A15",
		"0xd01": "Cortex-A32",
		"0xd02": "Cortex-A34",
		"0xd03": "Cortex-A53",
		"0xd04": "Cortex-A35",
		"0xd05": "Cortex-A55",
		"0xd06": "Cortex-A65",
		"0xd07": "Cortex-A57",
		"0xd08": "Cortex-A72",
		"0xd09": "Cortex-A73",
		"0xd0a": "Cortex-A75",
		"0xd0b": "Cortex-A76",
		"0xd0c": "Neoverse-N1",
		"0xd0d": "Cortex-A77",
		"0xd0e": "Cortex-A76AE",
		"0xd40": "Neoverse-V1",
		"0xd41": "Cortex-A78",
		"0xd42": "Cortex-A78AE",
		"0xd43": "Cortex-A65AE",
		"0xd44": "Cortex-X1",
		"0xd46": "Cortex-A510",
		"0xd47": "Cortex-A710",
		"0xd48": "Cortex-X2",
		"0xd49": "Neoverse-N2",
		"0xd4a": "Neoverse-E1",
		"0xd4b": "Cortex-A78C",
		"0xd4c": "Cortex-X1C",
		"0xd4d": "Cortex-A715",
		"0xd4e": "Cortex-X3",
		"0xd4f": "Neoverse-V2",
		"0xd80": "Cortex-A520",
		"0xd81": "Cortex-A720",
		"0xd82": "Cortex-X4",
		"0xd83": "Neoverse-V3AE",
		"0xd84": "Neoverse-V3",
		"0xd85": "Cortex-X925",
		"0xd87": "Cortex-A725",
		"0xd88": "Cortex-A520AE",
		"0xd89": "Cortex-A720AE",
		"0xd8e": "Neoverse-N3",
		"0xd8f": "Cortex-A320",
	}
)

func (si *SysInfo) getCPUInfo() {
	si.CPU.Threads = uint(runtime.NumCPU())

	f, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()

	si.parseCPUInfo(f)
}

func (si *SysInfo) parseCPUInfo(r io.Reader) {
	cpu := make(map[string]bool)
	core := make(map[string]bool)

	var cpuID, implementer string
	var threads uint
	var parts []string // each core type once, in the order listed

	s := bufio.NewScanner(r)
	for s.Scan() {
		if sl := reTwoColumns.Split(s.Text(), 2); sl != nil {
			switch sl[0] {
			case "processor":
				threads++
			case "physical id":
				cpuID = sl[1]
				cpu[cpuID] = true
			case "core id":
				coreID := fmt.Sprintf("%s/%s", cpuID, sl[1])
				core[coreID] = true
			case "vendor_id":
				if si.CPU.Vendor == "" {
					si.CPU.Vendor = sl[1]
				}
			case "model name":
				if si.CPU.Model == "" {
					// CPU model, as reported by /proc/cpuinfo, can be a bit ugly. Clean up...
					model := reExtraSpace.ReplaceAllLiteralString(sl[1], " ")
					si.CPU.Model = strings.Replace(model, "- ", "-", 1)
				}
			case "cache size":
				if si.CPU.Cache == 0 {
					if m := reCacheSize.FindStringSubmatch(sl[1]); m != nil {
						if cache, err := strconv.ParseUint(m[1], 10, 64); err == nil {
							si.CPU.Cache = uint(cache)
						}
					}
				}
			case "CPU implementer":
				if implementer == "" {
					implementer = sl[1]
				}
			case "CPU part":
				if name := armParts[sl[1]]; name != "" && implementer == "0x41" && !slices.Contains(parts, name) {
					parts = append(parts, name)
				}
			}
		}
	}
	if s.Err() != nil {
		return
	}

	// ARM kernels print neither vendor_id nor, on arm64, model name. A
	// big.LITTLE chip lists each of its core types.
	if si.CPU.Vendor == "" {
		si.CPU.Vendor = armImplementers[implementer]
	}
	if si.CPU.Model == "" {
		si.CPU.Model = strings.Join(parts, "/")
	}

	// runtime.NumCPU() counts only the CPUs in the process affinity mask, /proc/cpuinfo lists all online CPUs.
	if threads > 0 {
		si.CPU.Threads = threads
	}
	si.CPU.Cpus = uint(len(cpu))
	si.CPU.Cores = uint(len(core))
}
