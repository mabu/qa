package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"
)

type server struct{ *sql.DB }

type tmplInput struct {
	loads                      [3]float64
	uptime                     int64
	Time, SrvTime              time.Time
	IP, Errors, HWAddrs, Addrs string
}

func (ti *tmplInput) Loads() string {
	return fmt.Sprintf("%.2f %.2f %.2f", ti.loads[0], ti.loads[1], ti.loads[2])
}

func (ti *tmplInput) Uptime() string {
	return (time.Duration(ti.uptime) * time.Second).String()
}

func (db *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" { // list all machines
		rows, err := db.Query("SELECT hwaddrs, group_concat(DISTINCT addrs), group_concat(DISTINCT ip), uptime, load1, load5, load15, time, srv_time, coalesce(group_concat(DISTINCT errors), '') FROM " + TABLE + " GROUP BY hwaddrs")
		if err != nil {
			panic(err)
		}
		var ti []*tmplInput
		for rows.Next() {
			data := new(tmplInput)
			if err := rows.Scan(&data.HWAddrs, &data.Addrs, &data.IP, &data.uptime, &data.loads[0], &data.loads[1], &data.loads[2], &data.Time, &data.SrvTime, &data.Errors); err != nil {
				panic(err)
			}
			ti = append(ti, data)
		}
		if err := tmplList.Execute(w, ti); err != nil {
			panic(err)
		}
	} else { // list all data received from a selected machine
		hwaddrs := r.URL.Path[1:]
		rows, err := db.Query("SELECT addrs, ip, uptime, load1, load5, load15, time, srv_time, errors FROM "+TABLE+" WHERE hwaddrs = ?", hwaddrs)
		if err != nil {
			panic(err)
		}
		var ti []*tmplInput
		for rows.Next() {
			data := new(tmplInput)
			if err := rows.Scan(&data.Addrs, &data.IP, &data.uptime, &data.loads[0], &data.loads[1], &data.loads[2], &data.Time, &data.SrvTime, &data.Errors); err != nil {
				panic(err)
			}
			ti = append(ti, data)
		}
		if err := tmplDetails.Execute(w, &struct {
			HWAddrs string
			Data    []*tmplInput
		}{hwaddrs, ti}); err != nil {
			panic(err)
		}
	}
}
