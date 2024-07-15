package schedule

import (
	"fmt"
	"inspection-server/pkg/apis"
	"time"
)

func AddTimeTask(task *apis.Task) error {
	startTime, err := time.ParseInLocation(time.DateTime, task.StartTime, GetLoc())
	if err != nil {
		return fmt.Errorf("Error parsing start time for schedule %s: %v\n", task.ID, err)
	}

	duration := time.Until(startTime)
	fmt.Printf("task %s will execute after %f minutes\n", task.ID, duration.Minutes())
	if duration <= 0 {
		task.StartTime = time.Now().Format(time.DateTime)
		go ExecuteTask(task)
		fmt.Printf("task %s 的 timer 是过去的时间\n", task.ID)
		return nil
	}

	TaskMutex.Lock()
	defer TaskMutex.Unlock()

	timer := time.AfterFunc(duration, func() {
		go ExecuteTask(task)
	})

	TaskMap[task.ID] = &Schedule{
		Timer: timer,
	}
	fmt.Printf("Scheduled schedule %s to execute at %s\n", task.ID, task.StartTime)
	return nil
}

func RemoveTimetask(taskID string) error {
	TaskMutex.Lock()
	defer TaskMutex.Unlock()

	if s, exists := TaskMap[taskID]; exists {
		s.Timer.Stop()
		delete(TaskMap, taskID)
		fmt.Printf("Deleted scheduled schedule %s\n", taskID)
	} else {
		fmt.Printf("No scheduled schedule found with ID %s\n", taskID)
	}

	return nil
}
