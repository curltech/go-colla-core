package net

import (
	"net"
	"strings"
)

func GetExternalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		panic(err)
	}
	for _, value := range addrs {
		if ipnet, ok := value.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}

	return ""
}

func GetAddrInfo(addr string) (string, string, string) {
	segs := strings.Split(addr, "/")
	if len(segs) > 4 {
		return segs[1], segs[3], segs[5]
	}
	return segs[1], segs[3], ""
}

func GetP2PAddr(addr string, peerId string) string {
	var address string
	if strings.Index(addr, ":") > 0 {
		segs := strings.Split(addr, ":")
		address = "/ip4/" + segs[0] + "/tcp/" + segs[1]
	} else {
		address = addr
	}
	address = address + "/p2p/" + peerId

	return address
}
