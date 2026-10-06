// Copyright © 2018 Zlatko Čalušić
//
// Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.

//go:build tinygo && !(386 || amd64)

// Package cpuid gives Go programs access to CPUID opcode.
package cpuid

// CPUID returns zeros where there is no CPUID instruction, as cpuid_default.s does.
func CPUID(info *[4]uint32, ax uint32) {}
