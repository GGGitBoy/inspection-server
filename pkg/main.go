package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

const (
	mysqlUser     = "your_mysql_username"
	mysqlPassword = "your_mysql_password"
	mysqlDB       = "your_database_name"
	mysqlHost     = "localhost"
	mysqlPort     = 3306
)

func main() {
	// 构建 MySQL 连接字符串
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", mysqlUser, mysqlPassword, mysqlHost, mysqlPort, mysqlDB)

	// 连接到 MySQL 数据库
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to MySQL: %v", err)
	}
	defer db.Close()

	// 创建表格的 SQL 语句
	createTableQueries := []string{
		`CREATE TABLE IF NOT EXISTS report (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            name TEXT,
            rating INT,
            report_time TEXT,
            data TEXT
        );`,
		`CREATE TABLE IF NOT EXISTS task (
            id VARCHAR(255) NOT NULL PRIMARY KEY,
            name TEXT,
            timer TEXT,
            cron TEXT,
            mode INT,
            state TEXT,
            template_id TEXT,
            notify_id TEXT
        );`,
		`CREATE TABLE IF NOT EXISTS task (
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

	// 执行每个创建表格的 SQL 语句
	for _, query := range createTableQueries {
		_, err := db.Exec(query)
		if err != nil {
			log.Fatalf("Failed to create table: %v", err)
		}
		fmt.Println("Table created successfully")
	}
}
