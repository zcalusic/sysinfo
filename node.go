// Copyright © 2016 Zlatko Čalušić
//
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file.

package sysinfo

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"os"
	"strings"
	"time"
)

// Node information.
type Node struct {
	Hostname   string `json:"hostname,omitempty"`
	MachineID  string `json:"machineid,omitempty"`
	Hypervisor string `json:"hypervisor,omitempty"`
	Timezone   string `json:"timezone,omitempty"`
}

func GetHostname() string {
	return slurpFile("/proc/sys/kernel/hostname")
}

func GetSetMachineID() string {
	const pathSystemdMachineID = "/etc/machine-id"
	const pathDbusMachineID = "/var/lib/dbus/machine-id"

	systemdMachineID := slurpFile(pathSystemdMachineID)
	dbusMachineID := slurpFile(pathDbusMachineID)

	if systemdMachineID != "" && dbusMachineID != "" {
		// All OK, just return the machine id.
		if systemdMachineID == dbusMachineID {
			return systemdMachineID
		}

		// They both exist, but they don't match! Copy systemd machine id to DBUS machine id.
		spewFile(pathDbusMachineID, systemdMachineID, 0444)
		return systemdMachineID
	}

	// Copy DBUS machine id to non-existent systemd machine id.
	if systemdMachineID == "" && dbusMachineID != "" {
		spewFile(pathSystemdMachineID, dbusMachineID, 0444)
		return dbusMachineID
	}

	// Copy systemd machine id to non-existent DBUS machine id.
	if systemdMachineID != "" && dbusMachineID == "" {
		spewFile(pathDbusMachineID, systemdMachineID, 0444)
		return systemdMachineID
	}

	// Generate and write fresh new machine ID to both locations, conforming to the DBUS specification:
	// https://dbus.freedesktop.org/doc/dbus-specification.html#uuids

	random := make([]byte, 12)
	if _, err := rand.Read(random); err != nil {
		return ""
	}
	newMachineID := fmt.Sprintf("%x%x", random, time.Now().Unix())

	spewFile(pathSystemdMachineID, newMachineID, 0444)
	spewFile(pathDbusMachineID, newMachineID, 0444)
	return newMachineID
}

func GetTimezone() string {
	const zoneInfoPrefix = "/usr/share/zoneinfo/"

	if fi, err := os.Lstat("/etc/localtime"); err == nil {
		if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
			if tzfile, err := os.Readlink("/etc/localtime"); err == nil {
				tzfile = strings.TrimPrefix(tzfile, "..")
				if strings.HasPrefix(tzfile, zoneInfoPrefix) {
					return strings.TrimPrefix(tzfile, zoneInfoPrefix)
				}
			}
		}
	}

	if timezone := slurpFile("/etc/timezone"); timezone != "" {
		return timezone
	}

	if f, err := os.Open("/etc/sysconfig/clock"); err == nil {
		defer f.Close()
		s := bufio.NewScanner(f)
		for s.Scan() {
			if sl := strings.Split(s.Text(), "="); len(sl) == 2 {
				if sl[0] == "ZONE" {
					return strings.Trim(sl[1], `"`)
				}
			}
		}
	}
	return ""
}

func GetNodeInfo(biosVendor ...string) Node {
	return Node{
		Hostname:   GetHostname(),
		MachineID:  GetSetMachineID(),
		Hypervisor: GetHypervisor(biosVendor...),
		Timezone:   GetTimezone(),
	}
}
