package schedule

import (
	"fmt"
	"inspection-server/pkg/apis"
)

func AddCornPlan(plan *apis.Plan) error {
	TaskMutex.Lock()
	defer TaskMutex.Unlock()

	entryID, err := CronClient.AddFunc(plan.Cron, func() {
		go ExecuteTask(plan)
		fmt.Printf("Executing schedule: %+v\n", plan)
	})
	if err != nil {
		return fmt.Errorf("Error adding cron job: %v\n", err)
	}
	TaskMap[plan.ID] = &Schedule{
		Cron: entryID,
	}

	fmt.Printf("Scheduled schedule %s to execute at %s\n", plan.ID, plan.Cron)

	return nil
}

func RemoveCornPlan(planID string) error {
	TaskMutex.Lock()
	defer TaskMutex.Unlock()

	if s, exists := TaskMap[planID]; exists {
		CronClient.Remove(s.Cron)
		delete(TaskMap, planID)
		fmt.Printf("Deleted scheduled schedule %s\n", planID)
	} else {
		fmt.Printf("No scheduled schedule found with ID %s\n", planID)
	}

	fmt.Printf("Removed schedule: %+v\n", planID)

	return nil
}
