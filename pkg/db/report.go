package db

import (
	"encoding/json"
	"inspection-server/pkg/apis"
	"log"
)

// CreateReport inserts a new report into the database.
func CreateReport(report *apis.Report) error {
	data, err := json.Marshal(report.Kubernetes)
	if err != nil {
		log.Printf("Error marshaling Kubernetes data: %v", err)
		return err
	}

	tx, err := DB.Begin()
	if err != nil {
		log.Printf("Error beginning transaction: %v", err)
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO report(id, name, rating, report_time, data) VALUES(?, ?, ?, ?, ?)")
	if err != nil {
		log.Printf("Error preparing statement: %v", err)
		tx.Rollback() // Rollback transaction on error
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(report.ID, report.Global.Name, report.Global.Rating, report.Global.ReportTime, string(data))
	if err != nil {
		log.Printf("Error executing statement: %v", err)
		tx.Rollback() // Rollback transaction on error
		return err
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		return err
	}

	log.Printf("Report created successfully with ID: %s", report.ID)
	return nil
}

// GetReport retrieves a report from the database by ID.
func GetReport(reportID string) (*apis.Report, error) {
	row := DB.QueryRow("SELECT id, name, rating, report_time, data FROM report WHERE id = ? LIMIT 1", reportID)

	var id, name, rating, reportTime, data string
	report := apis.NewReport()
	err := row.Scan(&id, &name, &rating, &reportTime, &data)
	if err != nil {
		log.Printf("Error scanning row: %v", err)
		return nil, err
	}

	var dataKubernetes []*apis.Kubernetes
	err = json.Unmarshal([]byte(data), &dataKubernetes)
	if err != nil {
		log.Printf("Error unmarshaling Kubernetes data: %v", err)
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

	log.Printf("Report retrieved successfully with ID: %s", report.ID)
	return report, nil
}

// DeleteReport removes a report from the database by ID.
func DeleteReport(reportID string) error {
	result, err := DB.Exec("DELETE FROM report WHERE id = ?", reportID)
	if err != nil {
		log.Printf("Error executing delete statement: %v", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		return err
	}

	if rowsAffected == 0 {
		log.Printf("No report found to delete with ID: %s", reportID)
	} else {
		log.Printf("Report deleted successfully with ID: %s", reportID)
	}

	return nil
}
