package main

import (
	"flag"
	"github.com/mabu/qa"
	"log"
	"net"
	"strconv"
	"time"
)

// retry looking for a server every
const RETRY_INTERVAL = 10 * time.Second

func main() {
	port := flag.Int("p", qa.PORT, "server port")
	interval := flag.Int("i", 1, "minimum interval between sending messages (seconds)")
	flag.Parse()
	servInfo := connect(*port)
	if *interval < servInfo.interval {
		*interval = servInfo.interval
	}
	log.Printf("Connected to %s, will send data every %d seconds.",	servInfo.addr, *interval)
	conn, err := net.DialUDP("udp4", nil, servInfo.addr)
	if err != nil {
		panic(err)
	}
	send := func() {
		log.Println("Sending data to", conn.RemoteAddr())
		_, err := conn.Write(collect())
		if err != nil {
			panic(err)
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

func connect(port int) *serverInfo {
	conn, err := net.ListenUDP("udp4", nil)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	greet := func() {
		log.Println("Greeting")
		_, err := conn.WriteToUDP([]byte(qa.GREETING), &net.UDPAddr{
			IP:   net.IPv4bcast,
			Port: port,
		})
		if err != nil {
			panic(err)
		}
	}
	ticker, resChan := time.Tick(RETRY_INTERVAL), make(chan *serverInfo)
	go listen(conn, resChan)
	greet()
	for {
		select {
		case <-ticker:
			greet()
		case r := <-resChan:
			return r
		}
	}
	return nil
}

func listen(conn *net.UDPConn, resChan chan<- *serverInfo) {
	data := make([]byte, 32)
	log.Println("Listening on", conn.LocalAddr())
	n, addr, err := conn.ReadFromUDP(data)
	data = data[:n]
	log.Println("Got", data, "from", addr)
	if err != nil {
		panic(err)
	}
	interval, err := strconv.Atoi(string(data))
	if err != nil {
		panic(err)
	}
	resChan <- &serverInfo{interval, addr}
}
