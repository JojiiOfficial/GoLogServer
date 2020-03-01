package main

import (
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func initDB(config *Config) (*sqlx.DB, error) {
	return sqlx.Open("mysql", config.Database.Username+":"+config.Database.Pass+"@tcp("+config.Database.Host+":"+strconv.Itoa(config.Database.DatabasePort)+")/"+config.Database.Username)
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
