package db

import (
	"inspection-server/pkg/apis"
	"log"
)

// GetTask retrieves a task from the database by ID.
func GetTask(ID string) (*apis.Task, error) {
	DB, err := GetDB()
	if err != nil {
		log.Printf("Error getting database connection: %v", err)
		return nil, err
	}
	defer func() {
		if err := DB.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	row := DB.QueryRow("SELECT id, name, start_time, end_time, cron, state, rating, report_id, template_id, notify_id, task_id, mode, err_message FROM task WHERE id = ? LIMIT 1", ID)

	task := apis.NewTask()
	err = row.Scan(&task.ID, &task.Name, &task.StartTime, &task.EndTime, &task.Cron, &task.State, &task.Rating, &task.ReportID, &task.TemplateID, &task.NotifyID, &task.TaskID, &task.Mode, &task.ErrMessage)
	if err != nil {
		log.Printf("Error scanning task row: %v", err)
		return nil, err
	}

	log.Printf("Task retrieved successfully with ID: %s", task.ID)
	return task, nil
}

// ListTask retrieves all tasks from the database.
func ListTask() ([]*apis.Task, error) {
	DB, err := GetDB()
	if err != nil {
		log.Printf("Error getting database connection: %v", err)
		return nil, err
	}
	defer func() {
		if err := DB.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	rows, err := DB.Query("SELECT id, name, start_time, end_time, cron, state, rating, report_id, template_id, notify_id, task_id, mode, err_message FROM task")
	if err != nil {
		log.Printf("Error executing query: %v", err)
		return nil, err
	}
	defer rows.Close()

	tasks := apis.NewTasks()
	for rows.Next() {
		var task apis.Task
		err = rows.Scan(&task.ID, &task.Name, &task.StartTime, &task.EndTime, &task.Cron, &task.State, &task.Rating, &task.ReportID, &task.TemplateID, &task.NotifyID, &task.TaskID, &task.Mode, &task.ErrMessage)
		if err != nil {
			log.Printf("Error scanning task row: %v", err)
			return nil, err
		}

		tasks = append(tasks, &task)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating over task rows: %v", err)
		return nil, err
	}

	log.Printf("Tasks retrieved successfully, total count: %d", len(tasks))
	return tasks, nil
}

// CreateTask inserts a new task into the database.
func CreateTask(task *apis.Task) error {
	DB, err := GetDB()
	if err != nil {
		log.Printf("Error getting database connection: %v", err)
		return err
	}
	defer func() {
		if err := DB.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	tx, err := DB.Begin()
	if err != nil {
		log.Printf("Error beginning transaction: %v", err)
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO task(id, name, start_time, end_time, cron, state, rating, report_id, template_id, notify_id, task_id, mode, err_message) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Printf("Error preparing statement: %v", err)
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(task.ID, task.Name, task.StartTime, task.EndTime, task.Cron, task.State, task.Rating, task.ReportID, task.TemplateID, task.NotifyID, task.TaskID, task.Mode, task.ErrMessage)
	if err != nil {
		log.Printf("Error executing statement: %v", err)
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		return err
	}

	log.Printf("Task created successfully with ID: %s", task.ID)
	return nil
}

// UpdateTask updates an existing task in the database.
func UpdateTask(task *apis.Task) error {
	DB, err := GetDB()
	if err != nil {
		log.Printf("Error getting database connection: %v", err)
		return err
	}
	defer func() {
		if err := DB.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	_, err = DB.Exec("UPDATE task SET name = ?, start_time = ?, end_time = ?, cron = ?, state = ?, rating = ?, report_id = ?, template_id = ?, notify_id = ?, task_id = ?, mode = ?, err_message = ? WHERE id = ?", task.Name, task.StartTime, task.EndTime, task.Cron, task.State, task.Rating, task.ReportID, task.TemplateID, task.NotifyID, task.TaskID, task.Mode, task.ErrMessage, task.ID)
	if err != nil {
		log.Printf("Error updating task: %v", err)
		return err
	}

	log.Printf("Task updated successfully with ID: %s", task.ID)
	return nil
}

// DeleteTask removes a task from the database by ID.
func DeleteTask(ID string) error {
	DB, err := GetDB()
	if err != nil {
		log.Printf("Error getting database connection: %v", err)
		return err
	}
	defer func() {
		if err := DB.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	result, err := DB.Exec("DELETE FROM task WHERE id = ?", ID)
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
		log.Printf("No task found to delete with ID: %s", ID)
	} else {
		log.Printf("Task deleted successfully with ID: %s", ID)
	}

	return nil
}
