package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
)

func pushLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var report PushLogsRequest
	if !handleUserInput(w, r, &report) {
		return
	}
	if isStructInvalid(report) {
		sendError("input missing", w, WrongInputFormatError, 422)
		return
	}

	if len(report.Token) != 24 {
		sendError("wrong token length", w, InvalidTokenError, 422)
		return
	}

	logType := vars["logtype"]
	status := -1

	if logType == "syslog" {
		var syslogs []SyslogEntry
		mapstructure.Decode(report.Logs, &syslogs)
		status = insertSyslogs(report.Token, report.StartTime, syslogs)
	} else if logType == "custom" {
		var customLogs []CustomLogEntry
		mapstructure.Decode(report.Logs, &customLogs)
		status = insertCustomlogs(report.Token, report.StartTime, customLogs)
	}

	handleError(sendSuccess(w, status), w, ServerError, 500)
}

func fetchLogs(w http.ResponseWriter, r *http.Request) {
	var fetchRequestData FetchLogsRequest
	if !handleUserInput(w, r, &fetchRequestData) {
		return
	}
	if isStructInvalid(fetchRequestData) {
		sendError("input missing", w, WrongInputFormatError, 422)
		return
	}
	if len(fetchRequestData.Token) != 24 {
		sendError("wrong token length", w, InvalidTokenError, 422)
		return
	}
	switch fetchRequestData.LogType {
	case 0:
		{
			c := 0
			for ok := true; ok; ok = fetchRequestData.Follow {
				status, logs := fetchSyslogLogs(fetchRequestData)
				if status == -1 {
					sendError("wrong token", w, InvalidTokenError, 422)
					return
				} else if status == -2 {
					sendError("server error", w, ServerError, 500)
					return
				}

				if len(logs) == 0 && fetchRequestData.Follow && c <= 6 {
					time.Sleep(2 * time.Second)
					c++
					continue
				}
				time := time.Now().Unix()
				resp := FetchSysLogResponse{
					Time: time,
					Logs: logs,
				}
				handleError(sendSuccess(w, resp), w, ServerError, 500)
				return
			}
		}
	default:
		{
			sendError("Wrong log type", w, WrongLogType, 422)
			return
		}
	}
}

func handleUserInput(w http.ResponseWriter, r *http.Request, p interface{}) bool {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1000000000))
	if err != nil {
		LogError("ReadError: " + err.Error())
		return false
	}
	if err := r.Body.Close(); err != nil {
		LogError("ReadError: " + err.Error())
		return false
	}

	errEncode := json.Unmarshal(body, p)
	if handleError(errEncode, w, WrongInputFormatError, 422) {
		return false
	}
	return true
}

func handleError(err error, w http.ResponseWriter, message ErrorMessage, statusCode int) bool {
	if err == nil {
		return false
	}
	sendError(err.Error(), w, message, statusCode)
	return true
}

func sendError(erre string, w http.ResponseWriter, message ErrorMessage, statusCode int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if statusCode >= 500 {
		LogCritical(erre)
	} else {
		LogError(erre)
	}
	w.WriteHeader(statusCode)

	var de []byte
	var err error
	if len(string(message)) == 0 {
		de, err = json.Marshal(&ResponseError)
	} else {
		de, err = json.Marshal(&Status{"error", string(message)})
	}

	if err != nil {
		panic(err)
	}
	_, _ = fmt.Fprintln(w, string(de))
}

func isStructInvalid(x interface{}) bool {
	s := reflect.TypeOf(x)
	for i := s.NumField() - 1; i >= 0; i-- {
		e := reflect.ValueOf(x).Field(i)

		if isEmptyValue(e) {
			return true
		}
	}
	return false
}

func isEmptyValue(e reflect.Value) bool {
	switch e.Type().Kind() {
	case reflect.String:
		if e.String() == "" || strings.Trim(e.String(), " ") == "" {
			return true
		}
	case reflect.Int:
		{
			return false
		}
	case reflect.Int64:
		{
			return false
		}
	case reflect.Bool:
		{
			return false
		}
	case reflect.Interface:
		{
			return false
		}
	case reflect.Array:
		for j := e.Len() - 1; j >= 0; j-- {
			isEmpty := isEmptyValue(e.Index(j))
			if isEmpty {
				return true
			}
		}
	case reflect.Slice:
		return isStructInvalid(e)
	case reflect.Uintptr:
		{
			return false
		}
	case reflect.Ptr:
		{
			return false
		}
	case reflect.UnsafePointer:
		{
			return false
		}
	case reflect.Struct:
		{
			return false
		}
	default:
		fmt.Println(e.Type().Kind(), e)
		return true
	}
	return false
}

func sendSuccess(w http.ResponseWriter, i interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	de, err := json.Marshal(i)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(de))
	if err != nil {
		return err
	}
	return nil
}
