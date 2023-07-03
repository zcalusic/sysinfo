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

func GetBoardInfo() Board {
	return Board{
		Name:     slurpFile("/sys/class/dmi/id/board_name"),
		Vendor:   slurpFile("/sys/class/dmi/id/board_vendor"),
		Version:  slurpFile("/sys/class/dmi/id/board_version"),
		Serial:   slurpFile("/sys/class/dmi/id/board_serial"),
		AssetTag: slurpFile("/sys/class/dmi/id/board_asset_tag"),
	}
}
