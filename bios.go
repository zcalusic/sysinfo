// Copyright © 2016 Zlatko Čalušić
//
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file.

package sysinfo

import (
	"fmt"
	"github.com/digitalocean/go-smbios/smbios"
)

// BIOS information.
type BIOS struct {
	Vendor        string `json:"vendor,omitempty"`
	Version       string `json:"version,omitempty"`
	Date          string `json:"date,omitempty"`
	SmbiosVersion string `json:"SmbiosVersion,omitempty"`
}

func (si *SysInfo) getBIOSInfo() {
	si.BIOS.Vendor = slurpFile("/sys/class/dmi/id/bios_vendor")
	si.BIOS.Version = slurpFile("/sys/class/dmi/id/bios_version")
	si.BIOS.Date = slurpFile("/sys/class/dmi/id/bios_date")
	si.BIOS.SmbiosVersion = getSmbiosVersion()
}

func getSmbiosVersion() string {
	var smbiosVersion string
	if rc, ep, err := smbios.Stream(); err == nil {
		defer rc.Close()

		major, minor, rev := ep.Version()
		smbiosVersion = fmt.Sprintf("%d.%d.%d", major, minor, rev)
	}

	return smbiosVersion
}
