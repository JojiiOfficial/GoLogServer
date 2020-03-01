package main

import (
	"strconv"
	"strings"
	"time"

	gaw "github.com/JojiiOfficial/GoAw"

	log "github.com/sirupsen/logrus"
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
			log.Fatalln("Error updating reported logs: " + err.Error())
		}
	})()

	for _, logItem := range logs {
		if len(logItem.Message) == 0 {
			continue
		}

		messC := make(chan uint, 1)
		go insertMessage(logItem.Message, messC)

		hstnC := make(chan uint, 1)
		go (func(a chan uint) {
			hstname, hs := hostnameMap[logItem.Hostname]
			if !hs {
				insertkv(logItem.Hostname, "Hostname", a)
			}
			a <- hstname
		})(hstnC)

		var tagID uint
		if len(logItem.Tag) > 0 {
			tgname, tg := tagMap[logItem.Tag]
			if !tg {
				tagC := make(chan uint, 1)
				insertkv(logItem.Tag, "Tag", tagC)
				tagID = <-tagC
				tagMap[logItem.Tag] = tagID
			} else {
				tagID = tgname
			}
		}

		var srcID uint
		if len(logItem.Source) > 0 {
			srcName, tg := sourceMap[logItem.Source]
			if !tg {
				srcC := make(chan uint, 1)
				insertkv(logItem.Source, "Source", srcC)
				srcID = <-srcC
				sourceMap[logItem.Source] = srcID
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
			hostname := <-hstnC
			err := execDB("INSERT INTO CustomLog (client, date, src, tag, hostname, message) VALUES (?,?,?,?,?,?)",
				uid,
				(int64(logItem.Date) + startTime),
				srcID,
				tagID,
				hostname,
				messID,
			)
			if err != nil {
				log.Fatalln("Error inserting SystemdLog: " + err.Error())
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
			log.Fatalln("Error updating reported logs: " + err.Error())
		}
	})()
	for _, litem := range logs {
		if len(litem.Message) == 0 {
			continue
		}

		messC := make(chan uint, 1)
		go insertMessage(litem.Message, messC)

		var tagID uint
		if len(litem.Tag) > 0 {
			tgname, tg := tagMap[litem.Tag]
			if !tg {
				tagC := make(chan uint, 1)
				insertkv(litem.Tag, "Tag", tagC)
				tagID = <-tagC
				tagMap[litem.Tag] = tagID
			} else {
				tagID = tgname
			}
		}

		hstnC := make(chan uint, 1)
		go (func(a chan uint) {
			hstname, hs := hostnameMap[litem.Hostname]
			if !hs {
				insertkv(litem.Hostname, "Hostname", a)
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
				(int64(litem.Date) + startTime),
				hstname,
				tagID,
				litem.PID,
				litem.LogLevel,
				messID,
			)
			if err != nil {
				log.Fatalln("Error inserting SystemdLog: " + err.Error())
				return -1
			}
		}
	}
	return 1
}

func fetchLogsDB(logRequest FetchLogsRequest) (int, []SyslogEntry, []CustomLogEntry) {
	uid := IsUserValid(logRequest.Token)
	if uid <= 0 {
		return -1, nil, nil
	}

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
		sqlWHERE += "(hostname in (SELECT pk_id FROM Hostname" + hostnameWHERE + ")" + ")"
	}
	if len(sqlWHERE) > 0 {
		sqlWHERE = "AND " + sqlWHERE
	}

	customLogQuery := "SELECT date," +
		"(SELECT value FROM Hostname WHERE pk_id=hostname) as hostname, " +
		"(SELECT value FROM Tag WHERE pk_id=tag) as tag, " +
		"(SELECT value FROM Message WHERE pk_id=message) as message," +
		"ifnull((SELECT value FROM `Source` WHERE pk_id=`src`),\"\") as `source`," +
		"IFNULL((SELECT count FROM CustLogMsgCount WHERE CustLogMsgCount.msgID=pk_id),1) as count " +
		"FROM CustomLog " +
		"LEFT JOIN CustLogMsgCount ON CustLogMsgCount.msgID=CustomLog.pk_id " +
		"WHERE date > ? AND date <= ? " +
		sqlWHERE +
		"ORDER BY date " + end

	syslogQuery := "SELECT date," +
		"(SELECT value FROM Hostname WHERE pk_id=hostname) as hostname, " +
		"(SELECT value FROM Tag WHERE pk_id=tag) as tag, pid, loglevel, " +
		"(SELECT value FROM Message WHERE pk_id=message) as message," +
		"IFNULL((SELECT count FROM SyslogMsgCount WHERE SyslogMsgCount.msgID=pk_id),1) as count " +
		"FROM SystemdLog " +
		"LEFT JOIN SyslogMsgCount ON SyslogMsgCount.msgID=SystemdLog.pk_id " +
		"WHERE date > ? AND date <= ? " +
		sqlWHERE +
		"ORDER BY date " + end

	until := logRequest.Until
	if until == 0 {
		until = time.Now().Unix() + 1
	}

	var syslogs []SyslogEntry
	var custlogs []CustomLogEntry

	err := queryRows(&syslogs, syslogQuery, logRequest.Since, until)
	if err != nil {
		log.Fatalln("Couldn't fetch syslogs: " + err.Error())
		return -2, nil, nil
	}
	err = queryRows(&custlogs, customLogQuery, logRequest.Since, until)
	if err != nil {
		log.Fatalln("Couldn't fetch custlogs: " + err.Error())
		return -2, nil, nil
	}
	return 1, syslogs, custlogs
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
		data[i] = gaw.EscapeSpecialChars(s)
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
		log.Fatalln("Error getting logid: " + err.Error())
		return -1
	}
	err = execDB("INSERT INTO "+countTableName+" (msgID,count) VALUES(?,2) ON DUPLICATE KEY UPDATE count=count+1", lgID)
	if err != nil {
		log.Fatalln("Error updating MessCount: " + err.Error())
		return -1
	}
	return 1
}
