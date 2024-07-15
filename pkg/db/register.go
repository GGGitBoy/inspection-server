package db

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

var (
	sqlTables = []string{
		`CREATE TABLE IF NOT EXISTS report (
			id TEXT NOT NULL PRIMARY KEY,
			rating INT,
			report_time TEXT,
			warnings TEXT,
			data TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS plan (
			id TEXT NOT NULL PRIMARY KEY,
			name TEXT,
			timer TEXT,
			cron TEXT,
			mode INT,
			state TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS record (
			id TEXT NOT NULL PRIMARY KEY,
			name TEXT,
			start_time TEXT,
			end_time TEXT,
			mode INT,
			report_id TEXT
		);`,
	}

	sqliteName   = "sqlite.db"
	sqliteDriver = "sqlite3"
)

func Register() error {
	DB, err := sql.Open(sqliteDriver, sqliteName)
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
