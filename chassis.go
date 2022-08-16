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

func GetChassisInfo() Chassis {
	chassisType, _ := strconv.ParseUint(slurpFile("/sys/class/dmi/id/chassis_type"), 10, 64)
	return Chassis{
		Type:     uint(chassisType),
		Vendor:   slurpFile("/sys/class/dmi/id/chassis_vendor"),
		Version:  slurpFile("/sys/class/dmi/id/chassis_version"),
		Serial:   slurpFile("/sys/class/dmi/id/chassis_serial"),
		AssetTag: slurpFile("/sys/class/dmi/id/chassis_asset_tag"),
	}
}
