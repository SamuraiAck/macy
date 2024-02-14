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
	"syscall"
)

func ClearDFbit4(fd uintptr) {
	err := syscall.SetsockoptInt(int(fd), syscall.IPPROTO_IP, syscall.IP_PMTUDISC, syscall.IP_PMTUDISC_DONT)
	if err != nil {
		Warn("syscall.SetSockoptInt(%v, %v, %v, %v): %v", fd, syscall.IPPROTO_IP, syscall.IP_PMTUDISC, syscall.IP_PMTUDISC_DONT, err)
	}
}

func ClearDFbit6(fd uintptr) {
	err := syscall.SetsockoptInt(int(fd), syscall.IPPROTO_IPV6, syscall.IPV6_MTU_DISCOVER, syscall.IPV6_PMTUDISC_DONT)
	if err != nil {
		Warn("syscall.SetSockoptInt(%v, %v, %v: %v): %v", fd, syscall.IPPROTO_IPV6, syscall.IPV6_MTU_DISCOVER, syscall.IPV6_PMTUDISC_DONT, err)
	}
}

func SetDFbit4(fd uintptr) {
	err := syscall.SetsockoptInt(int(fd), syscall.IPPROTO_IP, syscall.IP_PMTUDISC, syscall.IP_PMTUDISC_PROBE)
	if err != nil {
		Warn("syscall.SetSockoptInt(%v, %v, %v, %v): %v", fd, syscall.IPPROTO_IP, syscall.IP_PMTUDISC, syscall.IP_PMTUDISC_PROBE, err)
	}
}

func SetDFbit6(fd uintptr) {
	err := syscall.SetsockoptInt(int(fd), syscall.IPPROTO_IPV6, syscall.IPV6_MTU_DISCOVER, syscall.IPV6_PMTUDISC_PROBE)
	if err != nil {
		Warn("syscall.SetSockoptInt(%v, %v, %v: %v): %v", fd, syscall.IPPROTO_IPV6, syscall.IPV6_MTU_DISCOVER, syscall.IPV6_PMTUDISC_PROBE, err)
	}
}
