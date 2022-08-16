// Copyright © 2016 Zlatko Čalušić
//
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file.

// Package sysinfo is a Go library providing Linux OS / kernel / hardware system information.
package sysinfo

// SysInfo struct encapsulates all other information structs.
type SysInfo struct {
	Meta    Meta            `json:"sysinfo"`
	Node    Node            `json:"node"`
	OS      OS              `json:"os"`
	Kernel  Kernel          `json:"kernel"`
	Product Product         `json:"product"`
	Board   Board           `json:"board"`
	Chassis Chassis         `json:"chassis"`
	BIOS    BIOS            `json:"bios"`
	CPU     CPU             `json:"cpu"`
	Memory  Memory          `json:"memory"`
	Storage []StorageDevice `json:"storage,omitempty"`
	Network []NetworkDevice `json:"network,omitempty"`
}

// GetSysInfo gathers all available system information.
func GetSysInfo() SysInfo {
	si := SysInfo{}
	// Meta info
	si.Meta = GetMetaInfo()

	// DMI info
	si.Product = GetProductInfo()
	si.Board = GetBoardInfo()
	si.Chassis = GetChassisInfo()
	si.BIOS = GetBIOSInfo()

	// SMBIOS info
	si.Memory, si.CPU.Speed = GetMemoryInfoAndCPUSpeed()

	// Node info
	si.Node = GetNodeInfo(si.BIOS.Vendor)

	// Hardware info

	// we need to detect if we're dealing with a virtualized CPU! Detecting number of
	// physical processors and/or cores is totally unreliable in virtualized environments, so let's not do it.
	virtualEnv := si.Node.Hostname == "" || si.Node.Hypervisor != ""
	si.CPU = GetCPUInfo(virtualEnv)
	si.Storage = GetStorageInfo()
	si.Network = GetNetworkInfo()

	// Software info
	si.OS = GetOSInfo()
	si.Kernel = GetKernelInfo()
	return si
}
