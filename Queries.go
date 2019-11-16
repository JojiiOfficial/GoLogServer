package main

import (
	"strconv"
	"strings"
	"time"
)

var hostnameMap = make(map[string]uint)
var tagMap = make(map[string]uint)
var sourceMap = make(map[string]uint)

var lastSyslogMessage uint
var lastCustomlogMessage uint

func insertCustomlogs(token string, startTime int64, logs []CustomLogEntry) int {
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
		if len(log.Message) == 0 {
			continue
		}

		messC := make(chan uint, 1)
		go insertMessage(log.Message, messC)

		var tagID uint
		if len(log.Tag) > 0 {
			tgname, tg := tagMap[log.Tag]
			if !tg {
				tagC := make(chan uint, 1)
				insertkv(log.Tag, "Tag", tagC)
				tagID = <-tagC
				tagMap[log.Tag] = tagID
			} else {
				tagID = tgname
			}
		}

		var srcID uint
		if len(log.Source) > 0 {
			srcName, tg := sourceMap[log.Source]
			if !tg {
				srcC := make(chan uint, 1)
				insertkv(log.Source, "Source", srcC)
				srcID = <-srcC
				sourceMap[log.Source] = srcID
			} else {
				srcID = srcName
			}
		}

		messID := <-messC
		if lastCustomlogMessage == messID && messID != 0 {
			e := updateMessageCount(lastCustomlogMessage, "CustomLog", "CustLogMsgCount")
			if e == -1 {
				return -1
			}
		} else {
			lastCustomlogMessage = messID
			err := execDB("INSERT INTO CustomLog (client, date, src, tag, message) VALUES (?,?,?,?,?)",
				uid,
				(int64(log.Date) + startTime),
				srcID,
				tagID,
				messID,
			)
			if err != nil {
				LogCritical("Error inserting SystemdLog: " + err.Error())
				return -1
			}
		}

	}
	return 1
}

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
		if len(log.Message) == 0 {
			continue
		}

		messC := make(chan uint, 1)
		go insertMessage(log.Message, messC)

		var tagID uint
		if len(log.Tag) > 0 {
			tgname, tg := tagMap[log.Tag]
			if !tg {
				tagC := make(chan uint, 1)
				insertkv(log.Tag, "Tag", tagC)
				tagID = <-tagC
				tagMap[log.Tag] = tagID
			} else {
				tagID = tgname
			}
		}

		hstnC := make(chan uint, 1)
		go (func(a chan uint) {
			hstname, hs := hostnameMap[log.Hostname]
			if !hs {
				insertkv(log.Hostname, "SystemdHostname", a)
			}
			a <- hstname
		})(hstnC)

		messID := <-messC
		if lastSyslogMessage == messID && messID != 0 {
			e := updateMessageCount(lastSyslogMessage, "SystemdLog", "SyslogMsgCount")
			if e == -1 {
				return -1
			}
		} else {
			lastSyslogMessage = messID
			hstname := <-hstnC
			err := execDB("INSERT INTO SystemdLog (client, date, hostname, tag, pid, loglevel, message) VALUES (?,?,?,?,?,?,?)",
				uid,
				(int64(log.Date) + startTime),
				hstname,
				tagID,
				log.PID,
				log.LogLevel,
				messID,
			)
			if err != nil {
				LogCritical("Error inserting SystemdLog: " + err.Error())
				return -1
			}
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

	hasMessageFilter := len(logRequest.MessageFilter) > 0
	hasTagFilter := len(logRequest.TagFilter) > 0
	hasHostnameFilter := len(logRequest.HostnameFilter) > 0

	hasMultipleFilter := getBoolCount([]bool{
		hasTagFilter,
		hasHostnameFilter,
		hasMessageFilter,
	}) > 1

	hasFilter := hasHostnameFilter || hasTagFilter || hasMessageFilter

	var fop string
	if hasMultipleFilter {
		if logRequest.FilterOperator {
			fop = " OR "
		} else {
			fop = " AND "
		}
	}

	var hostnameWHERE, tagWhere, messageWHERE string
	if hasFilter {
		if hasHostnameFilter {
			hostnameWHERE = arrToSQL(logRequest.HostnameFilter)
		}

		if hasTagFilter {
			tagWhere = arrToSQL(logRequest.TagFilter)
		}

		if hasMessageFilter {
			messageWHERE = filterToContains(logRequest.MessageFilter)
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

	var sqlWHERE string
	if hasTagFilter {
		sqlWHERE += "(tag in (SELECT pk_id FROM Tag" + tagWhere + ")" + ")"
	}

	if hasMessageFilter {
		if len(sqlWHERE) > 0 {
			sqlWHERE += fop
		}
		sqlWHERE += "(message in (SELECT pk_id FROM Message" + messageWHERE + ")" + ")"
	}
	if hasHostnameFilter {
		if len(sqlWHERE) > 0 {
			sqlWHERE += fop
		}
		sqlWHERE += "(hostname in (SELECT pk_id FROM SystemdHostname" + hostnameWHERE + ")" + ")"
	}
	if len(sqlWHERE) > 0 {
		sqlWHERE = "AND " + sqlWHERE
	}
	sqlQuery := "SELECT date," +
		"(SELECT value FROM SystemdHostname WHERE pk_id=hostname) as hostname, " +
		"(SELECT value FROM Tag WHERE pk_id=tag) as tag, pid, loglevel, " +
		"(SELECT value FROM Message WHERE pk_id=message) as message," +
		"IFNULL((SELECT count FROM SyslogMsgCount WHERE SyslogMsgCount.msgID=pk_id),1) as count " +
		"FROM SystemdLog " +
		"LEFT JOIN SyslogMsgCount ON SyslogMsgCount.msgID=SystemdLog.pk_id " +
		"WHERE date > ? AND date <= ? " +
		sqlWHERE +
		"ORDER BY date " + order + end

	until := logRequest.Until
	if until == 0 {
		until = time.Now().Unix() + 1
	}
	err := queryRows(&syslogs, sqlQuery, logRequest.Since, until)
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

func filterToContains(arr []string) string {
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
		and = " WHERE value " + not + " REGEXP " + "\"" + and[:len(and)-1] + "\""
	}
	return and
}

func arrToSQL(arr []string) string {
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
		and = " WHERE value " + not + " in " + inBlock
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

func insertMessage(message string, chann chan uint) {
	var messID uint
	queryRow(&messID, "SELECT pk_id FROM Message WHERE value=?", message)
	if messID == 0 {
		err := execDB("INSERT INTO Message (value) VALUES(?)", message)
		if err != nil {
			panic(err)
		}
		err = queryRow(&messID, "SELECT MAX(pk_id) FROM Message WHERE value=?", message)
		if err != nil {
			panic(err)
		}
	}
	chann <- messID
}

func insertkv(val, tableName string, chann chan uint) {
	var tsdID uint
	queryRow(&tsdID, "SELECT MAX(pk_id) FROM "+tableName+" WHERE value=?", val)
	if tsdID == 0 {
		err := execDB("INSERT INTO "+tableName+" (value) VALUES(?)", val)
		if err != nil {
			panic(err)
		}
		err = queryRow(&tsdID, "SELECT MAX(pk_id) FROM "+tableName+" WHERE value=?", val)
		if err != nil {
			panic(err)
		}
	}
	chann <- tsdID
}
func updateMessageCount(messID uint, logTableName, countTableName string) int {
	var lgID uint
	err := queryRow(&lgID, "SELECT MAX(pk_id) FROM "+logTableName+" WHERE message=?", messID)
	if err != nil {
		LogCritical("Error getting logid: " + err.Error())
		return -1
	}
	err = execDB("INSERT INTO "+countTableName+" (msgID,count) VALUES(?,2) ON DUPLICATE KEY UPDATE count=count+1", lgID)
	if err != nil {
		LogCritical("Error updating MessCount: " + err.Error())
		return -1
	}
	return 1
}
