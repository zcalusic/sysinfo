// Copyright © 2016 Zlatko Čalušić
//
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file.

package sysinfo

import (
	"debug/elf"
	"encoding/binary"
	"os"
	"testing"
)

// elfHeader builds a minimal ELF header with the given class, byte order, machine and flags.
func elfHeader(class elf.Class, data elf.Data, machine elf.Machine, flags uint32) []byte {
	size, flagsOffset := 52, 36
	if class == elf.ELFCLASS64 {
		size, flagsOffset = 64, 48
	}

	var bo binary.ByteOrder = binary.BigEndian
	if data == elf.ELFDATA2LSB {
		bo = binary.LittleEndian
	}

	hdr := make([]byte, size)
	copy(hdr, elf.ELFMAG)
	hdr[elf.EI_CLASS] = byte(class)
	hdr[elf.EI_DATA] = byte(data)
	hdr[elf.EI_VERSION] = byte(elf.EV_CURRENT)
	bo.PutUint16(hdr[16:], uint16(elf.ET_EXEC))
	bo.PutUint16(hdr[18:], uint16(machine))
	bo.PutUint32(hdr[flagsOffset:], flags)

	return hdr
}

func TestELFArch(t *testing.T) {
	const (
		c32, c64 = elf.ELFCLASS32, elf.ELFCLASS64
		le, be   = elf.ELFDATA2LSB, elf.ELFDATA2MSB
	)

	for _, tc := range []struct {
		name string
		hdr  []byte
		want string
	}{
		{"amd64", elfHeader(c64, le, elf.EM_X86_64, 0), "amd64"},
		{"x32", elfHeader(c32, le, elf.EM_X86_64, 0), "x32"},
		{"i386", elfHeader(c32, le, elf.EM_386, 0), "i386"},
		{"arm64", elfHeader(c64, le, elf.EM_AARCH64, 0), "arm64"},
		{"armhf", elfHeader(c32, le, elf.EM_ARM, 0x05000400), "armhf"},
		{"armel", elfHeader(c32, le, elf.EM_ARM, 0x05000200), "armel"},
		{"riscv64", elfHeader(c64, le, elf.EM_RISCV, 0x5), "riscv64"},
		{"s390x", elfHeader(c64, be, elf.EM_S390, 0), "s390x"},
		{"ppc64el", elfHeader(c64, le, elf.EM_PPC64, 2), "ppc64el"},
		{"ppc64", elfHeader(c64, be, elf.EM_PPC64, 1), "ppc64"},
		{"powerpc", elfHeader(c32, be, elf.EM_PPC, 0), "powerpc"},
		{"loong64", elfHeader(c64, le, elf.EM_LOONGARCH, 0x43), "loong64"},
		{"mips64el", elfHeader(c64, le, elf.EM_MIPS, 0), "mips64el"},
		{"mipsel", elfHeader(c32, le, elf.EM_MIPS, 0), "mipsel"},
		{"mips", elfHeader(c32, be, elf.EM_MIPS, 0), "mips"},
		{"riscv32 is not a Debian architecture", elfHeader(c32, le, elf.EM_RISCV, 0), ""},
		{"unknown machine", elfHeader(c64, be, elf.EM_SPARCV9, 0), ""},
		{"not ELF", []byte("#!/bin/sh\n" + string(make([]byte, 60))), ""},
		{"truncated 64-bit header", elfHeader(c64, le, elf.EM_X86_64, 0)[:56], ""},
		{"empty", nil, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := elfArch(tc.hdr); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestGetOSArch(t *testing.T) {
	if _, err := os.Stat("/bin/sh"); err != nil {
		t.Skip("no /bin/sh")
	}

	if arch := getOSArch(); arch == "" {
		t.Fatal("no OS architecture detected from /bin/sh")
	}
}
