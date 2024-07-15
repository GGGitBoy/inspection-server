package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"inspection-server/pkg/apis"
	"log"
)

func CreateReport(report *apis.Report) error {
	DB, err := GetDB()
	if err != nil {
		return err
	}
	defer DB.Close()

	data, err := json.Marshal(report.Kubernetes)
	if err != nil {
		return err
	}

	tx, err := DB.Begin()
	if err != nil {
		log.Fatal(err)
	}
	stmt, err := tx.Prepare("INSERT INTO report(id, name, rating, report_time, data) VALUES(?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(report.ID, report.Global.Name, report.Global.Rating, report.Global.ReportTime, string(data))
	if err != nil {
		log.Fatal(err)
	}
	tx.Commit()

	return nil
}

func GetReport(reportID string) (*apis.Report, error) {
	DB, err := GetDB()
	if err != nil {
		return nil, err
	}
	defer DB.Close()

	row := DB.QueryRow("SELECT id, name, rating, report_time, data FROM report WHERE id = ? LIMIT 1", reportID)

	var id, name, rating, reportTime, data string
	report := apis.NewReport()
	err = row.Scan(&id, &name, &rating, &reportTime, &data)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("没有找到匹配的数据")
		} else {
			return nil, err
		}
	} else {
		var dataKubernetes []*apis.Kubernetes
		err := json.Unmarshal([]byte(data), &dataKubernetes)
		if err != nil {
			return nil, err
		}

		report = &apis.Report{
			ID: id,
			Global: &apis.Global{
				Name:       name,
				Rating:     rating,
				ReportTime: reportTime,
			},
			Kubernetes: dataKubernetes,
		}
	}

	return report, nil
}

func DeleteReport(reportID string) error {
	DB, err := GetDB()
	if err != nil {
		return err
	}
	defer DB.Close()

	_, err = DB.Exec("DELETE FROM report WHERE id = ?", reportID)
	if err != nil {
		return err
	}

	return nil
}
