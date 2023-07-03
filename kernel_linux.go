// Copyright © 2016 Zlatko Čalušić
//
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file.

//go:build linux
// +build linux

package sysinfo

import (
	"strings"
	"syscall"
	"unsafe"
)

func GetKernelInfo() Kernel {
	kernelInfo := Kernel{}
	kernelInfo.Release = slurpFile("/proc/sys/kernel/osrelease")
	kernelInfo.Version = slurpFile("/proc/sys/kernel/version")

	var uname syscall.Utsname
	if err := syscall.Uname(&uname); err != nil {
		return kernelInfo
	}

	kernelInfo.Architecture = strings.TrimRight(string((*[65]byte)(unsafe.Pointer(&uname.Machine))[:]), "\000")
	return kernelInfo
}
