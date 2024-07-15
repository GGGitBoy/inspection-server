package schedule

import (
	"fmt"
	"inspection-server/pkg/apis"
	"inspection-server/pkg/common"
	"inspection-server/pkg/db"
	"time"
)

func AddCornTask(task *apis.Task) error {
	TaskMutex.Lock()
	defer TaskMutex.Unlock()

	entryID, err := CronClient.AddFunc(task.Cron, func() {
		now := time.Now().Format(time.DateTime)
		newTask := &apis.Task{
			ID:         common.GetUUID(),
			Name:       fmt.Sprintf("%s(%s)", task.Name, now),
			StartTime:  now,
			EndTime:    "",
			Cron:       task.Cron,
			State:      "巡检中",
			Rating:     "",
			ReportID:   "",
			TemplateID: task.TemplateID,
			NotifyID:   task.NotifyID,
			TaskID:     task.ID,
			Mode:       task.Mode,
		}

		err := db.CreateTask(newTask)
		if err != nil {
			fmt.Errorf("Scheduled schedule %s to execute at %s\n", newTask.ID, newTask.Cron)
		}

		go ExecuteTask(newTask)
		fmt.Printf("Executing schedule: %+v\n", newTask)
	})
	if err != nil {
		return fmt.Errorf("Error adding cron job: %v\n", err)
	}
	TaskMap[task.ID] = &Schedule{
		Cron: entryID,
	}

	fmt.Printf("Scheduled schedule %s to execute at %s\n", task.ID, task.Cron)

	return nil
}

func RemoveCorntask(taskID string) error {
	TaskMutex.Lock()
	defer TaskMutex.Unlock()

	if s, exists := TaskMap[taskID]; exists {
		CronClient.Remove(s.Cron)
		delete(TaskMap, taskID)
		fmt.Printf("Deleted scheduled schedule %s\n", taskID)
	} else {
		fmt.Printf("No scheduled schedule found with ID %s\n", taskID)
	}

	fmt.Printf("Removed schedule: %+v\n", taskID)

	return nil
}
