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
	"encoding/binary"
	"strings"
	"time"
)

type Report struct {
	Host  string
	Heard map[string]time.Duration
}

func MakeReport() *Report {
	r := &Report{
		Host:  Host,
		Heard: make(map[string]time.Duration),
	}
	now := time.Now()
	Mutex.Lock()
	for ip, t := range HeardIPs {
		r.Heard[ip] = now.Sub(t)
	}
	Mutex.Unlock()
	return r
}

func SendReport() {
	b := Encode0(MakeReport())
	Mutex.Lock()
	for _, s := range Senders {
		s.Send(b)
	}
	Mutex.Unlock()
}

func Encode0(r *Report) []byte {
	z := make([]byte, 0, 70000)
	z = append(z, []byte("macy")...)

	c := append([]byte{uint8(len(r.Host))}, []byte(r.Host)...)
	i := make([]byte, 8)
	for heard, duration := range r.Heard {
		c = append(c, uint8(len(heard)))
		c = append(c, []byte(heard)...)
		binary.BigEndian.PutUint64(i, uint64(duration))
		c = append(c, i...)
	}
	z = ZstdEncoder.EncodeAll(c, z)

	return z
}

func Decode(b []byte) (z *Report) {
	// Check header
	if len(b) < 4 || strings.ToUpper(string(b[0:4])) != "MACY" {
		return nil
	}
	version := 0
	if b[0] == 'M' {
		version += 1
	}
	if b[1] == 'A' {
		version += 2
	}
	if b[2] == 'C' {
		version += 4
	}
	if b[3] == 'Y' {
		version += 8
	}

	switch version {
	case 0:
		z = Decode0(b[4:])
	default:
		Debug("Decode: protocol version %d not supported", version)
		return nil
	}
	return z
}

func Decode0(b []byte) *Report {
	// Decompress
	if len(b) < 4 {
		Debug("Decode: buffer is too short to identify compression type")
		return nil
	}
	switch {
	case b[3] == 0xfd && b[2] == 0x2f && b[1] == 0xb5 && b[0] == 0x28:
		c, err := ZstdDecoder.DecodeAll(b, nil)
		if err != nil {
			Debug("Decode: zstd.DecodeAll: %v", err)
			return nil
		}
		b = c
	default:
		Debug("Decode: unrecognized compression magic %x", b[0:4])
		return nil
	}

	// Parse hostname
	if len(b) < 1 {
		Debug("Decode: buffer is too short to decode host")
		return nil
	}
	l := int(uint8(b[0]))
	if len(b) < 1+l {
		Debug("Decode: buffer is too short to decode host")
		return nil
	}
	z := &Report{
		Host:  string(b[1 : 1+l]),
		Heard: make(map[string]time.Duration),
	}
	b = b[1+l:]

	// Parse heard records
	for len(b) > 0 {
		l = int(uint8(b[0]))
		if l == 0 {
			b = b[1:]
			continue
		}
		if len(b) < 1+l+8 {
			break
		}
		z.Heard[string(b[1:1+l])] = time.Duration(binary.BigEndian.Uint64(b[1+l : 1+l+8]))
		b = b[1+l+8:]
	}

	return z
}
