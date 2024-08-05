package schedule

import (
	"fmt"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"inspection-server/pkg/apis"
	"inspection-server/pkg/core"
	"inspection-server/pkg/db"
	"log"
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
		return err
	}

	for _, task := range tasks {
		err = AddSchedule(task)
		if err != nil {
			return err
		}
	}

	return nil
}

func AddSchedule(task *apis.Task) error {
	var err error
	if task.Mode == "计划任务" {
		err = AddTimeTask(task)
	} else if task.Mode == "周期任务" {
		err = AddCornTask(task)
	}

	for id, _ := range TaskMap {
		fmt.Println(id)
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

	for id, _ := range TaskMap {
		fmt.Println(id)
	}

	return err
}

func GetLoc() *time.Location {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		log.Fatal(err)
	}

	return loc
}

func ExecuteTask(task *apis.Task) {
	logrus.Infof("Executing task %s: %s\n", task.ID, task.Name)

	err, errMessage := core.Inspection(task)
	if err != nil {
		errMessage.WriteString(fmt.Sprintf("巡检过程中报错: %v\n", err))
		task.State = "巡检失败"
		task.ErrMessage = errMessage.String()
		err = db.UpdateTask(task)
		if err != nil {
			logrus.Errorf("update task ErrMessage error: %v\n", err)
		}
	} else {
		err = RemoveSchedule(task)
		if err != nil {
			logrus.Errorf("remove task schedule error: %v\n", err)
		}
	}
}
