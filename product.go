// Copyright © 2016 Zlatko Čalušić
//
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file.

package sysinfo

// Product information.
type Product struct {
	Name    string `json:"name,omitempty"`
	Vendor  string `json:"vendor,omitempty"`
	Version string `json:"version,omitempty"`
	Serial  string `json:"serial,omitempty"`
}

func (p *Product) GetInfo() {
	p.Name = slurpFile("/sys/class/dmi/id/product_name")
	p.Vendor = slurpFile("/sys/class/dmi/id/sys_vendor")
	p.Version = slurpFile("/sys/class/dmi/id/product_version")
	p.Serial = slurpFile("/sys/class/dmi/id/product_serial")
}
