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

// From a QEMU/KVM virtual machine (q35, SeaBIOS) with 16 GB of RAM.
func TestParseDMIQEMU(t *testing.T) {
	dmi, err := os.ReadFile("testdata/dmi-qemu-q35.bin")
	if err != nil {
		t.Fatal(err)
	}

	var si SysInfo
	si.parseDMI(dmi)
	if want := (Memory{Type: "RAM", Size: 16384}); si.Memory != want || si.CPU.Speed != 2000 {
		t.Fatalf("got memory %+v CPU speed %d, want %+v 2000", si.Memory, si.CPU.Speed, want)
	}
}

func FuzzParseDMI(f *testing.F) {
	if dmi, err := os.ReadFile("testdata/dmi-qemu-q35.bin"); err == nil {
		f.Add(dmi)
	}
	for _, tc := range dmiTests {
		f.Add(tc.table)
	}
	f.Fuzz(func(_ *testing.T, dmi []byte) {
		var si SysInfo
		si.parseDMI(dmi)
	})
}
