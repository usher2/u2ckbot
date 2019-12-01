package main

import (
	"encoding/binary"
	"net"
	"os"
	"testing"
)

var ips []string = []string{"1.2.3.4", "1.0.3.4", "1.2.0.4", "1.2.3.0", "1.0.0.4", "1.2.0.0",
	"11.2.3.4", "1.22.3.4", "1.2.33.4", "1.2.3.44", "1.22.33.4", "1.2.33.44", "11.22.33.4", "11.2.33.4", "11.2.3.44",
	"111.2.3.4", "1.111.3.4", "1.2.222.4", "1.2.3.111", "1.222.111.4", "1.2.111.222", "111.222.111.4", "111.2.222.4", "111.2.3.222",
	"101.2.3.4", "1.101.3.4", "1.2.202.4", "1.2.3.101", "1.20.111.4", "1.2.10.222", "111.222.10.4", "10.2.20.4", "10.2.3.20"}

func Test_IP4(t *testing.T) {
	logInit(os.Stderr, os.Stdout, os.Stderr, os.Stderr)
	for _, ip := range ips {
		i1 := parseIp4(ip)
		i2 := binary.BigEndian.Uint32(net.ParseIP(ip)[12:16])
		if i1 != i2 {
			t.Errorf("parseIP: Not match (%s): %d vs %d\n", ip, i1, i2)
		}
		s := int2Ip4(i1)
		if s != ip {
			t.Errorf("int2IP: Not match: %s vs %s\n", ip, s)
		}
	}
}
