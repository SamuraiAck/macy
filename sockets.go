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
	"code.rocketnine.space/tslocum/cview"
	"errors"
	"fmt"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
	"net"
	"syscall"
	"time"
)

var (
	Receiver *Socket
	Senders  = make(map[string]*Socket)
)

type Socket struct {
	Iface net.Interface
	IP    net.IP
	Conn  *net.UDPConn
	Conn4 *ipv4.PacketConn
	Conn6 *ipv6.PacketConn
	Err   error
	Send  func([]byte)
}

func (s *Socket) Close() {
	var err error
	if s.Conn4 != nil {
		err = s.Conn4.Close()
		if err != nil {
			Debug("Socket.Conn4.Close: %v", err)
		}
	}
	if s.Conn6 != nil {
		err = s.Conn6.Close()
		if err != nil {
			Debug("Socket.Conn6.Close: %v", err)
		}
	}
	if s.Conn != nil {
		err = s.Conn.Close()
		if err != nil {
			Debug("Socket.Conn.Close: %v", err)
		}
	}
}

func CheckSockets() {
	if Receiver != nil {
		if Receiver.Err != nil {
			Warn("Deleting Receiver due to error: %v", Receiver.Err)
			Receiver.Close()
			Receiver = nil
		}
	}

	for key, s := range Senders {
		if s.Err != nil {
			Warn("Deleting %s due to error: %v", key, s.Err)
			s.Close()
			delete(Senders, key)
		}
	}
}

func MakeSockets() {
	MakeReceiver()
	MakeSenders()
}

func MakeReceiver() {
	var err error

	if Receiver == nil {
		Info("Making Receiver for address %s port %d", Group.String(), Port)
		s := &Socket{
			IP: Group,
		}

		// By joining the group instead of the wildcard address, multiple instances of macy can receive reports at the same time
		a := net.UDPAddr{IP: Group, Port: Port}
		s.Conn, s.Err = net.ListenUDP(Transport, &a)
		if s.Err != nil {
			Warn("Receiver: net.ListenUDP(%s, %v): %v", Transport, a, s.Err)
		}

		if s.Conn != nil {
			Debug("Receiver: Conn.LocalAddr = %s", cview.Escape(s.Conn.LocalAddr().String()))

			switch Transport {
			case "udp4":
				s.Conn4 = ipv4.NewPacketConn(s.Conn)
				err = s.Conn4.SetMulticastLoopback(true)
				if err != nil {
					Warn("Receiver: Conn4.SetMulticastLoopback(true): %v", err)
				}
			case "udp6":
				s.Conn6 = ipv6.NewPacketConn(s.Conn)
				err = s.Conn6.SetMulticastLoopback(true)
				if err != nil {
					Warn("Receiver: Conn6.SetMulticastLoopback(true): %v", err)
				}
			}

			go func() {
				b := make([]byte, 70000)
				var n int
				var from *net.UDPAddr
				var t time.Time
				var r *Report
				for {
					n, from, s.Err = s.Conn.ReadFromUDP(b)
					t = time.Now()
					if n > 0 {
						r = Decode(b[:n])
						if r == nil {
							continue
						}
						Mutex.Lock()
						HeardHosts[r.Host] = t
						HeardIPs[from.IP.String()] = t
						HeardDb[r.Host] = r.Heard
						Mutex.Unlock()
					}
					if s.Err != nil {
						Warn("Receiver: Conn.ReadFromUDP(b): %v", s.Err)
						return
					}
				}
			}()

			Receiver = s
		}
	}

	// Interfaces come and go, and there's no way to see if our socket is joined to the group on a particular interface, so rejoin on all usable interfaces on each loop.
	if Receiver != nil {
		a := net.UDPAddr{IP: Group}
		for _, iface := range GetUsableInterfaces() {
			switch Transport {
			case "udp4":
				err = Receiver.Conn4.JoinGroup(&iface, &a)
				if err != nil { //nolint:staticcheck
					// This appears to generate errors if Conn4 is already joined
				}
			case "udp6":
				err = Receiver.Conn6.JoinGroup(&iface, &a)
				if err != nil { //nolint:staticcheck
					// This appears to generate errors if Conn6 is already joined
				}
			}

			groups, err := iface.MulticastAddrs()
			if err != nil {
				Warn("%s.MulticastAddrs(): %v", iface.Name, err)
			}
			key := fmt.Sprintf("%d %s groups", iface.Index, iface.Name)
			msg := fmt.Sprintf("Multicast groups joined on %s: %v", iface.Name, groups)
			if LogCandidates[key] != msg {
				LogCandidates[key] = msg
				Debug(msg)
			}
		}
	}
}

