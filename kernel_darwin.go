//go:build darwin
// +build darwin

package sysinfo

func GetKernelInfo() Kernel {
	return Kernel{}
}
