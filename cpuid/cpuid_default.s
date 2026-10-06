// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !386 && !amd64 && !gccgo && !tinygo

#include "textflag.h"

TEXT ·CPUID(SB),NOSPLIT,$0-0
	RET