func MakeSenders() {
	for _, iface := range GetUsableInterfaces() {
		for _, ip := range GetUsableIPs(iface) {
			key := fmt.Sprintf("Sender for interface %d(%s) address %s", iface.Index, iface.Name, ip.String())
			if Senders[key] != nil {
				continue
			}

			Info("Making %s", key)
			s := &Socket{
				Iface: iface,
				IP:    ip,
			}

			a := net.UDPAddr{IP: ip}
			if ip.IsLinkLocalUnicast() {
				a.Zone = iface.Name
			}
			s.Conn, s.Err = net.ListenUDP(Transport, &a)
			if s.Err != nil {
				Warn("%s: net.ListenUDP(%s, %v): %v", key, Transport, a, s.Err)
			}

			if s.Conn != nil {
				Debug("%s: Conn.LocalAddr = %v", key, s.Conn.LocalAddr())

				rawConn, err := s.Conn.SyscallConn()
				if err != nil {
					Warn("%s: Conn.SyscallConn: %v", key, err)
				}

				switch Transport {
				case "udp4":
					// Set the DF-bit
					err = rawConn.Control(func(fd uintptr) {
						if Fragments {
							ClearDFbit4(fd)
						} else {
							SetDFbit4(fd)
						}
					})
					if err != nil {
						Warn("%s: rawConn.Control: %v", key, err)
					}

					// Set TTL, QoS, and interface this connection uses to send multicast
					s.Conn4 = ipv4.NewPacketConn(s.Conn)
					err := s.Conn4.SetMulticastTTL(TTL)
					if err != nil {
						Warn("%s: Conn4.SetMulticastTTL(%d): %v", key, TTL, err)
					}
					err = s.Conn4.SetTOS(QoS)
					if err != nil {
						Warn("%s: Conn4.SetTOS(%d): %v", key, QoS, err)
					}
					err = s.Conn4.SetMulticastInterface(&iface)
					if err != nil {
						Warn("%s: Conn4.SetMulticastInterface(%s) = %v", key, iface.Name, err)
					}

					// Convenience function for sending packets to the multicast Group
					s.Send = func(b []byte) {
						a := net.UDPAddr{IP: Group, Port: Port}
						_, err := s.Conn4.WriteTo(b, nil, &a)
						if err != nil {
							Warn("%s: %v", key, err)
							if !errors.Is(err, syscall.EMSGSIZE) {
								s.Err = err
							}
						}
					}
				case "udp6":
					// Set the DF-bit
					err = rawConn.Control(func(fd uintptr) {
						if Fragments {
							ClearDFbit6(fd)
						} else {
							SetDFbit6(fd)
						}
					})
					if err != nil {
						Warn("%s: rawConn.Control: %v", key, err)
					}

					// Set TTL, QoS, and interface this connection uses to send multicast
					s.Conn6 = ipv6.NewPacketConn(s.Conn)
					err := s.Conn6.SetMulticastHopLimit(TTL)
					if err != nil {
						Warn("%s: Conn6.SetMulticastHopLimit(%d): %v", key, TTL, err)
					}
					err = s.Conn6.SetTrafficClass(QoS)
					if err != nil {
						Warn("%s: Conn6.SetTrafficClass(%d): %v", key, QoS, err)
					}
					err = s.Conn6.SetMulticastInterface(&iface)
					if err != nil {
						Debug("%s: Conn6.SetMulticastInterface(%s) = %v", key, iface.Name, err)
					}

					// Convenience function for sending packets to the multicast Group
					s.Send = func(b []byte) {
						a := net.UDPAddr{IP: Group, Port: Port}
						_, err := s.Conn6.WriteTo(b, nil, &a)
						if err != nil {
							Warn("%s: %v", key, err)
							if !errors.Is(err, syscall.EMSGSIZE) {
								s.Err = err
							}
						}
					}
				}

				go func() {
					b := make([]byte, 70000)
					var n int
					var from *net.UDPAddr
					for {
						n, from, s.Err = s.Conn.ReadFromUDP(b)
						if n > 0 {
							Warn("%s: unexpected packet received from %v: %x", key, from, b[:n])
						}
						if s.Err != nil {
							Warn("%s: Conn.ReadFromUDP(b): %v", key, s.Err)
							return
						}
					}
				}()

				Senders[key] = s
			}
		}
	}
}

