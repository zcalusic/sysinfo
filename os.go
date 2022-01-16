// Copyright © 2016 Zlatko Čalušić
//
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file.

package sysinfo

import (
	"bufio"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

// OS information.
type OS struct {
	Name         string `json:"name,omitempty"`
	Vendor       string `json:"vendor,omitempty"`
	Version      string `json:"version,omitempty"`
	Release      string `json:"release,omitempty"`
	Architecture string `json:"architecture,omitempty"`
}

var (
	rePrettyName = regexp.MustCompile(`^PRETTY_NAME=(.*)$`)
	reID         = regexp.MustCompile(`^ID=(.*)$`)
	reVersionID  = regexp.MustCompile(`^VERSION_ID=(.*)$`)
	reUbuntu     = regexp.MustCompile(`[\( ]([\d\.]+)`)
	reCentOS     = regexp.MustCompile(`^CentOS( Linux)? release ([\d\.]+) `)
	reRedHat     = regexp.MustCompile(`[\( ]([\d\.]+)`)
)

func file_is_exists(f string) bool {
	_, err := os.Stat(f)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

func getCentosOrRedHatOldSystem(si *SysInfo) {
	// centos 5/6 or redhat enterprise 5/6 need to use /etc/redhat-release to detect system info
	fileName := "/etc/redhat-release"
	// detect if the system is centos or red hat
	if si.OS.Vendor == "" {
		data, err := ioutil.ReadFile(fileName)
		if err != nil {
			return
		}
		if strings.Contains(string(data), "CentOS") {
			si.OS.Vendor = "centos"
		}

		if strings.Contains(string(data), "Red Hat") {
			si.OS.Vendor = "rhel"
		}
	}
	switch si.OS.Vendor {
	case "centos":
		{
			// if centos release not exist then use red hat file
			if file_is_exists("/etc/centos-release") {
				fileName = "/etc/centos-release"
			}
			if release := slurpFile(fileName); release != "" {
				if m := reCentOS.FindStringSubmatch(release); m != nil {
					si.OS.Release = m[2]
				}
			}
		}
	case "rhel":
		{
			if release := slurpFile("/etc/redhat-release"); release != "" {
				if m := reRedHat.FindStringSubmatch(release); m != nil {
					si.OS.Release = m[1]
				}
			}
			if si.OS.Release == "" {
				if m := reRedHat.FindStringSubmatch(si.OS.Name); m != nil {
					si.OS.Release = m[1]
				}
			}
		}
	}
	// fix centos 5/6 and red hat 5/6 release bug
	if si.OS.Release != "" {
		if strings.HasPrefix(si.OS.Release, "5.") {
			si.OS.Version = "5"
		}
		if strings.HasPrefix(si.OS.Release, "6.") {
			si.OS.Version = "6"
		}
	}
}

func (si *SysInfo) getOSInfo() {
	// This seems to be the best and most portable way to detect OS architecture (NOT kernel!)
	if _, err := os.Stat("/lib64/ld-linux-x86-64.so.2"); err == nil {
		si.OS.Architecture = "amd64"
	} else if _, err := os.Stat("/lib/ld-linux.so.2"); err == nil {
		si.OS.Architecture = "i386"
	}

	f, err := os.Open("/etc/os-release")
	if err != nil {
		getCentosOrRedHatOldSystem(si)
		return
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		if m := rePrettyName.FindStringSubmatch(s.Text()); m != nil {
			si.OS.Name = strings.Trim(m[1], `"`)
		} else if m := reID.FindStringSubmatch(s.Text()); m != nil {
			si.OS.Vendor = strings.Trim(m[1], `"`)
		} else if m := reVersionID.FindStringSubmatch(s.Text()); m != nil {
			si.OS.Version = strings.Trim(m[1], `"`)
		}
	}

	switch si.OS.Vendor {
	case "debian":
		si.OS.Release = slurpFile("/etc/debian_version")
	case "ubuntu":
		if m := reUbuntu.FindStringSubmatch(si.OS.Name); m != nil {
			si.OS.Release = m[1]
		}
	case "centos":
		if release := slurpFile("/etc/centos-release"); release != "" {
			if m := reCentOS.FindStringSubmatch(release); m != nil {
				si.OS.Release = m[2]
			}
		}
	case "rhel":
		if release := slurpFile("/etc/redhat-release"); release != "" {
			if m := reRedHat.FindStringSubmatch(release); m != nil {
				si.OS.Release = m[1]
			}
		}
		if si.OS.Release == "" {
			if m := reRedHat.FindStringSubmatch(si.OS.Name); m != nil {
				si.OS.Release = m[1]
			}
		}
	}
}
