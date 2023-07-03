// Copyright © 2016 Zlatko Čalušić
//
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file.

package sysinfo

import (
	"bytes"
	"encoding/binary"
	"io/ioutil"
	"strconv"
)

// Memory information.
type Memory struct {
	Type  string `json:"type,omitempty"`
	Speed uint   `json:"speed,omitempty"` // RAM data rate in MT/s
	Size  uint   `json:"size,omitempty"`  // RAM size in MB
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

func GetMemoryInfo() Memory {
	memory, _ := GetMemoryInfoAndCPUSpeed()
	return memory
}

func GetMemoryInfoAndCPUSpeed() (Memory, uint) {
	memory := Memory{}
	dmi, err := ioutil.ReadFile("/sys/firmware/dmi/tables/DMI")
	if err != nil {
		// Xen hypervisor
		if targetKB := slurpFile("/sys/devices/system/xen_memory/xen_memory0/target_kb"); targetKB != "" {
			memory.Type = "DRAM"
			size, _ := strconv.ParseUint(targetKB, 10, 64)
			memory.Size = uint(size) / 1024
		}
		return memory, 0
	}

	memory.Size = 0
	var memSizeAlt uint
	var cpuSpeed uint
loop:
	for p := 0; p < len(dmi)-1; {
		recType := dmi[p]
		recLen := dmi[p+1]

		switch recType {
		case 4:
			cpuSpeed = uint(word(dmi, p+0x16))
		case 17:
			size := uint(word(dmi, p+0x0c))
			if size == 0 || size == 0xffff || size&0x8000 == 0x8000 {
				break
			}
			if size == 0x7fff {
				if recLen >= 0x20 {
					size = uint(dword(dmi, p+0x1c))
				} else {
					break
				}
			}

			memory.Size += size

			if memory.Type == "" {
				// SMBIOS Reference Specification Version 3.0.0, page 92
				memTypes := [...]string{
					"Other", "Unknown", "DRAM", "EDRAM", "VRAM", "SRAM", "RAM", "ROM", "FLASH",
					"EEPROM", "FEPROM", "EPROM", "CDRAM", "3DRAM", "SDRAM", "SGRAM", "RDRAM",
					"DDR", "DDR2", "DDR2 FB-DIMM", "Reserved", "Reserved", "Reserved", "DDR3",
					"FBD2", "DDR4", "LPDDR", "LPDDR2", "LPDDR3", "LPDDR4",
				}

				if index := int(dmi[p+0x12]); index >= 1 && index <= len(memTypes) {
					memory.Type = memTypes[index-1]
				}
			}

			if memory.Speed == 0 && recLen >= 0x17 {
				if speed := uint(word(dmi, p+0x15)); speed != 0 {
					memory.Speed = speed
				}
			}
		case 19:
			start := uint(dword(dmi, p+0x04))
			end := uint(dword(dmi, p+0x08))
			if start == 0xffffffff && end == 0xffffffff {
				if recLen >= 0x1f {
					start64 := qword(dmi, p+0x0f)
					end64 := qword(dmi, p+0x17)
					memSizeAlt += uint((end64 - start64 + 1) / 1048576)
				}
			} else {
				memSizeAlt += (end - start + 1) / 1024
			}
		case 127:
			break loop
		}

		for p += int(recLen); p < len(dmi)-1; {
			if bytes.Equal(dmi[p:p+2], []byte{0, 0}) {
				p += 2
				break
			}
			p++
		}
	}

	// Sometimes DMI type 17 has no information, so we fall back to DMI type 19, to at least get the RAM size.
	if memory.Size == 0 && memSizeAlt > 0 {
		memory.Type = "DRAM"
		memory.Size = memSizeAlt
	}
	return memory, cpuSpeed
}
