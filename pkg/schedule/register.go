package schedule

import (
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"inspection-server/pkg/apis"
	"inspection-server/pkg/core"
	"inspection-server/pkg/db"
	"sync"
	"time"
)

var (
	TaskMap    = make(map[string]*Schedule)
	TaskMutex  sync.Mutex
	CronClient = cron.New()
)

type Schedule struct {
	Timer *time.Timer  `json:"timer"`
	Cron  cron.EntryID `json:"cron"`
}

func Register() error {
	CronClient.Start()

	tasks, err := db.ListTask()
	if err != nil {
		logrus.Errorf("Failed to list tasks from DB: %v", err)
		return err
	}

	for _, task := range tasks {
		err = AddSchedule(task)
		if err != nil {
			logrus.Errorf("Failed to add schedule for task %s: %v", task.ID, err)
			return err
		}
	}

	logrus.Infof("All tasks registered successfully.")
	return nil
}

func AddSchedule(task *apis.Task) error {
	var err error
	if task.Mode == "计划任务" {
		err = AddTimeTask(task)
	} else if task.Mode == "周期任务" {
		err = AddCornTask(task)
	}

	if err != nil {
		logrus.Errorf("Failed to add schedule for task %s: %v", task.ID, err)
	} else {
		logrus.Infof("Schedule added for task %s", task.ID)
	}

	logrus.Debug("Current TaskMap keys:")
	for id := range TaskMap {
		logrus.Debug(id)
	}

	return err
}

func RemoveSchedule(task *apis.Task) error {
	var err error
	if task.Mode == "计划任务" {
		err = RemoveTimetask(task.ID)
	} else if task.Mode == "周期任务" {
		err = RemoveCorntask(task.ID)
	}

	if err != nil {
		logrus.Errorf("Failed to remove schedule for task %s: %v", task.ID, err)
	} else {
		logrus.Infof("Schedule removed for task %s", task.ID)
	}

	logrus.Debug("Current TaskMap keys:")
	for id := range TaskMap {
		logrus.Debug(id)
	}

	return err
}

func GetLoc() *time.Location {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		logrus.Fatalf("Failed to load location: %v", err)
	}

	return loc
}

func ExecuteTask(task *apis.Task) {
	logrus.Infof("Executing task %s: %s", task.ID, task.Name)

	err, errMessage := core.Inspection(task)
	if err != nil {
		logrus.Errorf("Inspection failed for task %s: %v", task.ID, err)
		task.State = "巡检失败"
		task.ErrMessage = errMessage.String()
		updateErr := db.UpdateTask(task)
		if updateErr != nil {
			logrus.Errorf("Failed to update task %s with error message: %v", task.ID, updateErr)
		}
	} else {
		removeErr := RemoveSchedule(task)
		if removeErr != nil {
			logrus.Errorf("Failed to remove schedule for task %s: %v", task.ID, removeErr)
		}
	}
}
