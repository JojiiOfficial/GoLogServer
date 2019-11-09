package main

import "time"

// ------------- Database structs ----------------

//User user in db
type User struct {
	Pkid      int       `db:"pk_id"`
	Username  string    `db:"username"`
	Token     string    `db:"token"`
	CreatedAt time.Time `db:"createdAt"`
	IsValid   bool      `db:"isValid"`
}

// ------------- REST structs -----------------------

//StoreSyslogRequest request to store
type StoreSyslogRequest struct {
	Token     string        `json:"t"`
	StartTime int64         `json:"st"`
	Syslogs   []SyslogEntry `json:"lgs"`
}

//SyslogEntry a log entry in the syslog
type SyslogEntry struct {
	Date     int    `json:"d"`
	Hostname string `json:"h"`
	Tag      string `json:"t"`
	PID      int    `json:"p"`
	LogLevel int    `json:"l"`
	Message  string `json:"m"`
}
