package main

import (
	"strconv"
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

	hasMultipleFilter := getBoolCount([]bool{
		len(logRequest.HostnameFilter) > 0,
		len(logRequest.TagFilter) > 0,
		len(logRequest.MessageFilter) > 0,
	}) > 1

	hasFilter := len(logRequest.HostnameFilter) > 0 || len(logRequest.TagFilter) > 0 || len(logRequest.MessageFilter) > 0

	var fop string
	if hasMultipleFilter {
		if logRequest.FilterOperator {
			fop = "OR "
		} else {
			fop = "AND "
		}
	}

	var sqlWhere string
	if hasFilter {
		sqlWhere = "AND "
		if len(logRequest.HostnameFilter) > 0 {
			hostNameFilter := arrToSQL(logRequest.HostnameFilter, "hostname")
			sqlWhere += hostNameFilter + fop
		}

		if len(logRequest.TagFilter) > 0 {
			tagFilter := arrToSQL(logRequest.TagFilter, "tag")
			sqlWhere += tagFilter + fop
		}

		if len(logRequest.MessageFilter) > 0 {
			messageFilter := filterToContains(logRequest.MessageFilter, "message")
			sqlWhere += messageFilter + fop
		}
		if strings.HasSuffix(sqlWhere, fop) {
			sqlWhere = sqlWhere[:len(sqlWhere)-len(fop)]
		}
	}

	order := "ASC"
	if logRequest.Reverse {
		order = "DESC"
	}
	var end string
	if logRequest.Limit > 0 {
		end = " LIMIT " + strconv.Itoa(logRequest.Limit)
	}
	sqlQuery := "SELECT date, hostname, tag, pid, loglevel, message FROM SystemdLog WHERE date > ? " + sqlWhere + " ORDER BY date " + order + end
	err := queryRows(&syslogs, sqlQuery, logRequest.Since)
	if err != nil {
		LogCritical("Couldn't fetch: " + err.Error())
		return -2, nil
	}
	return 1, syslogs
}

func getBoolCount(arr []bool) int {
	if len(arr) == 0 {
		return 0
	}
	var c int
	for _, v := range arr {
		if v {
			c++
		}
	}
	return c
}
func filterInpArr(arr []string) (negate bool, data []string) {
	if len(arr) == 0 {
		return false, arr
	}
	data = make([]string, len(arr))
	for i, d := range arr {
		data[i] = d
	}
	negate = strings.HasPrefix(data[0], "!")
	if negate {
		data[0] = data[0][1:]
	}
	for i, s := range data {
		data[i] = EscapeSpecialChars(s)
	}
	return
}

func filterToContains(arr []string, tableName string) string {
	var and string
	if len(arr) > 0 {
		negate, hnFilter := filterInpArr(arr)
		for _, s := range hnFilter {
			and += s + "|"
		}
		var not string
		if negate {
			not = " NOT"
		}
		and = tableName + not + " REGEXP " + "\"" + and[:len(and)-1] + "\""
	}
	return and
}

func arrToSQL(arr []string, tableName string) string {
	var and string
	if len(arr) > 0 {
		negate, hnFilter := filterInpArr(arr)
		inBlock := "("
		for _, e := range hnFilter {
			inBlock += "\"" + e + "\","
		}
		inBlock = inBlock[:len(inBlock)-1] + ") "
		not := ""
		if negate {
			not = "not"
		}
		and = tableName + " " + not + " in " + inBlock
	}
	return and
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
