package db

import (
	"database/sql"
	"fmt"
	"inspection-server/pkg/apis"
	"log"
)

func GetTask(ID string) (*apis.Task, error) {
	DB, err := GetDB()
	if err != nil {
		return nil, err
	}
	defer DB.Close()

	row := DB.QueryRow("SELECT id, name, start_time, end_time, cron, state, rating, report_id, template_id, notify_id, task_id, mode, err_message FROM task WHERE id = ? LIMIT 1", ID)

	var id, name, startTime, endTime, cron, state, rating, reportID, templateID, notifyID, taskID, mode, errMessage string
	err = row.Scan(&id, &name, &startTime, &endTime, &cron, &state, &rating, &reportID, &templateID, &notifyID, &taskID, &mode, &errMessage)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("没有找到匹配的数据")
		} else {
			return nil, err
		}
	}

	return &apis.Task{
		ID:         id,
		Name:       name,
		StartTime:  startTime,
		EndTime:    endTime,
		Cron:       cron,
		State:      state,
		Rating:     rating,
		ReportID:   reportID,
		TemplateID: templateID,
		NotifyID:   notifyID,
		TaskID:     taskID,
		Mode:       mode,
		ErrMessage: errMessage,
	}, nil
}

func ListTask() ([]*apis.Task, error) {
	DB, err := GetDB()
	if err != nil {
		return nil, err
	}
	defer DB.Close()

	rows, err := DB.Query("SELECT id, name, start_time, end_time, cron, state, rating, report_id, template_id, notify_id, task_id, mode, err_message FROM task")
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	tasks := apis.NewTasks()
	for rows.Next() {
		var id, name, startTime, endTime, cron, state, rating, reportID, templateID, notifyID, taskID, mode, errMessage string
		err = rows.Scan(&id, &name, &startTime, &endTime, &cron, &state, &rating, &reportID, &templateID, &notifyID, &taskID, &mode, &errMessage)
		if err != nil {
			if err == sql.ErrNoRows {
				fmt.Println("没有找到匹配的数据")
			} else {
				return nil, err
			}
		}

		tasks = append(tasks, &apis.Task{
			ID:         id,
			Name:       name,
			StartTime:  startTime,
			EndTime:    endTime,
			Cron:       cron,
			State:      state,
			Rating:     rating,
			ReportID:   reportID,
			TemplateID: templateID,
			NotifyID:   notifyID,
			TaskID:     taskID,
			Mode:       mode,
			ErrMessage: errMessage,
		})
	}

	return tasks, nil
}

func CreateTask(task *apis.Task) error {
	DB, err := GetDB()
	if err != nil {
		return err
	}
	defer DB.Close()

	tx, err := DB.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := tx.Prepare("INSERT INTO task(id, name, start_time, end_time, cron, state, rating, report_id, template_id, notify_id, task_id, mode, err_message) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(task.ID, task.Name, task.StartTime, task.EndTime, task.Cron, task.State, task.Rating, task.ReportID, task.TemplateID, task.NotifyID, task.TaskID, task.Mode, task.ErrMessage)
	if err != nil {
		log.Fatal(err)
	}
	tx.Commit()

	return nil
}

func Updatetask(task *apis.Task) error {
	DB, err := GetDB()
	if err != nil {
		return err
	}
	defer DB.Close()

	_, err = DB.Exec("UPDATE task SET name = ?, start_time = ?, end_time = ?, cron = ?, state = ?, rating = ?, report_id = ?, template_id = ?, notify_id = ?, task_id = ?, mode = ?, err_message = ?  WHERE id = ?", task.Name, task.StartTime, task.EndTime, task.Cron, task.State, task.Rating, task.ReportID, task.TemplateID, task.NotifyID, task.TaskID, task.Mode, task.ID, task.ErrMessage)
	if err != nil {
		return err
	}

	return nil
}

func Deletetask(ID string) error {
	DB, err := GetDB()
	if err != nil {
		return err
	}
	defer DB.Close()

	_, err = DB.Exec("DELETE FROM task WHERE id = ?", ID)
	if err != nil {
		return err
	}

	return nil
}
