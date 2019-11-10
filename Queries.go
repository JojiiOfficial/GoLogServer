package main

import (
	"strings"
)

func insertSyslogs(token string, startTime int64, logs []SyslogEntry) int {
	uid := IsUserValid(token)
	if uid <= 0 {
		return -1
	}
	go (func() {
		err := execDB("UPDATE User SET reportedLogs=reportedLogs+?,lastPush=CURRENT_TIMESTAMP WHERE pk_id=?", len(logs), uid)
		if err != nil {
			LogCritical("Error updating reported logs: " + err.Error())
		}
	})()
	for _, log := range logs {
		err := execDB("INSERT INTO SystemdLog (client, date, hostname, tag, pid, loglevel, message) VALUES (?,?,?,?,?,?,?)",
			uid,
			(int64(log.Date) + startTime),
			log.Hostname,
			log.Tag,
			log.PID,
			log.LogLevel,
			log.Message,
		)
		if err != nil {
			LogCritical("Error inserting SystemdLog: " + err.Error())
			return -1
		}
	}
	return 1
}

func fetchSyslogLogs(logRequest FetchLogsRequest) (int, []SyslogEntry) {
	uid := IsUserValid(logRequest.Token)
	if uid <= 0 {
		return -1, nil
	}
	var syslogs []SyslogEntry

	var and string

	if len(logRequest.HostnameFilter) > 0 {
		negate := false
		hnFilter := logRequest.HostnameFilter
		negate = strings.HasPrefix(hnFilter[0], "!")
		if negate {
			hnFilter[0] = hnFilter[0][1:]
		}
		inBlock := "("
		for _, e := range hnFilter {
			inBlock += "\"" + EscapeSpecialChars(e) + "\","
		}
		inBlock = inBlock[:len(inBlock)-1] + ")"
		not := ""
		if negate {
			not = "not"
		}
		and = " AND hostname " + not + " in " + inBlock
	}
	err := queryRows(&syslogs, "SELECT date, hostname, tag, pid, loglevel, message FROM SystemdLog WHERE date > ? "+and, logRequest.Since)
	if err != nil {
		LogCritical("Couldn't fetch: " + err.Error())
		return -2, nil
	}

	return 1, syslogs
}

//IsUserValid returns userid if valid or -1 if invalid
func IsUserValid(token string) int {
	sqlCheckUserValid := "SELECT User.pk_id FROM User WHERE token=? AND User.isValid=1"
	var uid int
	err := queryRow(&uid, sqlCheckUserValid, token)
	if err != nil && uid > 0 {
		return -1
	} else if err != nil {
		panic(err)
	}
	return uid
}
