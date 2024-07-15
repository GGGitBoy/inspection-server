package schedule

import (
	"fmt"
	"inspection-server/pkg/apis"
	"time"
)

func AddTimePlan(plan *apis.Plan) error {
	startTime, err := time.ParseInLocation(time.DateTime, plan.Timer, GetLoc())
	if err != nil {
		return fmt.Errorf("Error parsing start time for schedule %s: %v\n", plan.ID, err)
	}

	duration := time.Until(startTime)
	fmt.Printf("plan %s will execute after %f minutes\n", plan.ID, duration.Minutes())
	if duration <= 0 {
		return fmt.Errorf("plan %s 的 timer 是过去的时间\n", plan.ID)
	}

	TaskMutex.Lock()
	defer TaskMutex.Unlock()

	timer := time.AfterFunc(duration, func() {
		go ExecuteTask(plan)
	})

	TaskMap[plan.ID] = &Schedule{
		Timer: timer,
	}
	fmt.Printf("Scheduled schedule %s to execute at %s\n", plan.ID, plan.Timer)
	return nil
}

func RemoveTimePlan(planID string) error {
	TaskMutex.Lock()
	defer TaskMutex.Unlock()

	if s, exists := TaskMap[planID]; exists {
		s.Timer.Stop()
		delete(TaskMap, planID)
		fmt.Printf("Deleted scheduled schedule %s\n", planID)
	} else {
		fmt.Printf("No scheduled schedule found with ID %s\n", planID)
	}

	return nil
}
