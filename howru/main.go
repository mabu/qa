package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"flag"
	"github.com/mabu/qa"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	SERVER   = ":8080"
	DATABASE = ":memory:"
	TABLE    = "data"
	SI_LOAD_SHIFT = 16
)


func main() {
	port := flag.Int("p", qa.PORT, "server port")
	interval := flag.Int("i", qa.INTERVAL, "minimum interval between sending messages (seconds)")
	database := flag.String("d", DATABASE, "SQLite database")
	addr := flag.String("s", SERVER, "web server address.")
	flag.Parse()

	db := connectDB(*database)
	defer db.Close()

	go listen(*port, []byte(strconv.Itoa(*interval)), db)
	http.ListenAndServe(*addr, &server{db})
}

func connectDB(name string) *sql.DB {
	db, err := sql.Open("sqlite3", name)
	if err != nil {
		panic(err)
	}
	var n int
	if err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?", TABLE).Scan(&n); err != nil {
		panic(err)
	}
	if n == 0 {
		if _, err := db.Exec("CREATE TABLE " + TABLE + " (id INTEGER PRIMARY KEY, uptime INTEGER, load1 REAL, load5 REAL, load15 REAL, hwaddrs TEXT, addrs TEXT, ip TEXT, time TIMESTAMP, srv_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP, errors TEXT)"); err != nil {
			panic(err)
		}
	}
	return db
}

func listen(port int, response []byte, db *sql.DB) {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: port,
	})
	if err != nil {
		panic(err)
	}
	for {
		data := make([]byte, qa.MAX_LEN)
		n, addr, err := conn.ReadFromUDP(data)
		if err != nil {
			panic(err)
		}
		data = data[:n]
		if bytes.Equal(data, []byte(qa.GREETING)) {
			log.Println("Greeting from", addr)
			_, err := conn.WriteToUDP(response, addr)
			if err != nil {
				panic(err)
			}
		} else {
			log.Println("Data", string(data), "from", addr)
			var parsed info
			json.Unmarshal(data, &parsed)
			l := parsed.loads()
			// get strings from errors
			e := make([]string, len(parsed.Errors))
			for i := range e {
				e[i] = parsed.Errors[i].Error()
			}
			_, err = db.Exec("INSERT INTO "+TABLE+" (uptime, load1, load5, load15, hwaddrs, addrs, ip, time, errors) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)", parsed.Uptime, l[0], l[1], l[2], strings.Join(parsed.HWAddrs, ", "), strings.Join(parsed.Addrs, ", "), addr.IP.String(), parsed.Time, strings.Join(e, ", "))
			if err != nil {
				panic(err)
			}
		}
	}
}

type info struct {
	Uptime  int64
	Loads   [3]uint64
	HWAddrs []string
	Addrs   []string
	Time    time.Time
	Errors  []error
}

// Get loads as floats.
func (inf *info) loads() (res [3]float64) {
	for i := range res {
		res[i] = float64(inf.Loads[i]) / (1.0 << SI_LOAD_SHIFT)
	}
	return
}
