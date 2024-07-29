package db

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"inspection-server/pkg/common"
)

var (
	sqlTables = []string{
		`CREATE TABLE IF NOT EXISTS report (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            name TEXT,
            rating INT,
            report_time TEXT,
            data TEXT
        );`,
		`CREATE TABLE IF NOT EXISTS plan (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            name TEXT,
            timer TEXT,
            cron TEXT,
            mode INT,
            state TEXT,
            template_id TEXT,
            notify_id TEXT
        );`,
		`CREATE TABLE IF NOT EXISTS record (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            name TEXT,
            start_time TEXT,
            end_time TEXT,
            mode INT,
            report_id TEXT
        );`,
		`CREATE TABLE IF NOT EXISTS template (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            name TEXT,
            data TEXT
        );`,
		`CREATE TABLE IF NOT EXISTS notify (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            name TEXT,
            app_id TEXT,
            app_secret TEXT
        );`,
	}

	sqliteName = "sqlite.db"

	mysqlUser     = "root"
	mysqlPassword = "rancher123456"
	mysqlDB       = "mydatabase"
	mysqlHost     = "15.164.221.150"
	mysqlPort     = 3306
)

func Register() error {
	DB, err := GetDB()
	if err != nil {
		return err
	}
	defer DB.Close()

	//var sqlTables []string
	//if common.MySQL == "true" {
	//	sqlTables = mysqlTables
	//} else {
	//	sqlTables = sqliteTables
	//}

	for _, table := range sqlTables {
		_, err = DB.Exec(table)
		if err != nil {
			return err
		}
	}

	return nil
}

func GetDB() (*sql.DB, error) {
	var DB *sql.DB
	var err error
	if common.MySQL == "true" {
		DB, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", mysqlUser, mysqlPassword, mysqlHost, mysqlPort, mysqlDB))
		if err != nil {
			return nil, err
		}
	} else {
		DB, err = sql.Open("sqlite3", sqliteName)
		if err != nil {
			return nil, err
		}
	}

	return DB, nil
}
