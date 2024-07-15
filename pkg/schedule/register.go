package schedule

import (
	"fmt"
	"github.com/robfig/cron/v3"
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

	plans, err := db.ListPlan()
	if err != nil {
		return err
	}

	for _, plan := range plans {
		err = AddSchedule(plan)
		if err != nil {
			return err
		}
	}

	return nil
}

func AddSchedule(plan *apis.Plan) error {
	var err error
	if plan.Mode == 0 {
		go ExecuteTask(plan)
	} else if plan.Mode == 1 {
		err = AddTimePlan(plan)
	} else if plan.Mode == 2 {
		err = AddCornPlan(plan)
	}

	for id, _ := range TaskMap {
		fmt.Println(id)
	}

	return err
}

func RemoveSchedule(plan *apis.Plan) error {
	var err error
	if plan.Mode == 1 {
		err = RemoveTimePlan(plan.ID)
	} else if plan.Mode == 2 {
		err = RemoveCornPlan(plan.ID)
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

func ExecuteTask(plan *apis.Plan) {
	fmt.Printf("Executing task %s: %s\n", plan.ID, plan.Name)
	err := core.Inspection(plan)
	if err != nil {
		fmt.Println(err)
		err = TaskError(plan)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		if plan.Mode == 0 {
			err = db.DeletePlan(plan.ID)
			if err != nil {
				log.Fatal(err)
			}
		} else if plan.Mode == 1 {
			err = RemoveSchedule(plan)
			if err != nil {
				log.Fatal(err)
			}

			err = db.DeletePlan(plan.ID)
			if err != nil {
				log.Fatal(err)
			}
		} else if plan.Mode == 2 {
			plan.State = "计划中"
			err = db.UpdatePlan(plan)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func TaskError(plan *apis.Plan) error {
	plan.State = "巡检失败"
	err := db.UpdatePlan(plan)
	if err != nil {
		return err
	}

	return nil
}
