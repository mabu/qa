package main

import (
	"flag"
	"fmt"
	"github.com/mabu/qa"
	"log"
	"net"
	"strconv"
	"time"
)

// retry looking for a server every
const retryInterval = 10 * time.Second

func main() {
	server := flag.String("s", fmt.Sprintf("%v:%d", net.IPv4bcast, qa.Port), "server address")
	interval := flag.Int("i", 1, "minimum interval between sending messages (seconds)")
	flag.Parse()
	servInfo := connect(*server)
	if *interval < servInfo.interval {
		*interval = servInfo.interval
	}
	log.Printf("Connected to %s, will send data every %d seconds.", servInfo.addr, *interval)
	conn, err := net.DialUDP("udp4", nil, servInfo.addr)
	if err != nil {
		log.Panic(err)
	}
	send := func() {
		log.Println("Sending data to", conn.RemoteAddr())
		_, err := conn.Write(collect())
		if err != nil {
			log.Panic(err)
		}
	}
	send()
	for _ = range time.Tick(time.Duration(*interval) * time.Second) {
		send()
	}
}

type serverInfo struct {
	interval int
	addr     *net.UDPAddr
}

func connect(server string) *serverInfo {
	addr, err := net.ResolveUDPAddr("udp4", server)
	if err != nil {
		log.Panic(err)
	}
	conn, err := net.ListenUDP("udp4", nil)
	if err != nil {
		log.Panic(err)
	}
	defer conn.Close()
	greet := func() {
		log.Println("Greeting")
		_, err := conn.WriteToUDP([]byte(qa.Greeting), addr)
		if err != nil {
			log.Panic(err)
		}
	}
	ticker, resC := time.Tick(retryInterval), make(chan *serverInfo)
	go listen(conn, resC)
	greet()
	for {
		select {
		case <-ticker:
			greet()
		case r := <-resC:
			return r
		}
	}
	return nil
}

func listen(conn *net.UDPConn, resC chan<- *serverInfo) {
	data := make([]byte, 32)
	log.Println("Listening on", conn.LocalAddr())
	n, addr, err := conn.ReadFromUDP(data)
	data = data[:n]
	log.Println("Got", data, "from", addr)
	if err != nil {
		log.Panic(err)
	}
	interval, err := strconv.Atoi(string(data))
	if err != nil {
		log.Panic(err)
	}
	resC <- &serverInfo{interval, addr}
}
