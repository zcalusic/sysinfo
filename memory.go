// Copyright © 2016 Zlatko Čalušić
//
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file.

package sysinfo

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"syscall"
)

// Memory Detailed information.
type MemoryDevice struct {
	Type                 string   `json:"type,omitempty"`
	TypeDetail           []string `json:"typeDetail,omitempty"`
	Speed                uint     `json:"speed,omitempty"` // RAM data rate in MT/s
	Size                 uint     `json:"size,omitempty"`  // RAM size in MB
	DataWidth            uint     `json:"dataWidth,omitempty"`
	Factor               string   `json:"factor,omitempty"`
	Locator              string   `json:"locator,omitempty"`
	Bank                 string   `json:"bank,omitempty"`
	Manufacturer         string   `json:"manufacturer,omitempty"`
	SerialNumber         string   `json:"serialNumber,omitempty"`
	AssetTag             string   `json:"assetTag,omitempty"`
	PartNumber           string   `json:"partNumber,omitempty"`
	ConfiguredClockSpeed uint     `json:"configuredClockSpeed,omitempty"`
}

// Memory information.
type Memory struct {
	Type     string          `json:"type,omitempty"`
	Speed    uint            `json:"speed,omitempty"`    // RAM data rate in MT/s
	Size     uint            `json:"size,omitempty"`     // RAM size in MB
	Memories []*MemoryDevice `json:"memories,omitempty"` // RAM Details
}

const epsSize = 0x1f

// ErrNotExist indicates that SMBIOS entry point could not be found.
var ErrNotExist = errors.New("SMBIOS entry point not found")

func word(data []byte, index int) uint16 {
	return binary.LittleEndian.Uint16(data[index : index+2])
}

func dword(data []byte, index int) uint32 {
	return binary.LittleEndian.Uint32(data[index : index+4])
}

func qword(data []byte, index int) uint64 {
	return binary.LittleEndian.Uint64(data[index : index+8])
}

func epsChecksum(sl []byte) (sum byte) {
	for _, v := range sl {
		sum += v
	}

	return
}

func epsValid(eps []byte) bool {
	if epsChecksum(eps) == 0 && bytes.Equal(eps[0x10:0x15], []byte("_DMI_")) && epsChecksum(eps[0x10:]) == 0 {
		return true
	}

	return false
}

func getStructureTableAddressEFI(f *os.File) (address int64, length int, err error) {
	systab, err := os.Open("/sys/firmware/efi/systab")
	if err != nil {
		return 0, 0, err
	}
	defer systab.Close()

	s := bufio.NewScanner(systab)
	for s.Scan() {
		sl := strings.Split(s.Text(), "=")
		if len(sl) != 2 || sl[0] != "SMBIOS" {
			continue
		}

		addr, err := strconv.ParseInt(sl[1], 0, 64)
		if err != nil {
			return 0, 0, err
		}

		eps, err := syscall.Mmap(int(f.Fd()), addr, epsSize, syscall.PROT_READ, syscall.MAP_SHARED)
		if err != nil {
			return 0, 0, err
		}
		defer syscall.Munmap(eps)

		if !epsValid(eps) {
			break
		}

		return int64(dword(eps, 0x18)), int(word(eps, 0x16)), nil
	}
	if err := s.Err(); err != nil {
		return 0, 0, err
	}

	return 0, 0, ErrNotExist
}

func getStructureTableAddress(f *os.File) (address int64, length int, err error) {
	// SMBIOS Reference Specification Version 3.0.0, page 21
	mem, err := syscall.Mmap(int(f.Fd()), 0xf0000, 0x10000, syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		return 0, 0, err
	}
	defer syscall.Munmap(mem)

	for i := range mem {
		if i > len(mem)-epsSize {
			break
		}

		// Search for the anchor string on paragraph (16 byte) boundaries.
		if i%16 != 0 || !bytes.Equal(mem[i:i+4], []byte("_SM_")) {
			continue
		}

		eps := mem[i : i+epsSize]
		if !epsValid(eps) {
			continue
		}

		return int64(dword(eps, 0x18)), int(word(eps, 0x16)), nil
	}

	return 0, 0, ErrNotExist
}

func getStructureTable() ([]byte, error) {
	f, err := os.Open("/dev/mem")
	if err != nil {
		dmi, err := ioutil.ReadFile("/sys/firmware/dmi/tables/DMI")
		if err != nil {
			return nil, err
		}
		return dmi, nil
	}
	defer f.Close()

	address, length, err := getStructureTableAddressEFI(f)
	if err != nil {
		if address, length, err = getStructureTableAddress(f); err != nil {
			return nil, err
		}
	}

	// Mandatory page aligning for mmap() system call, lest we get EINVAL
	align := address & (int64(os.Getpagesize()) - 1)
	mem, err := syscall.Mmap(int(f.Fd()), address-align, length+int(align), syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		return nil, err
	}

	return mem[align:], nil
}

func dmiString(dmiRawData []byte, baseOffset int, offset int) string {
	var slot = int(dmiRawData[baseOffset+offset])
	if slot == 0 {
		return "Not Specified"
	}
	var dmiLen = int(dmiRawData[baseOffset+1])
	var lastOffset = baseOffset + dmiLen
	for i := lastOffset; i < len(dmiRawData); i++ {
		if bytes.Equal(dmiRawData[i:i+2], []byte{0, 0}) {
			lastOffset = i + 2
			break
		}
	}
	var dmiData = dmiRawData[baseOffset:lastOffset]
	var dmiAdditionData = bytes.Split(dmiData[dmiLen:lastOffset], []byte{0})
	return strings.TrimSpace(string(dmiAdditionData[slot-1]))
}

