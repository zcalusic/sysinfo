// Copyright © 2016 Zlatko Čalušić
//
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file.

package sysinfo

import (
	"bytes"
	"encoding/binary"
	"os"
	"strconv"
	"strings"
)

// Memory information.
type Memory struct {
	Type            string         `json:"type,omitempty"`
	Speed           uint           `json:"speed,omitempty"`           // RAM rated (maximum) data rate in MT/s
	ConfiguredSpeed uint           `json:"configuredspeed,omitempty"` // RAM configured (running) data rate in MT/s
	Size            uint           `json:"size,omitempty"`            // RAM size in MB
	Slots           uint           `json:"slots,omitempty"`           // number of memory device sockets, populated or not
	Devices         []MemoryDevice `json:"devices,omitempty"`         // populated memory devices
}

// MemoryDevice information, for one memory module or soldered memory.
type MemoryDevice struct {
	Locator         string `json:"locator,omitempty"` // socket or board position, e.g. "DIMMA1"
	Bank            string `json:"bank,omitempty"`    // bank locator, e.g. "P0 CHANNEL A"
	Vendor          string `json:"vendor,omitempty"`
	Model           string `json:"model,omitempty"` // part number
	Serial          string `json:"serial,omitempty"`
	Size            uint   `json:"size,omitempty"` // size in MB
	Type            string `json:"type,omitempty"`
	FormFactor      string `json:"formfactor,omitempty"`
	Rank            uint   `json:"rank,omitempty"`
	Speed           uint   `json:"speed,omitempty"`           // rated (maximum) data rate in MT/s
	ConfiguredSpeed uint   `json:"configuredspeed,omitempty"` // configured (running) data rate in MT/s
	Voltage         uint   `json:"voltage,omitempty"`         // configured voltage in mV
}

// SMBIOS Reference Specification Version 3.8.0, 7.18.2.
var memoryTypes = [...]string{
	"Other", "Unknown", "DRAM", "EDRAM", "VRAM", "SRAM", "RAM", "ROM", "FLASH",
	"EEPROM", "FEPROM", "EPROM", "CDRAM", "3DRAM", "SDRAM", "SGRAM", "RDRAM",
	"DDR", "DDR2", "DDR2 FB-DIMM", "Reserved", "Reserved", "Reserved", "DDR3",
	"FBD2", "DDR4", "LPDDR", "LPDDR2", "LPDDR3", "LPDDR4", "Logical non-volatile device",
	"HBM", "HBM2", "DDR5", "LPDDR5", "HBM3",
}

// SMBIOS Reference Specification Version 3.8.0, 7.18.1.
var memoryFormFactors = [...]string{
	"Other", "Unknown", "SIMM", "SIP", "Chip", "DIP", "ZIP", "Proprietary Card",
	"DIMM", "TSOP", "Row of chips", "RIMM", "SODIMM", "SRIMM", "FB-DIMM", "Die",
}

// Strings firmware puts in memory device fields that hold no information.
var dmiPlaceholders = map[string]bool{
	"not specified": true, "unknown": true, "to be filled by o.e.m.": true, "default string": true,
	"none": true, "n/a": true, "no dimm": true,
}

// lookup returns the 1-based entry index of names, or "" if out of range.
func lookup(names []string, index byte) string {
	if index >= 1 && int(index) <= len(names) {
		return names[index-1]
	}
	return ""
}

func word(data []byte, index int) uint16 {
	return binary.LittleEndian.Uint16(data[index : index+2])
}

func dword(data []byte, index int) uint32 {
	return binary.LittleEndian.Uint32(data[index : index+4])
}

func qword(data []byte, index int) uint64 {
	return binary.LittleEndian.Uint64(data[index : index+8])
}

// memorySpeed returns a speed in MT/s from a DMI type 17 record: the word at index, or, if that is 0xffff (SMBIOS 3.3),
// the extended speed dword at ext. It returns 0 if the speed is unknown or the record is too short.
func memorySpeed(rec []byte, index, ext int) uint {
	if len(rec) < index+2 {
		return 0
	}
	speed := uint(word(rec, index))
	if speed == 0xffff {
		if len(rec) < ext+4 {
			return 0
		}
		speed = uint(dword(rec, ext) & 0x7fffffff)
	}
	return speed
}

func (si *SysInfo) getMemoryInfo() {
	dmi, err := os.ReadFile("/sys/firmware/dmi/tables/DMI")
	if err != nil {
		// Xen hypervisor
		if targetKB := slurpFile("/sys/devices/system/xen_memory/xen_memory0/target_kb"); targetKB != "" {
			si.Memory.Type = "DRAM"
			size, _ := strconv.ParseUint(targetKB, 10, 64)
			si.Memory.Size = uint(size) / 1024
		}
		return
	}

	si.parseDMI(dmi)
}

// dmiStructure is one SMBIOS structure: its formatted area and its strings.
type dmiStructure struct {
	data    []byte
	strings [][]byte
}

