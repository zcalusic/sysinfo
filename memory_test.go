// Copyright © 2016 Zlatko Čalušić
//
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file.

package sysinfo

import (
	"encoding/binary"
	"os"
	"testing"
)

// dmiRecord is an SMBIOS structure: the formatted area followed by an empty string set.
type dmiRecord []byte

func newRecord(recType byte, length int) dmiRecord {
	r := make(dmiRecord, length+2)
	r[0] = recType
	r[1] = byte(length)
	return r
}

func (r dmiRecord) byte(index int, v byte) dmiRecord {
	r[index] = v
	return r
}

func (r dmiRecord) word(index int, v uint16) dmiRecord {
	binary.LittleEndian.PutUint16(r[index:], v)
	return r
}

func (r dmiRecord) dword(index int, v uint32) dmiRecord {
	binary.LittleEndian.PutUint32(r[index:], v)
	return r
}

func (r dmiRecord) qword(index int, v uint64) dmiRecord {
	binary.LittleEndian.PutUint64(r[index:], v)
	return r
}

func dmiTable(records ...dmiRecord) []byte {
	var table []byte
	for _, r := range records {
		table = append(table, r...)
	}
	return append(table, newRecord(127, 4)...)
}

func memoryDevice(size uint16, memType byte, speed uint16) dmiRecord {
	return newRecord(17, 0x28).word(0x0c, size).byte(0x12, memType).word(0x15, speed)
}

func memoryArray(startKB, endKB uint32) dmiRecord {
	return newRecord(19, 0x1f).dword(0x04, startKB).dword(0x08, endKB)
}

var dmiTests = []struct {
	name     string
	table    []byte
	memory   Memory
	cpuSpeed uint
}{
	{
		name: "two DIMMs",
		table: dmiTable(
			newRecord(4, 0x2a).word(0x16, 3600),
			memoryDevice(8192, 0x1a, 3200),
			memoryDevice(8192, 0x1a, 3200),
		),
		memory:   Memory{Type: "DDR4", Speed: 3200, Size: 16384},
		cpuSpeed: 3600,
	},
	{
		name:   "extended size",
		table:  dmiTable(memoryDevice(0x7fff, 0x22, 4800).dword(0x1c, 65536)),
		memory: Memory{Type: "DDR5", Speed: 4800, Size: 65536},
	},
	{
		name: "empty and unknown slots",
		table: dmiTable(
			memoryDevice(0, 0x02, 0),
			memoryDevice(0xffff, 0x02, 0),
			memoryDevice(4096, 0x18, 1600),
		),
		memory: Memory{Type: "DDR3", Speed: 1600, Size: 4096},
	},
	{
		name:   "type 19 fallback",
		table:  dmiTable(memoryDevice(0, 0x02, 0), memoryArray(0, 0x3fffff), memoryArray(0x400000, 0x7fffff)),
		memory: Memory{Type: "DRAM", Size: 8192},
	},
	{
		name:   "type 19 fallback, extended addresses",
		table:  dmiTable(memoryArray(0xffffffff, 0xffffffff).qword(0x0f, 0x100000000).qword(0x17, 0x2ffffffff)),
		memory: Memory{Type: "DRAM", Size: 8192},
	},
	{
		name:  "records too short",
		table: dmiTable(newRecord(4, 0x16), newRecord(17, 0x12).word(0x0c, 4096), newRecord(19, 0x0e)),
	},
	{
		name:   "truncated table",
		table:  append(memoryDevice(4096, 0x1a, 2400), 17, 0x28, 0, 0),
		memory: Memory{Type: "DDR4", Speed: 2400, Size: 4096},
	},
	{
		name:   "records after end of table",
		table:  append(dmiTable(memoryDevice(4096, 0x1a, 2400)), memoryDevice(4096, 0x1a, 2400)...),
		memory: Memory{Type: "DDR4", Speed: 2400, Size: 4096},
	},
	{
		name: "configured speed",
		table: dmiTable(
			memoryDevice(0, 0x02, 0),
			memoryDevice(16384, 0x1a, 3200).word(0x20, 2666),
			memoryDevice(16384, 0x1a, 2400).word(0x20, 2400),
		),
		memory: Memory{Type: "DDR4", Speed: 3200, ConfiguredSpeed: 2666, Size: 32768},
	},
	{
		name:   "configured speed, record too short",
		table:  dmiTable(newRecord(17, 0x21).word(0x0c, 4096).byte(0x12, 0x18).word(0x15, 1600)),
		memory: Memory{Type: "DDR3", Speed: 1600, Size: 4096},
	},
	{
		name: "extended speeds",
		table: dmiTable(newRecord(17, 0x5c).word(0x0c, 16384).byte(0x12, 0x22).
			word(0x15, 0xffff).word(0x20, 0xffff).dword(0x54, 70000).dword(0x58, 0x80000000|68000)),
		memory: Memory{Type: "DDR5", Speed: 70000, ConfiguredSpeed: 68000, Size: 16384},
	},
	{
		name:   "extended speeds, record too short",
		table:  dmiTable(memoryDevice(16384, 0x22, 0xffff).word(0x20, 0xffff)),
		memory: Memory{Type: "DDR5", Size: 16384},
	},
}

