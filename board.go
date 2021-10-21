// Copyright © 2016 Zlatko Čalušić
//
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file.

package sysinfo

// Board information.
type Board struct {
	Name     string `json:"name,omitempty"`
	Vendor   string `json:"vendor,omitempty"`
	Version  string `json:"version,omitempty"`
	Serial   string `json:"serial,omitempty"`
	AssetTag string `json:"assettag,omitempty"`
}

func (b *Board) GetInfo() {
	b.Name = slurpFile("/sys/class/dmi/id/board_name")
	b.Vendor = slurpFile("/sys/class/dmi/id/board_vendor")
	b.Version = slurpFile("/sys/class/dmi/id/board_version")
	b.Serial = slurpFile("/sys/class/dmi/id/board_serial")
	b.AssetTag = slurpFile("/sys/class/dmi/id/board_asset_tag")
}
