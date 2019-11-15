package main

import (
	"strconv"
	"strings"
	"time"
)

var hostnameMap = make(map[string]uint)
var tagMap = make(map[string]uint)
var lastMessage uint

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
		messC := make(chan uint, 1)

		go (func(chann chan uint) {
			var messID uint
			queryRow(&messID, "SELECT pk_id FROM Message WHERE value=?", log.Message)
			if messID == 0 {
				err := execDB("INSERT INTO Message (value) VALUES(?)", log.Message)
				if err != nil {
					panic(err)
				}
				err = queryRow(&messID, "SELECT MAX(pk_id) FROM Message WHERE value=?", log.Message)
				if err != nil {
					panic(err)
				}
			}
			messC <- messID
		})(messC)

		hstnC := make(chan uint, 1)
		go (func(a chan uint) {
			hstname, hs := hostnameMap[log.Hostname]
			if !hs {
				var htsDI uint
				queryRow(&htsDI, "SELECT MAX(pk_id) FROM SystemdHostname WHERE value=?", log.Hostname)
				if htsDI > 0 {
					hstname = htsDI
					hostnameMap[log.Hostname] = htsDI
				} else {
					var hstid uint
					err := execDB("INSERT INTO SystemdHostname (value) VALUES(?)", log.Hostname)
					if err != nil {
						panic(err)
					}
					err = queryRow(&hstid, "SELECT MAX(pk_id) FROM SystemdHostname WHERE value=?", log.Hostname)
					if err != nil {
						panic(err)
					}
					hostnameMap[log.Hostname] = hstid
					hstname = hstid
				}
			}
			a <- hstname
		})(hstnC)

		tgname, tg := tagMap[log.Tag]
		if !tg {
			var tsdID uint
			queryRow(&tsdID, "SELECT MAX(pk_id) FROM SystemdTag WHERE value=?", log.Tag)
			if tsdID > 0 {
				tgname = tsdID
				tagMap[log.Tag] = tsdID
			} else {
				var tgid uint
				err := execDB("INSERT INTO SystemdTag (value) VALUES(?)", log.Tag)
				if err != nil {
					panic(err)
				}
				err = queryRow(&tgid, "SELECT MAX(pk_id) FROM SystemdTag WHERE value=?", log.Tag)
				if err != nil {
					panic(err)
				}
				tagMap[log.Tag] = tgid
				tgname = tgid
			}
		}
		messID := <-messC
		if lastMessage == messID && messID != 0 {
			var lgID int
			lastMessage = messID
			err := queryRow(&lgID, "SELECT MAX(pk_id) FROM SystemdLog WHERE message=?", messID)
			if err != nil {
				LogCritical("Error getting syslogid: " + err.Error())
				return -1
			}
			err = execDB("INSERT INTO MsgCount (msgID,count) VALUES(?,2) ON DUPLICATE KEY UPDATE count=count+1", lgID)
			if err != nil {
				LogCritical("Error updating systemdMsgCount: " + err.Error())
				return -1
			}
		} else {
			lastMessage = messID
			hstname := <-hstnC
			err := execDB("INSERT INTO SystemdLog (client, date, hostname, tag, pid, loglevel, message) VALUES (?,?,?,?,?,?,?)",
				uid,
				(int64(log.Date) + startTime),
				hstname,
				tgname,
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
		sqlWHERE += "(tag in (SELECT pk_id FROM SystemdTag" + tagWhere + ")" + ")"
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
		"(SELECT value FROM SystemdTag WHERE pk_id=tag) as tag, pid, loglevel, " +
		"(SELECT value FROM Message WHERE pk_id=message) as message," +
		"IFNULL((SELECT count FROM MsgCount WHERE MsgCount.msgID=pk_id),1) as count " +
		"FROM SystemdLog " +
		"LEFT JOIN MsgCount ON MsgCount.msgID=SystemdLog.pk_id " +
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
