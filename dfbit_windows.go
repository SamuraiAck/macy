// Copyright 2024 Eric Johnson
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"golang.org/x/sys/windows"
)

const (
	// https://microsoft.github.io/windows-docs-rs/doc/windows/Win32/Networking/WinSock/constant.IP_DONTFRAGMENT.html
	// https://microsoft.github.io/windows-docs-rs/doc/windows/Win32/Networking/WinSock/constant.IPV6_DONTFRAG.html
	DFBIT = 14
)

func ClearDFbit4(fd uintptr) {
	err := windows.SetsockoptInt(windows.Handle(fd), windows.IPPROTO_IP, DFBIT, 0)
	if err != nil {
		Warn("windows.SetSockoptInt(%v, %v, %v, %v): %v", fd, windows.IPPROTO_IP, DFBIT, 0, err)
	}
}

func ClearDFbit6(fd uintptr) {
	err := windows.SetsockoptInt(windows.Handle(fd), windows.IPPROTO_IPV6, DFBIT, 0)
	if err != nil {
		Warn("windows.SetSockoptInt(%v, %v, %v: %v): %v", fd, windows.IPPROTO_IPV6, DFBIT, 0, err)
	}
}

func SetDFbit4(fd uintptr) {
	err := windows.SetsockoptInt(windows.Handle(fd), windows.IPPROTO_IP, DFBIT, 1)
	if err != nil {
		Warn("windows.SetSockoptInt(%v, %v, %v, %v): %v", fd, windows.IPPROTO_IP, DFBIT, 1, err)
	}
}

func SetDFbit6(fd uintptr) {
	err := windows.SetsockoptInt(windows.Handle(fd), windows.IPPROTO_IPV6, DFBIT, 1)
	if err != nil {
		Warn("windows.SetSockoptInt(%v, %v, %v: %v): %v", fd, windows.IPPROTO_IPV6, DFBIT, 1, err)
	}
}
