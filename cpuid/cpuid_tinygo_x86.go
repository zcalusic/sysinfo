// Copyright © 2018 Zlatko Čalušić
//
// Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.

//go:build tinygo && (386 || amd64)

// Package cpuid gives Go programs access to CPUID opcode.
package cpuid

/*
static void sysinfo_cpuid(unsigned int *info, unsigned int ax) {
	unsigned int a, b, c, d;
	__asm__ volatile("cpuid" : "=a"(a), "=b"(b), "=c"(c), "=d"(d) : "a"(ax), "c"(0));
	info[0] = a; info[1] = b; info[2] = c; info[3] = d;
}
*/
import "C"

// CPUID returns processor identification and feature information.
// TinyGo does not assemble .s files, so the instruction comes from C here.
func CPUID(info *[4]uint32, ax uint32) {
	var out [4]C.uint
	C.sysinfo_cpuid(&out[0], C.uint(ax))
	for i := range out {
		info[i] = uint32(out[i])
	}
}
