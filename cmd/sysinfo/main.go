// Copyright © 2016 Zlatko Čalušić
//
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file.

// sysinfo is a very simple utility demonstrating sysinfo library capabilites. Start it (as the superuser) to get pretty
// formatted JSON output of all the info that sysinfo library provides. Due to its simplicity, the source code of the
// utility also doubles down as an example of how to use the library.
package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/zcalusic/sysinfo"
)

func main() {
	var si sysinfo.SysInfo

	si.GetSysInfo()

	data, err := json.MarshalIndent(&si, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(data))
}
