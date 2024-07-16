package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"inspection-server/pkg/apis"
	"log"
)

func CreateReport(id, reportTime, data string, rating int) error {
	DB, err := sql.Open(sqliteDriver, sqliteName)
	if err != nil {
		log.Fatal(err)
	}
	defer DB.Close()

	tx, err := DB.Begin()
	if err != nil {
		log.Fatal(err)
	}
	stmt, err := tx.Prepare("INSERT INTO report(id, rating, report_time, data) VALUES(?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(id, rating, reportTime, data)
	if err != nil {
		log.Fatal(err)
	}
	tx.Commit()

	return nil
}

func GetReport(reportID string) (*apis.Report, error) {
	DB, err := sql.Open(sqliteDriver, sqliteName)
	if err != nil {
		return nil, err
	}
	defer DB.Close()

	row := DB.QueryRow("SELECT id, rating, report_time, data FROM report WHERE id = ? LIMIT 1", reportID)

	var id, reportTime, data string
	var rating int
	report := apis.NewReport()
	err = row.Scan(&id, &rating, &reportTime, &data)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("没有找到匹配的数据")
		} else {
			return nil, err
		}
	} else {
		dataMap := make(map[string]interface{})
		err := json.Unmarshal([]byte(data), &dataMap)
		if err != nil {
			return nil, err
		}
		for name, kubernetes := range dataMap {
			kubernetesBytes, err := json.Marshal(kubernetes)
			if err != nil {
				return nil, err
			}

			k := apis.NewKubernetes()
			err = json.Unmarshal(kubernetesBytes, k)
			if err != nil {
				return nil, err
			}

			report.Kubernetes[name] = k
		}

		report.ID = id
		report.Global.Rating = rating
		report.Global.ReportTime = reportTime
	}

	return report, nil
}

func DeleteReport(reportID string) error {
	DB, err := sql.Open(sqliteDriver, sqliteName)
	if err != nil {
		log.Fatal(err)
	}
	defer DB.Close()

	_, err = DB.Exec("DELETE FROM report WHERE id = ?", reportID)
	if err != nil {
		return err
	}

	return nil
}
