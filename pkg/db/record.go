package db

import (
	"database/sql"
	"fmt"
	"inspection-server/pkg/apis"
	"log"
)

func GetRecord(recordID string) (*apis.Record, error) {
	DB, err := GetDB()
	if err != nil {
		return nil, err
	}
	defer DB.Close()

	row := DB.QueryRow("SELECT id, name, start_time, end_time, mode, report_id FROM record WHERE id = ? LIMIT 1", recordID)

	var id, name, startTime, endTime, reportID string
	var mode int
	err = row.Scan(&id, &name, &startTime, &endTime, &mode, &reportID)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("没有找到匹配的数据")
		} else {
			return nil, err
		}
	}

	return &apis.Record{
		ID:        id,
		Name:      name,
		StartTime: startTime,
		EndTime:   endTime,
		Mode:      mode,
		ReportID:  reportID,
	}, nil
}

func ListRecord() ([]*apis.Record, error) {
	DB, err := GetDB()
	if err != nil {
		return nil, err
	}
	defer DB.Close()

	rows, err := DB.Query("SELECT id, name, start_time, end_time, mode, report_id FROM record")
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	records := apis.NewRecords()
	for rows.Next() {
		var id, name, startTime, endTime, reportID string
		var mode int
		err = rows.Scan(&id, &name, &startTime, &endTime, &mode, &reportID)
		if err != nil {
			if err == sql.ErrNoRows {
				fmt.Println("没有找到匹配的数据")
			} else {
				return nil, err
			}
		}

		records = append(records, &apis.Record{
			ID:        id,
			Name:      name,
			StartTime: startTime,
			EndTime:   endTime,
			Mode:      mode,
			ReportID:  reportID,
		})
	}

	return records, nil
}

func CreateRecord(record *apis.Record) error {
	DB, err := GetDB()
	if err != nil {
		return err
	}
	defer DB.Close()

	tx, err := DB.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := tx.Prepare("INSERT INTO record(id, name, start_time, end_time, mode, report_id) VALUES(?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(record.ID, record.Name, record.StartTime, record.EndTime, record.Mode, record.ReportID)
	if err != nil {
		log.Fatal(err)
	}
	tx.Commit()

	return nil
}

func DeleteRecord(recordID string) error {
	DB, err := GetDB()
	if err != nil {
		return err
	}
	defer DB.Close()

	_, err = DB.Exec("DELETE FROM record WHERE id = ?", recordID)
	if err != nil {
		return err
	}

	return nil
}