// str returns the structure's string referenced by the byte at index (1-based, 0 for none), with surrounding spaces
// removed and placeholders like "Not Specified" returned as "".
func (s dmiStructure) str(index int) string {
	if index >= len(s.data) {
		return ""
	}
	n := int(s.data[index])
	if n < 1 || n > len(s.strings) {
		return ""
	}
	v := string(bytes.TrimSpace(s.strings[n-1]))
	if dmiPlaceholders[strings.ToLower(v)] {
		return ""
	}
	return v
}

// dmiStructures splits a raw SMBIOS table into structures, stopping at the end-of-table structure or at a malformed or
// truncated one.
func dmiStructures(dmi []byte) []dmiStructure {
	var list []dmiStructure
	for p := 0; p < len(dmi)-1; {
		recType := dmi[p]
		recLen := int(dmi[p+1])

		// Malformed or truncated table, stop parsing.
		if recLen < 4 || p+recLen > len(dmi) {
			break
		}

		s := dmiStructure{data: dmi[p : p+recLen]}
		end := bytes.Index(dmi[p+recLen:], []byte{0, 0})
		if end < 0 {
			// No string set terminator: the table is truncated.
			end = len(dmi) - p - recLen
		}
		if area := dmi[p+recLen : p+recLen+end]; len(area) > 0 {
			s.strings = bytes.Split(area, []byte{0})
		}
		list = append(list, s)

		if recType == 127 {
			break
		}
		p += recLen + end + 2
	}
	return list
}

// memorySize returns the size in MB of a DMI type 17 memory device, and false if no device is installed or its size is
// unknown.
func memorySize(rec []byte) (uint, bool) {
	size := uint(word(rec, 0x0c))
	switch {
	case size == 0 || size == 0xffff:
		return 0, false
	case size == 0x7fff:
		if len(rec) < 0x20 {
			return 0, false
		}
		size = uint(dword(rec, 0x1c) & 0x7fffffff)
	case size&0x8000 != 0:
		// Size in KB.
		size = (size & 0x7fff) / 1024
	}
	return size, size != 0
}

// parseDMI fills in memory information and CPU speed from a raw SMBIOS table.
func (si *SysInfo) parseDMI(dmi []byte) {
	si.Memory = Memory{}
	structures := dmiStructures(dmi)

	// Physical memory arrays used for system memory, not video memory or flash.
	systemArrays := map[uint16]bool{}
	for _, s := range structures {
		if s.data[0] == 16 && len(s.data) >= 0x07 && s.data[0x05] == 0x03 {
			systemArrays[word(s.data, 0x02)] = true
		}
	}

	var memSizeAlt uint
	for _, s := range structures {
		rec := s.data
		switch rec[0] {
		case 4:
			if si.CPU.Speed == 0 && len(rec) >= 0x18 {
				si.CPU.Speed = uint(word(rec, 0x16))
			}
		case 17:
			if len(rec) < 0x13 {
				break
			}
			if len(systemArrays) > 0 && !systemArrays[word(rec, 0x04)] {
				break
			}
			si.Memory.Slots++

			size, ok := memorySize(rec)
			if !ok {
				break
			}

			dev := MemoryDevice{
				Locator:    s.str(0x10),
				Bank:       s.str(0x11),
				Size:       size,
				Type:       lookup(memoryTypes[:], rec[0x12]),
				FormFactor: lookup(memoryFormFactors[:], rec[0x0e]),
			}
			if len(rec) >= 0x17 {
				dev.Speed = memorySpeed(rec, 0x15, 0x54)
			}
			if len(rec) >= 0x1b {
				dev.Vendor = s.str(0x17)
				dev.Serial = s.str(0x18)
				dev.Model = s.str(0x1a)
				// Firmware without serial numbers often fills them with zeros.
				if strings.Trim(dev.Serial, "0") == "" {
					dev.Serial = ""
				}
			}
			if len(rec) >= 0x1c {
				dev.Rank = uint(rec[0x1b] & 0x0f)
			}
			dev.ConfiguredSpeed = memorySpeed(rec, 0x20, 0x58)
			if len(rec) >= 0x28 {
				dev.Voltage = uint(word(rec, 0x26))
			}
			si.Memory.Devices = append(si.Memory.Devices, dev)

			// The summary: total size, the first device's type, and both speeds from the first device with a known
			// speed.
			si.Memory.Size += size
			if si.Memory.Type == "" {
				si.Memory.Type = dev.Type
			}
			if si.Memory.Speed == 0 && dev.Speed != 0 {
				si.Memory.Speed = dev.Speed
				si.Memory.ConfiguredSpeed = dev.ConfiguredSpeed
			}
		case 19:
			if len(rec) < 0x0f {
				break
			}
			start := uint(dword(rec, 0x04))
			end := uint(dword(rec, 0x08))
			if start == 0xffffffff && end == 0xffffffff {
				if len(rec) >= 0x1f {
					start64 := qword(rec, 0x0f)
					end64 := qword(rec, 0x17)
					memSizeAlt += uint((end64 - start64 + 1) / 1048576)
				}
			} else {
				memSizeAlt += (end - start + 1) / 1024
			}
		}
	}

	// Sometimes DMI type 17 has no information, so we fall back to DMI type 19, to at least get the RAM size.
	if si.Memory.Size == 0 && memSizeAlt > 0 {
		si.Memory.Type = "DRAM"
		si.Memory.Size = memSizeAlt
	}
}
