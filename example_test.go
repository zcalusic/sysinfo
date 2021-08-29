// Copyright © 2016 Zlatko Čalušić
//
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file.

package sysinfo_test

import (
	"encoding/json"
	"fmt"
	"log"
	"testing"

	"github.com/zcalusic/sysinfo"
)

func Test(t *testing.T) {
	var si sysinfo.SysInfo

	si.GetSysInfo()

	data, err := json.MarshalIndent(&si, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(data))
}
