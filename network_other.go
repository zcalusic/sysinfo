// Copyright © 2016 Zlatko Čalušić
//
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file.

//go:build !linux

package sysinfo

// getSupported returns no supported link modes: the ethtool ioctl is Linux-only.
func getSupported(string) uint32 {
	return 0
}
