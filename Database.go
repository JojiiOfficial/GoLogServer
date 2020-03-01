package main

import (
	"strconv"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
)

var db *sqlx.DB
var dbLock sync.Mutex

func initDB(config *Config) {
	var err error
	db, err = sqlx.Open("mysql", config.Database.Username+":"+config.Database.Pass+"@tcp("+config.Database.Host+":"+strconv.Itoa(config.Database.DatabasePort)+")/"+config.Database.Username)
	if err != nil {
		panic(err)
	}
	log.Info("Connected to DB")
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
	_, err := db.Exec(query, args...)
	return err

}
