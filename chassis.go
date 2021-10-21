// Copyright © 2016 Zlatko Čalušić
//
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file.

package sysinfo

import "strconv"

// Chassis information.
type Chassis struct {
	Type     uint   `json:"type,omitempty"`
	Vendor   string `json:"vendor,omitempty"`
	Version  string `json:"version,omitempty"`
	Serial   string `json:"serial,omitempty"`
	AssetTag string `json:"assettag,omitempty"`
}

func (c *Chassis) GetInfo() {
	if chtype, err := strconv.ParseUint(slurpFile("/sys/class/dmi/id/chassis_type"), 10, 64); err == nil {
		c.Type = uint(chtype)
	}
	c.Vendor = slurpFile("/sys/class/dmi/id/chassis_vendor")
	c.Version = slurpFile("/sys/class/dmi/id/chassis_version")
	c.Serial = slurpFile("/sys/class/dmi/id/chassis_serial")
	c.AssetTag = slurpFile("/sys/class/dmi/id/chassis_asset_tag")
}
