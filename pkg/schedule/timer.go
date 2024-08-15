package schedule

import (
	"fmt"
	"inspection-server/pkg/apis"
	"time"

	"github.com/sirupsen/logrus"
)

func AddTimeTask(task *apis.Task) error {
	startTime, err := time.ParseInLocation(time.DateTime, task.StartTime, GetLoc())
	if err != nil {
		logrus.Errorf("Error parsing start time for task %s: %v", task.ID, err)
		return fmt.Errorf("error parsing start time for schedule %s: %v", task.ID, err)
	}

	duration := time.Until(startTime)
	logrus.Infof("Task %s will execute after %.2f minutes", task.ID, duration.Minutes())
	if duration <= 0 {
		task.StartTime = time.Now().Format(time.DateTime)
		go ExecuteTask(task)
		logrus.Warnf("Task %s is scheduled with a past time. Executing immediately.", task.ID)
		return nil
	}

	TaskMutex.Lock()
	defer TaskMutex.Unlock()

	timer := time.AfterFunc(duration, func() {
		logrus.Infof("Executing task %s", task.ID)
		go ExecuteTask(task)
	})

	TaskMap[task.ID] = &Schedule{
		Timer: timer,
	}

	logrus.Infof("Scheduled task %s to execute at %s", task.ID, task.StartTime)
	return nil
}

func RemoveTimetask(taskID string) error {
	TaskMutex.Lock()
	defer TaskMutex.Unlock()

	if s, exists := TaskMap[taskID]; exists {
		if !s.Timer.Stop() {
			logrus.Warnf("Timer for task %s has already executed or stopped", taskID)
		}
		delete(TaskMap, taskID)
		logrus.Infof("Deleted scheduled task %s", taskID)
	} else {
		logrus.Warnf("No scheduled task found with ID %s", taskID)
	}

	return nil
}
