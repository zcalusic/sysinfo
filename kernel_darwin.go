// Copyright © 2016 Zlatko Čalušić
//
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file.

//go:build darwin

package sysinfo

// Kernel information.
type Kernel struct {
	Release      string `json:"release,omitempty"`
	Version      string `json:"version,omitempty"`
	Architecture string `json:"architecture,omitempty"`
}

func (si *SysInfo) getKernelInfo() {
}
