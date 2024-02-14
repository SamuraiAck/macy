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
	"fmt"
	"github.com/dlclark/regexp2"
	"github.com/klauspost/compress/zstd"
	"github.com/spf13/pflag"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	// User-controllable
	Group          net.IP
	Port           int
	TTL            int
	Rate           int
	QoS            int
	Fragments      bool
	Size           int
	LinkLocal      bool
	AddressRegex   string
	InterfaceRegex string
	Verbose        bool

	// Automatic
	Host      string
	Transport string

	// Data
	Mutex      = new(sync.Mutex)
	HeardHosts = make(map[string]time.Time)
	HeardIPs   = make(map[string]time.Time)
	HeardDb    = make(map[string]map[string]time.Duration)

	// Other
	AddressFilter   *regexp2.Regexp
	InterfaceFilter *regexp2.Regexp
	ZstdEncoder     *zstd.Encoder
	ZstdDecoder     *zstd.Decoder
)

func Configure() {
	var err error

	// Get Hostname
	Host, err = os.Hostname()
	if err != nil {
		Fatal("os.Hostname: %v", err)
	}
	if i := strings.Index(Host, "."); i != -1 {
		Host = Host[:i]
	}
	Info("Host = %s", Host)

	// Read command-line options
	flags := pflag.NewFlagSet("macy", pflag.ExitOnError)
	flags.SortFlags = false
	flags.IPVarP(&Group, "group", "g", net.ParseIP("239.239.239.239"), "multicast group address")
	flags.IntVarP(&Port, "port", "p", 23923, "UDP port number")
	flags.IntVarP(&TTL, "ttl", "t", 1, "maximum hop count aka Time To Live")
	flags.IntVarP(&Rate, "rate", "r", 2, "transmit rate in hertz")
	flags.IntVarP(&QoS, "qos", "q", 0, "DiffServ CodePoint for QoS (default 0)")
	flags.BoolVarP(&Fragments, "fragments", "f", false, "allow packet fragmentation")
	flags.IntVarP(&Size, "size", "s", 0, "payload size before fragmentation (default 0)")
	flags.BoolVarP(&LinkLocal, "linklocal", "l", false, "include link-local addresses")
	flags.StringVarP(&AddressRegex, "addresses", "a", "", "use addresses that match this regex (default \"\")")
	flags.StringVarP(&InterfaceRegex, "interfaces", "i", "", "use interfaces that match this regex (default \"\")")
	flags.BoolVarP(&Verbose, "verbose", "v", false, "include debug messages in log")
	var help bool
	flags.BoolVarP(&help, "help", "h", false, "display usage information")
	err = flags.MarkHidden("help")
	if err != nil {
		Warn("flags.MarkHidden: %v", err)
	}
	err = flags.Parse(os.Args[1:])
	if err != nil {
		Warn("flags.Parse(%v): %v", os.Args[1:], err)
	}
	if help {
		fmt.Printf("Usage of macy:\n")
		flags.PrintDefaults()
		os.Exit(0)
	}

	// Check command line options
	Info("Group = %v", Group)
	if !Group.IsMulticast() {
		Fatal("%s is not a multicast group address", Group)
	}
	if Group.To4() != nil {
		Transport = "udp4"
	} else {
		Transport = "udp6"
	}
	Info("Transport = %v", Transport)

	Info("Port = %v", Port)
	if Port < 1 || Port > 65535 {
		Fatal("Port must be between 1 and 65535")
	}

	Info("TTL = %v", TTL)
	if TTL < 0 || TTL > 255 {
		Fatal("TTL must be between 0 and 255")
	}
	if TTL <= 1 {
		Warn("Reports will not be forwarded beyond the attached subnets")
	}

	Info("Rate = %v Hz", Rate)
	if Rate < 1 {
		Fatal("Rate must be greater than 0")
	}

	Info("QoS = %v", QoS)
	if QoS < 0 || QoS > 63 {
		Fatal("QoS must be between 0 and 63")
	}

	Info("Fragments = %v", Fragments)

	Info("Size = %v", Size)
	switch Transport {
	case "udp4":
		if Size < 0 || Size > 65507 {
			Fatal("Size must be between 0 and 65507 for IPv4")
		}
	case "udp6":
		if Size < 0 || Size > 65527 {
			Fatal("Size must be between 0 and 65527 for IPv6")
		}
	}

	Info("LinkLocal = %v", LinkLocal)

	// Compile regex engines
	Info("Address regex = \"%s\"", AddressRegex)
	AddressFilter, err = regexp2.Compile(AddressRegex, regexp2.IgnoreCase)
	if err != nil {
		Fatal("regexp2.Compile(%s): %v", AddressRegex, err)
	}
	Info("Interface regex = \"%s\"", InterfaceRegex)
	InterfaceFilter, err = regexp2.Compile(InterfaceRegex, regexp2.IgnoreCase)
	if err != nil {
		Fatal("regexp2.Compile(%s): %v", InterfaceRegex, err)
	}

	// Initialize zstd en/decoders
	ZstdDecoder, _ = zstd.NewReader(nil)
	if Size == 0 {
		ZstdEncoder, _ = zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.SpeedBestCompression))
	} else {
		ZstdEncoder, _ = zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.SpeedBestCompression), zstd.WithEncoderPadding(Size))
	}

	Info("Verbose = %v", Verbose)
}
