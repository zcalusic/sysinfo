// Copyright © 2016 Zlatko Čalušić
//
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file.

package sysinfo

import (
	"bufio"
	"cmp"
	"debug/elf"
	"encoding/binary"
	"io"
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
	reAlma       = regexp.MustCompile(`^AlmaLinux release ([\d\.]+)`)
	reCentOS     = regexp.MustCompile(`^CentOS( Linux)? release ([\d\.]+)`)
	reRocky      = regexp.MustCompile(`^Rocky Linux release ([\d\.]+)`)
	reRedHat     = regexp.MustCompile(`[\( ]([\d\.]+)`)
)

// EF_ARM_ABI_FLOAT_HARD from the ARM ELF ABI, set in e_flags of hard-float binaries.
const efARMABIFloatHard = 0x400

// Architecture from an ELF header, in Debian's naming.
func elfArch(hdr []byte) string {
	if len(hdr) < 52 || string(hdr[:4]) != elf.ELFMAG {
		return ""
	}

	is64 := elf.Class(hdr[elf.EI_CLASS]) == elf.ELFCLASS64
	isLE := elf.Data(hdr[elf.EI_DATA]) == elf.ELFDATA2LSB

	var bo binary.ByteOrder = binary.BigEndian
	if isLE {
		bo = binary.LittleEndian
	}

	// e_flags follows e_entry, e_phoff and e_shoff, which are 4 or 8 bytes long.
	flagsOffset := 36
	if is64 {
		if len(hdr) < 64 {
			return ""
		}
		flagsOffset = 48
	}
	flags := bo.Uint32(hdr[flagsOffset:])

	switch elf.Machine(bo.Uint16(hdr[18:])) {
	case elf.EM_X86_64:
		if is64 {
			return "amd64"
		}
		return "x32"
	case elf.EM_386:
		return "i386"
	case elf.EM_AARCH64:
		return "arm64"
	case elf.EM_ARM:
		if flags&efARMABIFloatHard != 0 {
			return "armhf"
		}
		return "armel"
	case elf.EM_RISCV:
		if is64 {
			return "riscv64"
		}
	case elf.EM_S390:
		if is64 {
			return "s390x"
		}
		return "s390"
	case elf.EM_PPC64:
		if isLE {
			return "ppc64el"
		}
		return "ppc64"
	case elf.EM_PPC:
		return "powerpc"
	case elf.EM_LOONGARCH:
		if is64 {
			return "loong64"
		}
	case elf.EM_MIPS:
		switch {
		case is64 && isLE:
			return "mips64el"
		case is64:
			return "mips64"
		case isLE:
			return "mipsel"
		default:
			return "mips"
		}
	}

	return ""
}

func getOSArch() string {
	// The ELF header of a system binary tells the userland architecture (NOT kernel!), whatever the libc.
	if f, err := os.Open("/bin/sh"); err == nil {
		defer func() { _ = f.Close() }()
		hdr := make([]byte, 64)
		if n, err := io.ReadAtLeast(f, hdr, 52); err == nil {
			if arch := elfArch(hdr[:n]); arch != "" {
				return arch
			}
		}
	}

	// Fall back to the dynamic loader.
	if _, err := os.Stat("/lib64/ld-linux-x86-64.so.2"); err == nil {
		return "amd64"
	} else if _, err := os.Stat("/lib/ld-linux.so.2"); err == nil {
		return "i386"
	}

	return ""
}

func (si *SysInfo) getOSInfo() {
	si.OS.Architecture = getOSArch()

	f, err := os.Open("/etc/os-release")
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()

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
	if s.Err() != nil {
		return
	}

	switch si.OS.Vendor {
	case "debian":
		si.OS.Release = slurpFile("/etc/debian_version")
	case "ubuntu":
		if m := reUbuntu.FindStringSubmatch(si.OS.Name); m != nil {
			si.OS.Release = m[1]
		}
	case "almalinux":
		if release := slurpFile("/etc/almalinux-release"); release != "" {
			if m := reAlma.FindStringSubmatch(release); m != nil {
				si.OS.Release = m[1]
			}
		}

		si.OS.Version = strings.Split(cmp.Or(si.OS.Release, si.OS.Version), ".")[0]
	case "centos":
		if release := slurpFile("/etc/centos-release"); release != "" {
			if m := reCentOS.FindStringSubmatch(release); m != nil {
				si.OS.Release = m[2]
			}
		}
	case "rocky":
		if release := slurpFile("/etc/rocky-release"); release != "" {
			if m := reRocky.FindStringSubmatch(release); m != nil {
				si.OS.Release = m[1]
			}
		}

		si.OS.Version = strings.Split(cmp.Or(si.OS.Release, si.OS.Version), ".")[0]

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
