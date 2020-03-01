package main

import (
	"net"

	gaw "github.com/JojiiOfficial/GoAw"
)

//returns if ip is valid and a reason
func isIPValid(ip string) (bool, int) {
	pip := net.ParseIP(ip)
	if pip.To4() == nil {
		return false, 0
	}
	for _, reservedIP := range gaw.ReservedIPs {
		_, subnet, err := net.ParseCIDR(reservedIP)
		if err != nil {
			panic(err)
		}
		if subnet.Contains(pip) {
			return false, -1
		}
	}
	return true, 1
}
