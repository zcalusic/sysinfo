// Copyright © 2016 Zlatko Čalušić
//
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file.

package sysinfo

import (
	"syscall"
	"unsafe"
)

// getSupported returns the link modes the network device supports, from the ethtool ioctl.
func getSupported(name string) uint32 {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_IP)
	if err != nil {
		return 0
	}
	defer func() { _ = syscall.Close(fd) }()

	// struct ethtool_cmd from /usr/include/linux/ethtool.h
	var ethtool struct {
		Cmd           uint32
		Supported     uint32
		Advertising   uint32
		Speed         uint16
		Duplex        uint8
		Port          uint8
		PhyAddress    uint8
		Transceiver   uint8
		Autoneg       uint8
		MdioSupport   uint8
		Maxtxpkt      uint32
		Maxrxpkt      uint32
		SpeedHi       uint16
		EthTpMdix     uint8
		Reserved2     uint8
		LpAdvertising uint32
		Reserved      [2]uint32
	}

	// ETHTOOL_GSET from /usr/include/linux/ethtool.h
	const GSET = 0x1

	ethtool.Cmd = GSET

	// struct ifreq from /usr/include/linux/if.h, padded to its full size (the union is as big as struct ifmap),
	// because the kernel copies the whole struct back
	var ifr struct {
		Name [16]byte
		Data uintptr
		_    [unsafe.Sizeof(uintptr(0)) + 8]byte
	}

	copy(ifr.Name[:], name+"\000")
	ifr.Data = uintptr(unsafe.Pointer(&ethtool)) //nolint:gosec // ioctl argument, the kernel fills in struct ethtool_cmd

	// SIOCETHTOOL from /usr/include/linux/sockios.h
	const SIOCETHTOOL = 0x8946

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(SIOCETHTOOL), uintptr(unsafe.Pointer(&ifr))) //nolint:gosec // ioctl argument, struct ifreq is padded to its full kernel size
	if errno == 0 {
		return ethtool.Supported
	}

	return 0
}
