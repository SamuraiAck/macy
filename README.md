# macy

Short for MultiCAst thingY, macy is a tool for testing multicast-enabled networks.

<img alt="Demo" src="./examples/Demo.gif" width="700" />

## Features

- Works on Linux, MacOS, and Windows.
- Supports IPv4 and IPv6. Mode is determined by the group address.
- Control the TTL, DSCP, and DF-bit. Packets can be padded to any size to check for MTU issues.
- No reliance on the system routing table, packets will be sent from all routable addresses on all interfaces by default. Link-local addresses can be enabled with a command-line switch. Addresses and interfaces can be specified with regex.
- Multiple instances can run on the same machine at the same time without interfering with each other.
- Results are displayed in table format to make problems easy to spot.

## Installation

Macy was developed on Debian 12 using the golang-go package (Go 1.19). Please report any problems encountered with other build environments.

```
sudo apt install git golang-go
git clone https://github.com/SamuraiAck/macy
cd macy
go build
```

## Usage

By default macy will use multicast group 239.239.239.239, UDP port 23923, and a TTL of 1. This will transmit multicast from all attached IPv4 addresses except those in the link-local range 169.254.0.0/16 (which can be enabled with the -l/--linklocal option). Packets will not be forwarded by adjacent routers due to the TTL, so to test multicast routing, use -t/--ttl followed by a suitable maximum hop count. To test IPv6, use -g/--group followed by an address such as ff08::239.

Options:
```
  -g, --group ip            multicast group address (default 239.239.239.239)
  -p, --port int            UDP port number (default 23923)
  -t, --ttl int             maximum hop count aka Time To Live (default 1)
  -r, --rate int            transmit rate in hertz (default 2)
  -q, --qos int             DiffServ CodePoint for QoS (default 0)
  -f, --fragments           allow packet fragmentation
  -s, --size int            payload size before fragmentation (default 0)
  -l, --linklocal           include link-local addresses
  -a, --addresses string    use addresses that match this regex (default "")
  -i, --interfaces string   use interfaces that match this regex (default "")
  -v, --verbose             include debug messages in log
```

Macy provides a TUI to display information to the user. Labels along the top identify the available views.

- The Reports view presents a table of hosts that have been heard by the current instance and the IPs those hosts have received multicast packets from. The local host and IPs are highlighted in blue. The table shows the amount of time that has passed since each IP was last heard by each host, and is updated once per second. Pressing R will return the user to the Reports view from any other view.

<img alt="Reports 1" src="./examples/Reports 1.png" width="500" />

- Pressing L switches to the Log view which shows the running configuration and various events. The command-line option -v/--verbose includes debug messages in this log.

<img alt="Log 1" src="./examples/Log 1.png" width="500" />
<img alt="Log 2" src="./examples/Log 2.png" width="500" />

- Pressing A switches to the About view, which includes the Apache License so copies of the executable are compliant without any additional files.

<img alt="About 1" src="./examples/About 1.png" width="500" />

- Pressing Q exits the program.

## Bugs

- Setting the DSCP value for IPv6 is not supported on Windows.
- When fragmentation is disabled (the default), trying to send a packet larger than an interface MTU fails silently on Windows. On Linux and MacOS, an error appears in the log as intended.

## Roadmap

- Rework rate option for testing throughput
- Display statistics when the user selects a cell in the table
- Analyze PIM packets to identify common problems
- Add support for Source-Specific Multicast (SSM)
- Add daemon mode so macy can be run as a system service
- Maybe a web interface?



