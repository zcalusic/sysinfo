// Copyright © 2016 Zlatko Čalušić
//
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file.

package sysinfo

import "time"

// Meta information.
type Meta struct {
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
}

func GetMetaInfo() Meta {
	return Meta{
		Version:   Version,
		Timestamp: time.Now(),
	}
}
