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
            rating TEXT, 
            report_time TEXT,
            data TEXT
        );`,
		`CREATE TABLE IF NOT EXISTS task (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            name TEXT,
            start_time TEXT,
            end_time TEXT,
            cron TEXT,
			state TEXT,
			rating TEXT, 
            report_id TEXT,
			template_id TEXT,
            notify_id TEXT,
			task_id TEXT, 
            mode TEXT, 
			err_message TEXT
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
)

func Register() error {
	DB, err := GetDB()
	if err != nil {
		return err
	}
	defer DB.Close()

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
		DB, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", common.MySQLUser, common.MySQLPassword, common.MySQLHost, common.MySQLPort, common.MySQLDB))
		if err != nil {
			return nil, err
		}
	} else {
		DB, err = sql.Open("sqlite3", common.SQLiteName)
		if err != nil {
			return nil, err
		}
	}

	return DB, nil
}
