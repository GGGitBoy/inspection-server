package core

import (
	"encoding/json"
	"inspection-server/pkg/apis"
	"inspection-server/pkg/common"
	"inspection-server/pkg/db"
	"time"
)

func Inspection(plan *apis.Plan) error {
	record := apis.NewRecord()
	record.ID = common.GetUUID()
	record.Name = plan.Name
	record.Mode = plan.Mode
	record.StartTime = time.Now().Format(time.DateTime)

	plan.State = "巡检中"
	err := db.UpdatePlan(plan)
	if err != nil {
		return err
	}

	clients := apis.NewClients()
	err = common.GenerateKubeconfig(clients)
	if err != nil {
		return err
	}

	report := apis.NewReport()
	for name, client := range clients {
		report.Kubernetes[name] = apis.NewKubernetes()

		workloadArray, err := GetWorkloads(name, client)
		if err != nil {
			return err
		}
		report.Kubernetes[name].Workloads = workloadArray

		nodeArray, err := GetNodes(client)
		if err != nil {
			return err
		}
		report.Kubernetes[name].Nodes = nodeArray

		err = GetGlobal(report)

		if err != nil {
			return err
		}
	}

	dataWarnings, err := json.Marshal(report.Global.Warnings)
	if err != nil {
		return err
	}

	data, err := json.Marshal(report.Kubernetes)
	if err != nil {
		return err
	}

	reportID := common.GetUUID()
	err = db.CreateReport(reportID, report.Global.ReportTime, string(dataWarnings), string(data), report.Global.Rating)
	if err != nil {
		return err
	}

	record.EndTime = time.Now().Format(time.DateTime)
	record.ReportID = reportID
	err = db.CreateRecord(record)
	if err != nil {
		return err
	}

	time.Sleep(5 * time.Second)

	return nil
}
