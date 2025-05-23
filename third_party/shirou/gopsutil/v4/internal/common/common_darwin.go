// SPDX-License-Identifier: BSD-3-Clause
//go:build darwin

package common

import (
	"unsafe"
)

// Library represents a dynamic library loaded by purego.
type Library struct {
	addr  uintptr
	path  string
	close func()
}

func (lib *Library) Close() {
	lib.close()
}

// System functions and symbols.
type (
	ProcPidPathFunc func(pid int32, buffer uintptr, bufferSize uint32) int32
	ProcPidInfoFunc func(pid, flavor int32, arg uint64, buffer uintptr, bufferSize int32) int32
)

const (
	MAXPATHLEN = 1024
)

type CStr []byte

func (s CStr) Length() int32 {
	// Include null terminator to make CFStringGetCString properly functions
	return int32(len(s)) + 1
}

func (s CStr) Ptr() *byte {
	if len(s) < 1 {
		return nil
	}

	return &s[0]
}

func (s CStr) Addr() uintptr {
	return uintptr(unsafe.Pointer(s.Ptr()))
}

func (s CStr) GoString() string {
	if s == nil {
		return ""
	}

	var length int
	for _, char := range s {
		if char == '\x00' {
			break
		}
		length++
	}
	return string(s[:length])
}
