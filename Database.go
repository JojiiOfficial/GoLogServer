package main

import (
	"strconv"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var db *sqlx.DB
var dbLock sync.Mutex

func initDB(config *Config) {
	var err error
	db, err = sqlx.Open("mysql", config.Username+":"+config.Pass+"@tcp("+config.Host+":"+strconv.Itoa(config.DatabasePort)+")/"+config.Username)
	if err != nil {
		panic(err)
	}
}

func queryRow(a interface{}, query string, args ...interface{}) error {
	err := db.Get(a, query, args...)
	if err != nil {
		return err
	}
	return nil
}

func queryRows(a interface{}, query string, args ...interface{}) error {
	err := db.Select(a, query, args...)
	if err != nil {
		return err
	}
	return nil
}

func execDB(query string, args ...interface{}) error {
	dbLock.Lock()
	tx := db.MustBegin()
	tx.MustExec(query, args...)
	err := tx.Commit()
	dbLock.Unlock()
	if err != nil {
		return err
	}
	return nil
}
