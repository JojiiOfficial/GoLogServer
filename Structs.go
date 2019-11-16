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
	Date     int64  `json:"d" mapstructure:"d"`
	Hostname string `json:"h" mapstructure:"h"`
	Tag      string `json:"t" mapstructure:"t"`
	PID      int    `json:"p" mapstructure:"p"`
	LogLevel int    `json:"l" mapstructure:"l"`
	Message  string `json:"m" mapstructure:"m"`
	Count    int    `json:"c" mapstructure:"c"`
}

//FetchLogsRequest fetches logs from the server
type FetchLogsRequest struct {
	Token          string   `json:"t"`
	Since          int64    `json:"sin"`
	Until          int64    `json:"unt"`
	LogType        int      `json:"lt"`
	Follow         bool     `json:"foll"`
	HostnameFilter []string `json:"hnf,omitempty"`
	MessageFilter  []string `json:"mf,omitempty"`
	TagFilter      []string `json:"tf,omitempty"`
	Reverse        bool     `json:"r,omitempty"`
	FilterOperator bool     `json:"fi,omitempty"`
	Limit          int      `json:"lm,omitempty"`
}

//FetchSysLogResponse response for fetchlog
type FetchSysLogResponse struct {
	Time int64         `json:"t"`
	Logs []SyslogEntry `json:"lgs"`
}

//CustomLogEntry a log entry from a custom file
type CustomLogEntry struct {
	Date    int    `json:"d" mapstructure:"d"`
	Message string `json:"m" mapstructure:"m"`
	Tag     string `json:"t,omitempty" mapstructure:"t"`
	Source  string `json:"s" mapstructure:"s"`
}

//PushLogsRequest request to push syslog
type PushLogsRequest struct {
	Token     string      `json:"t"`
	StartTime int64       `json:"st"`
	Logs      interface{} `json:"lgs"`
}