func (si *SysInfo) getMemoryInfo() {
	mem, err := getStructureTable()
	if err != nil {
		if targetKB := slurpFile("/sys/devices/system/xen_memory/xen_memory0/target_kb"); targetKB != "" {
			si.Memory.Type = "DRAM"
			size, _ := strconv.ParseUint(targetKB, 10, 64)
			si.Memory.Size = uint(size) / 1024
		}
		return
	}
	defer syscall.Munmap(mem)

	var memSizeAlt uint
loop:
	for p := 0; p < len(mem)-1; {
		recType := mem[p]
		recLen := mem[p+1]

		switch recType {
		case 4:
			if si.CPU.Speed == 0 {
				si.CPU.Speed = uint(word(mem, p+0x16))
			}
		case 17:
			if si.Memory.Memories == nil {
				si.Memory.Memories = make([]*MemoryDevice, 0)
			}

			size := uint(word(mem, p+0x0c))
			if size == 0 || size == 0xffff || size&0x8000 == 0x8000 {
				break
			}
			if size == 0x7fff {
				if recLen >= 0x20 {
					size = uint(dword(mem, p+0x1c))
				} else {
					break
				}
			}
			si.Memory.Size += size

			var memType string
			// SMBIOS Reference Specification Version 3.0.0, page 92
			memTypes := [...]string{
				"Other", "Unknown", "DRAM", "EDRAM", "VRAM", "SRAM", "RAM", "ROM", "FLASH",
				"EEPROM", "FEPROM", "EPROM", "CDRAM", "3DRAM", "SDRAM", "SGRAM", "RDRAM",
				"DDR", "DDR2", "DDR2 FB-DIMM", "Reserved", "Reserved", "Reserved", "DDR3",
				"FBD2", "DDR4", "LPDDR", "LPDDR2", "LPDDR3", "LPDDR4",
			}
			if index := int(mem[p+0x12]); index >= 1 && index <= len(memTypes) {
				memType = memTypes[index-1]
			}
			if si.Memory.Type == "" {
				si.Memory.Type = memType
			}

			var memTypeDetailed = make([]string, 0)
			typeDetail := [...]string{
				"Other", "Unknown", "Fast-paged", "Static Column", "Pseudo-static",
				"RAMBus", "Synchronous", "CMOS", "EDO", "Window DRAM", "Cache DRAM",
				"Non-Volatile", "Registered (Buffered)", "Unbuffered (Unregistered)",
				"LRDIMM",
			}
			if w := word(mem, p+0x13); w&0xfffe != 0 {
				for i := 1; i <= len(typeDetail); i++ {
					if w&(1<<uint(i)) != 0 {
						memTypeDetailed = append(memTypeDetailed, typeDetail[i-1])
					}
				}
			}

			var memSpeed uint
			if recLen >= 0x17 {
				if speed := uint(word(mem, p+0x15)); speed != 0 {
					if si.Memory.Speed == 0 {
						si.Memory.Speed = speed
					}
					memSpeed = speed
				}
			}

			var memFactor string
			factors := [...]string{
				"Other", "Unknown", "SIMM", "SIP", "Chip", "DIP", "ZIP", "Proprietary Card",
				"DIMM", "TSOP", "Row Of Chips", "RIMM", "SODIMM", "SRIMM", "FB-DIMM",
			}
			if index := int(mem[p+0x0e]); index >= 1 && index <= 0x0f {
				memFactor = factors[index-1]
			}

			var mem = &MemoryDevice{
				Type:                 memType,
				TypeDetail:           memTypeDetailed,
				Speed:                memSpeed,
				Size:                 size,
				DataWidth:            uint(word(mem, p+0x0a)),
				Factor:               memFactor,
				Locator:              dmiString(mem, p, 0x10),
				Bank:                 dmiString(mem, p, 0x11),
				Manufacturer:         dmiString(mem, p, 0x17),
				SerialNumber:         dmiString(mem, p, 0x18),
				AssetTag:             dmiString(mem, p, 0x19),
				PartNumber:           dmiString(mem, p, 0x1a),
				ConfiguredClockSpeed: uint(word(mem, p+0x20)),
			}
			si.Memory.Memories = append(si.Memory.Memories, mem)
		case 19:
			start := uint(dword(mem, p+0x04))
			end := uint(dword(mem, p+0x08))
			if start == 0xffffffff && end == 0xffffffff {
				if recLen >= 0x1f {
					start64 := qword(mem, p+0x0f)
					end64 := qword(mem, p+0x17)
					memSizeAlt += uint((end64 - start64 + 1) / 1048576)
				}
			} else {
				memSizeAlt += (end - start + 1) / 1024
			}
		case 127:
			break loop
		}

		for p += int(recLen); p < len(mem)-1; {
			if bytes.Equal(mem[p:p+2], []byte{0, 0}) {
				p += 2
				break
			}
			p++
		}
	}

	// Sometimes DMI type 17 has no information, so we fall back to DMI type 19, to at least get the RAM size.
	if si.Memory.Size == 0 && memSizeAlt > 0 {
		si.Memory.Type = "DRAM"
		si.Memory.Size = memSizeAlt
	}
}
