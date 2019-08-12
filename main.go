package main

import (
	"fmt"
	"net"
	"os"

	"gopkg.in/alecthomas/kingpin.v2"
)

const Version = "0.0.1"

var (
	ifname = kingpin.Flag("interface", "interface name").Short('i').String()
	magic  = kingpin.Arg("magic", "target mac address").Required().String()
)

func Run() int {
	var addr string

	hw, err := net.ParseMAC(*magic)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	b := make([]byte, 6+(len(hw)*16))

	copy(b[0:6], []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
	for i := 0; i < 16; i++ {
		copy(b[6+(i*6):6+(i*6)+6], hw)
	}

	if len(*ifname) > 0 {
		ifi, err := net.InterfaceByName(*ifname)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		addrs, err := ifi.Addrs()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}

		for _, ar := range addrs {
			ip, _, err := net.ParseCIDR(ar.String())
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return 1
			}
			if ip.To4() != nil {
				addr = ip.String()
				break
			}
		}

		if len(addr) == 0 {
			fmt.Fprintf(os.Stderr, "interface '%s' have not bind ipv4 address\n", *ifname)
			return 1
		}
	}

	conn, err := net.ListenPacket("udp4", addr+":0")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer conn.Close()

	dst, err := net.ResolveUDPAddr("udp4", "255.255.255.255:40000")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	_, err = conn.WriteTo(b, dst)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	return 0
}

func main() {
	kingpin.Version(Version)
	kingpin.Parse()
	os.Exit(Run())
}
