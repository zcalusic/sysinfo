// Copyright © 2016 Zlatko Čalušić
//
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file.

package sysinfo

import (
	"os"
	"path"
	"strings"
)

// NetworkDevice information.
type NetworkDevice struct {
	Name       string `json:"name,omitempty"`
	Driver     string `json:"driver,omitempty"`
	MACAddress string `json:"macaddress,omitempty"`
	Port       string `json:"port,omitempty"`
	Speed      uint   `json:"speed,omitempty"` // device max supported speed in Mbps
}

func getPortType(supp uint32) (port string) {
	for i, p := range [...]string{"tp", "aui", "mii", "fibre", "bnc"} {
		if supp&(1<<uint(i+7)) > 0 {
			port += p + "/"
		}
	}

	port = strings.TrimRight(port, "/")
	return
}

func getMaxSpeed(supp uint32) (speed uint) {
	// Fancy, right?
	switch {
	case supp&0x78000000 > 0:
		speed = 56000
	case supp&0x07800000 > 0:
		speed = 40000
	case supp&0x00600000 > 0:
		speed = 20000
	case supp&0x001c1000 > 0:
		speed = 10000
	case supp&0x00008000 > 0:
		speed = 2500
	case supp&0x00020030 > 0:
		speed = 1000
	case supp&0x0000000c > 0:
		speed = 100
	case supp&0x00000003 > 0:
		speed = 10
	}

	return
}

func (si *SysInfo) getNetworkInfo() {
	sysClassNet := "/sys/class/net"
	devices, err := os.ReadDir(sysClassNet)
	if err != nil {
		return
	}

	si.Network = make([]NetworkDevice, 0)
	for _, link := range devices {
		fullpath := path.Join(sysClassNet, link.Name())
		dev, err := os.Readlink(fullpath)
		if err != nil {
			continue
		}

		if strings.HasPrefix(dev, "../../devices/virtual/") {
			continue
		}

		supp := getSupported(link.Name())

		device := NetworkDevice{
			Name:       link.Name(),
			MACAddress: slurpFile(path.Join(fullpath, "address")),
			Port:       getPortType(supp),
			Speed:      getMaxSpeed(supp),
		}

		if driver, err := os.Readlink(path.Join(fullpath, "device", "driver")); err == nil {
			device.Driver = path.Base(driver)
		}

		si.Network = append(si.Network, device)
	}
}
