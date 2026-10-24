// Copyright © 2016 Zlatko Čalušić
//
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file.

package sysinfo

import (
	"bufio"
	"os"
	"strings"
)

// Node information.
type Node struct {
	Hostname   string `json:"hostname,omitempty"`
	MachineID  string `json:"machineid,omitempty"`
	Hypervisor string `json:"hypervisor,omitempty"`
	Timezone   string `json:"timezone,omitempty"`
}

func (si *SysInfo) getHostname() {
	si.Node.Hostname = slurpFile("/proc/sys/kernel/hostname")
}

func (si *SysInfo) getMachineID() {
	// Read only, never write: an empty /etc/machine-id may be intentional (image templates, systemd first boot),
	// and during first boot it may contain "uninitialized".
	for _, path := range []string{"/etc/machine-id", "/var/lib/dbus/machine-id"} {
		if machineID := slurpFile(path); machineID != "" && machineID != "uninitialized" {
			si.Node.MachineID = machineID
			return
		}
	}
}

func (si *SysInfo) getTimezone() {
	const zoneInfoPrefix = "/usr/share/zoneinfo/"

	if fi, err := os.Lstat("/etc/localtime"); err == nil {
		if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
			if tzfile, err := os.Readlink("/etc/localtime"); err == nil {
				tzfile = strings.TrimPrefix(tzfile, "..")
				if timezone, ok := strings.CutPrefix(tzfile, zoneInfoPrefix); ok {
					si.Node.Timezone = timezone
					return
				}
			}
		}
	}

	if timezone := slurpFile("/etc/timezone"); timezone != "" {
		si.Node.Timezone = timezone
		return
	}

	if f, err := os.Open("/etc/sysconfig/clock"); err == nil {
		defer func() { _ = f.Close() }()
		s := bufio.NewScanner(f)
		for s.Scan() {
			if sl := strings.Split(s.Text(), "="); len(sl) == 2 {
				if sl[0] == "ZONE" {
					si.Node.Timezone = strings.Trim(sl[1], `"`)
					return
				}
			}
		}
		if s.Err() != nil {
			return
		}
	}
}

func (si *SysInfo) getNodeInfo() {
	si.getHostname()
	si.getMachineID()
	si.getHypervisor()
	si.getTimezone()
}
