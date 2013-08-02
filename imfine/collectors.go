package main

import (
	"encoding/json"
	"net"
	"syscall"
	"time"
)

var data = map[string]interface{}{
	"uptime":  collector(uptime),
	"loads":   collector(loads),
	"hwaddrs": collector(hwaddrs),
	"addrs":   collector(addrs),
	"time": collector(now),
	"errors":  collectErrors,
}

type collector func() interface{}

var collectErrors []interface{}

func (c collector) MarshalJSON() ([]byte, error) {
	defer func() {
		if r := recover(); r != nil {
			collectErrors = append(collectErrors, r)
		}
	}()
	return json.Marshal(c())
}

func sysinfo() *syscall.Sysinfo_t {
	info := &syscall.Sysinfo_t{}
	err := syscall.Sysinfo(info)
	if err != nil {
		panic(err)
	}
	return info
}

func uptime() interface{} {
	return sysinfo().Uptime
}

func loads() interface{} {
	return sysinfo().Loads
}

func hwaddrs() interface{} {
	var result []string
	ifs, err := net.Interfaces()
	if err != nil {
		panic(err)
	}
	for _, in := range ifs {
		if in.Name != "lo" {
			result = append(result, in.HardwareAddr.String())
		}
	}
	return result
}

func addrs() interface{} {
	var result []string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		panic(err)
	}
	for _, addr := range addrs {
		result = append(result, addr.String())
	}
	return result
}

func now() interface{} {
	return time.Now()
}

func collect() []byte {
	collectErrors = collectErrors[:0]
	result, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	return result
}