func GetUsableInterfaces() (usable []net.Interface) {
	ifaces, err := net.Interfaces()
	if err != nil {
		Warn("net.Interfaces: %v", err)
		return nil
	}

	var key, msg string
	for _, iface := range ifaces {
		match, err := InterfaceFilter.MatchString(iface.Name)
		if err != nil {
			Warn("InterfaceFilter.MatchString(%s): %v", iface.Name, err)
			continue
		}
		if !match {
			key = fmt.Sprintf("%d %s match", iface.Index, iface.Name)
			msg = fmt.Sprintf("%s does not match regex %s", iface.Name, InterfaceRegex)
			if LogCandidates[key] != msg {
				LogCandidates[key] = msg
				Debug(msg)
			}
			continue
		}
		if iface.Flags&net.FlagUp == 0 {
			key = fmt.Sprintf("%d %s up", iface.Index, iface.Name)
			msg = fmt.Sprintf("%s is not up", iface.Name)
			if LogCandidates[key] != msg {
				LogCandidates[key] = msg
				Debug(msg)
			}
			continue
		}
		if iface.Flags&net.FlagMulticast == 0 {
			key = fmt.Sprintf("%d %s multicast", iface.Index, iface.Name)
			msg = fmt.Sprintf("%s does not support multicast", iface.Name)
			if LogCandidates[key] != msg {
				LogCandidates[key] = msg
				Debug(msg)
			}
			continue
		}
		if len(GetUsableIPs(iface)) == 0 {
			key = fmt.Sprintf("%d %s addresses", iface.Index, iface.Name)
			msg = fmt.Sprintf("%s has no usable addresses", iface.Name)
			if LogCandidates[key] != msg {
				LogCandidates[key] = msg
				Debug(msg)
			}
			continue
		}

		key = fmt.Sprintf("%d %s usable", iface.Index, iface.Name)
		msg = fmt.Sprintf("%s is a usable interface", iface.Name)
		if LogCandidates[key] != msg {
			LogCandidates[key] = msg
			Debug(msg)
		}
		usable = append(usable, iface)
	}

	return usable
}

func GetUsableIPs(iface net.Interface) (usable []net.IP) {
	addrs, err := iface.Addrs()
	if err != nil {
		Warn("%s.Addrs: %v", iface.Name, err)
		return nil
	}

	var key, msg string
	for _, addr := range addrs {
		if addr.Network() != "ip+net" {
			Debug("%s network is not ip+net", addr)
			continue
		}
		ip, _, err := net.ParseCIDR(addr.String())
		if err != nil {
			Warn("net.ParseCIDR(%s): %v", addr.String(), err)
			continue
		}

		ipstr := ip.String()

		match, err := AddressFilter.MatchString(ipstr)
		if err != nil {
			Warn("AddressFilter.MatchString(%s): %v", ipstr, err)
			continue
		}
		if !match {
			key = fmt.Sprintf("%d %s %s match", iface.Index, iface.Name, ipstr)
			msg = fmt.Sprintf("%s does not match regex %s", ipstr, AddressRegex)
			if LogCandidates[key] != msg {
				LogCandidates[key] = msg
				Debug(msg)
			}
			continue
		}
		if Transport == "udp4" && ip.To4() == nil {
			key = fmt.Sprintf("%d %s %s ipv4", iface.Index, iface.Name, ipstr)
			msg = fmt.Sprintf("%s is not an IPv4 address", ipstr)
			if LogCandidates[key] != msg {
				LogCandidates[key] = msg
				Debug(msg)
			}
			continue
		}
		if Transport == "udp6" && ip.To4() != nil {
			key = fmt.Sprintf("%d %s %s ipv6", iface.Index, iface.Name, ipstr)
			msg = fmt.Sprintf("%s is not an IPv6 address", ipstr)
			if LogCandidates[key] != msg {
				LogCandidates[key] = msg
				Debug(msg)
			}
			continue
		}
		if !LinkLocal && ip.IsLinkLocalUnicast() {
			key = fmt.Sprintf("%d %s %s linklocal", iface.Index, iface.Name, ipstr)
			msg = fmt.Sprintf("%s is a link-local address but LinkLocal = false", ipstr)
			if LogCandidates[key] != msg {
				LogCandidates[key] = msg
				Debug(msg)
			}
			continue
		}

		key = fmt.Sprintf("%d %s %s usable", iface.Index, iface.Name, ipstr)
		msg = fmt.Sprintf("%s is a usable address", ipstr)
		if LogCandidates[key] != msg {
			LogCandidates[key] = msg
			Debug(msg)
		}
		usable = append(usable, ip)
	}

	return usable
}
