// Copyright © 2016 Zlatko Čalušić
//
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file.

// Package sysinfo is a Go library providing Linux OS / kernel / hardware system information.
package sysinfo

// SysInfo struct encapsulates all other information structs.
type SysInfo struct {
	Meta    Meta    `json:"sysinfo"`
	Node    Node    `json:"node"`
	OS      OS      `json:"os"`
	Kernel  Kernel  `json:"kernel"`
	Product Product `json:"product"`
	Board   Board   `json:"board"`
	Chassis Chassis `json:"chassis"`
	BIOS    BIOS    `json:"bios"`
	CPU     CPU     `json:"cpu"`
	Memory  Memory  `json:"memory"`
	Storage Storage `json:"storage,omitempty"`
	Network Network `json:"network,omitempty"`
}

// GetSysInfo gathers all available system information.
func (si *SysInfo) GetSysInfo() {
	// Meta info
	si.Meta.GetInfo()

	// DMI info
	si.Product.GetInfo()
	si.Board.GetInfo()
	si.Chassis.GetInfo()
	si.BIOS.GetInfo()

	// SMBIOS info
	si.Memory.GetInfo(&si.CPU)

	// Node info
	si.Node.GetInfo(&si.BIOS)

	// Hardware info
	si.CPU.GetInfo(si.Node.Hostname != "" && si.Node.Hypervisor == "")
	si.Storage.GetInfo()
	si.Network.GetInfo()

	// Software info
	si.OS.GetInfo()
	si.Kernel.GetInfo()
}
