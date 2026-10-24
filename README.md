# Sysinfo

[![Build Status](https://github.com/zcalusic/sysinfo/actions/workflows/go.yml/badge.svg)](https://github.com/zcalusic/sysinfo/actions/workflows/go.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/zcalusic/sysinfo.svg)](https://pkg.go.dev/github.com/zcalusic/sysinfo)
[![License](https://img.shields.io/badge/license-MIT-a31f34.svg?maxAge=2592000)](https://github.com/zcalusic/sysinfo/blob/master/LICENSE)
[![Powered by](https://img.shields.io/badge/powered_by-Go-5272b4.svg?maxAge=2592000)](https://go.dev/)
[![Platform](https://img.shields.io/badge/platform-Linux-009bde.svg?maxAge=2592000)](https://www.linuxfoundation.org/)

Package sysinfo is a Go library providing Linux OS / kernel / hardware system information. It's completely standalone,
has no dependencies on the host system and doesn't execute external programs. It collects only "inventory type"
information, things that don't change often.

## Code Example

This is also the source of the demo utility in [cmd/sysinfo](cmd/sysinfo/main.go).

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/user"

	"github.com/zcalusic/sysinfo"
)

func main() {
	current, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	if current.Uid != "0" {
		log.Fatal("requires superuser privilege")
	}

	var si sysinfo.SysInfo

	si.GetSysInfo()

	data, err := json.MarshalIndent(&si, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(data))
}
```

## Motivation

I couldn't find any self-contained library that would provide the set of data/features I needed. So another sysinfo was
born.

The purpose of the library is to collect only inventory info. No metrics like CPU usage or load average will be added.
The rule of thumb is, if it's changing during the day, every day, it doesn't belong in the library.

The library should work well on any modern/supported Linux distribution. There are no plans to add support for older
unsupported Linux distributions/kernels, to keep the code clean and robust and reduce the maintenance burden.

## Requirements

Sysinfo requires:

- Linux kernel 4.2 or later
- access to /sys & /proc Linux virtual file systems
- access to various files in /etc, /var, /run FS hierarchy
- superuser privileges (to access SMBIOS/DMI table and detect RAM size and properties)

Sysinfo doesn't require ANY other external utility on the target system, which is its primary strength, IMHO.

It depends on Linux internals heavily, so there are no plans to support other operating systems.

## Installation

Add it to your module with go get.

```
go get github.com/zcalusic/sysinfo
```

There's also a very simple utility demonstrating sysinfo library capabilities. Start it (as superuser) to get pretty
formatted JSON output of all the info that sysinfo library provides. Due to its simplicity, the source code of the
utility also doubles as an example of how to use the library.

```
go install github.com/zcalusic/sysinfo/cmd/sysinfo@latest
```

---

Build demo utility in Docker container: https://github.com/mattscilipoti/compile_sysinfo

## Sample output

```json
{
  "sysinfo": {
    "version": "1.1.3",
    "timestamp": "2026-10-24T11:31:40.799929752+02:00"
  },
  "node": {
    "hostname": "demo",
    "machineid": "a391d44fa1dc4680b1141b090144449e",
    "timezone": "Europe/Zagreb"
  },
  "os": {
    "name": "Debian GNU/Linux 13 (trixie)",
    "vendor": "debian",
    "version": "13",
    "release": "13.7",
    "architecture": "amd64"
  },
  "kernel": {
    "release": "6.12.113+deb13-amd64",
    "version": "#1 SMP PREEMPT_DYNAMIC Debian 6.12.113-1 (2026-10-15)",
    "architecture": "x86_64"
  },
  "product": {
    "name": "QuantaPlex T22HF-1U",
    "vendor": "Quanta Cloud Technology Inc.",
    "version": "---",
    "serial": "QTFCTH721000C4",
    "uuid": "4aef9dc1-7cff-2000-b68c-a8c49668cbed",
    "sku": "Default string"
  },
  "board": {
    "name": "S5HF MB",
    "vendor": "Quanta Cloud Technology Inc.",
    "version": "31S5HMB0010",
    "serial": "BTH83010204",
    "assettag": "---"
  },
  "chassis": {
    "type": 23,
    "vendor": "Quanta Cloud Technology Inc.",
    "version": "35S5SCA0000",
    "serial": "QTFCTH721000C4",
    "assettag": "---"
  },
  "bios": {
    "vendor": "American Megatrends Inc.",
    "version": "3B21.Q201",
    "date": "02/17/2022"
  },
  "cpu": {
    "vendor": "AuthenticAMD",
    "model": "AMD EPYC 7532 32-Core Processor",
    "speed": 2400,
    "cache": 512,
    "cpus": 1,
    "cores": 32,
    "threads": 64
  },
  "memory": {
    "type": "DDR4",
    "speed": 2133,
    "size": 32768
  },
  "storage": [
    {
      "name": "nvme0n1",
      "model": "INTEL SSDPEDME016T4S",
      "serial": "CVMD5215015U1Q5MAF",
      "size": 1600,
      "removable": false
    }
  ],
  "network": [
    {
      "name": "eth0",
      "driver": "ixgbe",
      "macaddress": "d8:c4:97:3a:77:53",
      "port": "fibre",
      "speed": 10000
    }
  ]
}
```

## Contributors

Contributions are welcome, just open a new issue or pull request.

## License

```
The MIT License (MIT)

Copyright © 2016 Zlatko Čalušić

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```
