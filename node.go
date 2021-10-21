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

func (n *Node) getHostname() {
	n.Hostname = slurpFile("/proc/sys/kernel/hostname")
}

func (n *Node) getSetMachineID() {
	const pathSystemdMachineID = "/etc/machine-id"
	const pathDbusMachineID = "/var/lib/dbus/machine-id"

	systemdMachineID := slurpFile(pathSystemdMachineID)
	dbusMachineID := slurpFile(pathDbusMachineID)

	if systemdMachineID != "" && dbusMachineID != "" {
		// All OK, just return the machine id.
		if systemdMachineID == dbusMachineID {
			n.MachineID = systemdMachineID
			return
		}

		// They both exist, but they don't match! Copy systemd machine id to DBUS machine id.
		spewFile(pathDbusMachineID, systemdMachineID, 0444)
		n.MachineID = systemdMachineID
		return
	}

	// Copy DBUS machine id to non-existent systemd machine id.
	if systemdMachineID == "" && dbusMachineID != "" {
		spewFile(pathSystemdMachineID, dbusMachineID, 0444)
		n.MachineID = dbusMachineID
		return
	}

	// Copy systemd machine id to non-existent DBUS machine id.
	if systemdMachineID != "" && dbusMachineID == "" {
		spewFile(pathDbusMachineID, systemdMachineID, 0444)
		n.MachineID = systemdMachineID
		return
	}

	// Generate and write fresh new machine ID to both locations, conforming to the DBUS specification:
	// https://dbus.freedesktop.org/doc/dbus-specification.html#uuids

	random := make([]byte, 12)
	if _, err := rand.Read(random); err != nil {
		return
	}
	newMachineID := fmt.Sprintf("%x%x", random, time.Now().Unix())

	spewFile(pathSystemdMachineID, newMachineID, 0444)
	spewFile(pathDbusMachineID, newMachineID, 0444)
	n.MachineID = newMachineID
}

func (n *Node) getTimezone() {
	if fi, err := os.Lstat("/etc/localtime"); err == nil {
		if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
			if tzfile, err := os.Readlink("/etc/localtime"); err == nil {
				if strings.HasPrefix(tzfile, "/usr/share/zoneinfo/") {
					n.Timezone = strings.TrimPrefix(tzfile, "/usr/share/zoneinfo/")
					return
				}
			}
		}
	}

	if timezone := slurpFile("/etc/timezone"); timezone != "" {
		n.Timezone = timezone
		return
	}

	if f, err := os.Open("/etc/sysconfig/clock"); err == nil {
		defer f.Close()
		s := bufio.NewScanner(f)
		for s.Scan() {
			if sl := strings.Split(s.Text(), "="); len(sl) == 2 {
				if sl[0] == "ZONE" {
					n.Timezone = strings.Trim(sl[1], `"`)
					return
				}
			}
		}
	}
}

func (n *Node) GetInfo(bios *BIOS) {
	n.getHostname()
	n.getSetMachineID()
	n.getHypervisor(bios)
	n.getTimezone()
}