func TestParseDMI(t *testing.T) {
	for _, tc := range dmiTests {
		t.Run(tc.name, func(t *testing.T) {
			var si SysInfo
			si.parseDMI(tc.table)
			if si.Memory != tc.memory || si.CPU.Speed != tc.cpuSpeed {
				t.Fatalf("got memory %+v CPU speed %d, want %+v %d", si.Memory, si.CPU.Speed, tc.memory, tc.cpuSpeed)
			}
		})
	}
}

// Real SMBIOS tables, with serial numbers, asset tags and UUIDs scrubbed. Expected values checked against dmidecode.
var dmiFiles = []struct {
	file     string
	memory   Memory
	cpuSpeed uint
}{
	// QEMU/KVM virtual machine (q35, SeaBIOS), 16 GB, no memory type or speed.
	{"dmi-qemu-q35.bin", Memory{Type: "RAM", Size: 16384}, 2000},
	// Supermicro X10DRH-CT, SMBIOS 3.0, 2x Xeon E5-2630 v3, 4x 8 GB DDR4 in 16 slots.
	{"dmi-supermicro-x10drh.bin", Memory{Type: "DDR4", Speed: 2133, ConfiguredSpeed: 1866, Size: 32768}, 2400},
	// Supermicro X10SRi-F, SMBIOS 3.0, Xeon E5-1650 v4, 4x 32 GB DDR4 in 8 slots (extended size field).
	{"dmi-supermicro-x10sri.bin", Memory{Type: "DDR4", Speed: 3200, ConfiguredSpeed: 2400, Size: 131072}, 3600},
	// ASUS ROG STRIX B850-E, SMBIOS 3.6, Ryzen 7 9700X, 4x 16 GB DDR5 (EXPO, configured above rated speed).
	{"dmi-asus-b850.bin", Memory{Type: "DDR5", Speed: 4800, ConfiguredSpeed: 6000, Size: 65536}, 3800},
	// Dell PowerEdge R240, SMBIOS 3.1, Xeon E-2274G, 2x 16 GB DDR4 in 4 slots.
	{"dmi-dell-r240.bin", Memory{Type: "DDR4", Speed: 3200, ConfiguredSpeed: 2666, Size: 32768}, 4000},
}

func TestParseDMIFiles(t *testing.T) {
	for _, tc := range dmiFiles {
		t.Run(tc.file, func(t *testing.T) {
			dmi, err := os.ReadFile("testdata/" + tc.file)
			if err != nil {
				t.Fatal(err)
			}

			var si SysInfo
			si.parseDMI(dmi)
			if si.Memory != tc.memory || si.CPU.Speed != tc.cpuSpeed {
				t.Fatalf("got memory %+v CPU speed %d, want %+v %d", si.Memory, si.CPU.Speed, tc.memory, tc.cpuSpeed)
			}
		})
	}
}

func FuzzParseDMI(f *testing.F) {
	for _, tc := range dmiFiles {
		if dmi, err := os.ReadFile("testdata/" + tc.file); err == nil {
			f.Add(dmi)
		}
	}
	for _, tc := range dmiTests {
		f.Add(tc.table)
	}
	f.Fuzz(func(_ *testing.T, dmi []byte) {
		var si SysInfo
		si.parseDMI(dmi)
	})
}
