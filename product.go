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

func GetProductInfo() Product {
	return Product{
		Name:    slurpFile("/sys/class/dmi/id/product_name"),
		Vendor:  slurpFile("/sys/class/dmi/id/sys_vendor"),
		Version: slurpFile("/sys/class/dmi/id/product_version"),
		Serial:  slurpFile("/sys/class/dmi/id/product_serial"),
	}
}
