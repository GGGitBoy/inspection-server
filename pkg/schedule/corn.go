package schedule

import (
	"fmt"
	"inspection-server/pkg/apis"
	"inspection-server/pkg/common"
	"inspection-server/pkg/db"
	"log"
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
			log.Printf("Failed to create task in DB for scheduled schedule %s: %v", newTask.ID, err)
			return
		}

		log.Printf("Executing task: %s", newTask.ID)
		go ExecuteTask(newTask)
		log.Printf("Task %s is executing", newTask.ID)
	})
	if err != nil {
		return fmt.Errorf("Error adding cron job: %v", err)
	}

	TaskMap[task.ID] = &Schedule{
		Cron: entryID,
	}

	log.Printf("Scheduled task %s to execute at %s", task.ID, task.Cron)

	return nil
}

func RemoveCorntask(taskID string) error {
	TaskMutex.Lock()
	defer TaskMutex.Unlock()

	if s, exists := TaskMap[taskID]; exists {
		CronClient.Remove(s.Cron)
		delete(TaskMap, taskID)
		log.Printf("Deleted scheduled task with ID %s", taskID)
	} else {
		log.Printf("No scheduled task found with ID %s", taskID)
	}

	log.Printf("Removed task with ID: %s", taskID)

	return nil
}
