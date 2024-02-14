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
	"golang.org/x/sys/unix"
)

func ClearDFbit4(fd uintptr) {
	err := unix.SetsockoptInt(int(fd), unix.IPPROTO_IP, unix.IP_DONTFRAG, 0)
	if err != nil {
		Warn("unix.SetSockoptInt(%v, %v, %v, %v): %v", fd, unix.IPPROTO_IP, unix.IP_DONTFRAG, 0, err)
	}
}

func ClearDFbit6(fd uintptr) {
	err := unix.SetsockoptInt(int(fd), unix.IPPROTO_IPV6, unix.IPV6_DONTFRAG, 0)
	if err != nil {
		Warn("unix.SetSockoptInt(%v, %v, %v: %v): %v", fd, unix.IPPROTO_IPV6, unix.IPV6_DONTFRAG, 0, err)
	}
}

func SetDFbit4(fd uintptr) {
	err := unix.SetsockoptInt(int(fd), unix.IPPROTO_IP, unix.IP_DONTFRAG, 1)
	if err != nil {
		Warn("unix.SetSockoptInt(%v, %v, %v, %v): %v", fd, unix.IPPROTO_IP, unix.IP_DONTFRAG, 1, err)
	}
}

func SetDFbit6(fd uintptr) {
	err := unix.SetsockoptInt(int(fd), unix.IPPROTO_IPV6, unix.IPV6_DONTFRAG, 3)
	if err != nil {
		Warn("unix.SetSockoptInt(%v, %v, %v: %v): %v", fd, unix.IPPROTO_IPV6, unix.IPV6_DONTFRAG, 3, err)
	}
}
