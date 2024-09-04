package db

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"inspection-server/pkg/common"
	"log"
	"time"
)

var (
	sqlTables = []string{
		`CREATE TABLE IF NOT EXISTS report (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            name TEXT,
            rating TEXT, 
            report_time TEXT,
            data LONGTEXT
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
            data LONGTEXT
        );`,
		`CREATE TABLE IF NOT EXISTS notify (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            name TEXT,
            app_id TEXT,
            app_secret TEXT,
            webhook_url TEXT,
			secret TEXT
        );`,
	}

	DB *sql.DB
)

// Register creates tables if they do not exist
func Register() error {
	DB, err := GetDB()
	if err != nil {
		log.Printf("Error getting database connection: %v", err)
		return err
	}

	for _, table := range sqlTables {
		_, err = DB.Exec(table)
		if err != nil {
			log.Printf("Error executing table creation SQL: %v\nSQL: %s", err, table)
			return err
		}
	}

	return nil
}

// GetDB returns a database connection based on configuration
func GetDB() (*sql.DB, error) {
	var err error

	if common.MySQL == "true" {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", common.MySQLUser, common.MySQLPassword, common.MySQLHost, common.MySQLPort, common.MySQLDB)
		DB, err = sql.Open("mysql", dsn)
		if err != nil {
			log.Printf("Error opening MySQL connection: %v, DSN: %s", err, dsn)
			return nil, err
		}
		log.Println("Connected to MySQL database successfully")
	} else {
		DB, err = sql.Open("sqlite3", common.SQLiteName)
		if err != nil {
			log.Printf("Error opening SQLite connection: %v, DB Name: %s", err, common.SQLiteName)
			return nil, err
		}
		log.Println("Connected to SQLite database successfully")
	}

	DB.SetMaxOpenConns(25)                 // 最大打开的连接数
	DB.SetMaxIdleConns(25)                 // 最大空闲连接数
	DB.SetConnMaxLifetime(5 * time.Minute) // 连接最长存活时间

	// Test database connection
	if err := DB.Ping(); err != nil {
		log.Printf("Error pinging database: %v", err)
		return nil, err
	}
	log.Println("Database connection successful")

	return DB, nil
}
