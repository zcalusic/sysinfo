// Copyright © 2016 Zlatko Čalušić
//
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file.

package sysinfo

// BIOS information.
type BIOS struct {
	Vendor  string `json:"vendor,omitempty"`
	Version string `json:"version,omitempty"`
	Date    string `json:"date,omitempty"`
}

func GetBIOSInfo() BIOS {
	return BIOS{
		Vendor:  slurpFile("/sys/class/dmi/id/bios_vendor"),
		Version: slurpFile("/sys/class/dmi/id/bios_version"),
		Date:    slurpFile("/sys/class/dmi/id/bios_date"),
	}
}
