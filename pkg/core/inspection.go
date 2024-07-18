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
		clusterCore := apis.NewClusterCore()
		clusterNode := apis.NewClusterNode()
		clusterResource := apis.NewClusterResource()

		coreInspections := apis.NewInspections()
		nodeInspections := apis.NewInspections()
		resourceInspections := apis.NewInspections()

		CoreWorkloadArray, ResourceWorkloadArray, coreInspectionArray, resourceInspectionArray, err := GetWorkloads(name, client)
		if err != nil {
			return err
		}
		coreInspections = append(coreInspections, coreInspectionArray...)
		resourceInspections = append(resourceInspections, resourceInspectionArray...)

		CoreNodeArray, NodeNodeArray, coreInspectionArray, nodeInspectionArray, err := GetNodes(name, client)
		if err != nil {
			return err
		}
		coreInspections = append(coreInspections, coreInspectionArray...)
		nodeInspections = append(nodeInspections, nodeInspectionArray...)

		ResourceNamespaceArray, resourceInspectionArray, err := GetNamespaces(name, client)
		if err != nil {
			return err
		}
		resourceInspections = append(resourceInspections, resourceInspectionArray...)

		//ResourcePersistentVolumeClaimArray, resourceInspectionArray, err := GetPersistentVolumeClaims(name, client)
		//if err != nil {
		//	return err
		//}
		//resourceInspections = append(resourceInspections, resourceInspectionArray...)

		ResourceServiceArray, resourceInspectionArray, err := GetServices(name, client)
		if err != nil {
			return err
		}
		resourceInspections = append(resourceInspections, resourceInspectionArray...)

		ResourceIngressArray, resourceInspectionArray, err := GetIngress(name, client)
		if err != nil {
			return err
		}
		resourceInspections = append(resourceInspections, resourceInspectionArray...)

		clusterCore.Workloads = CoreWorkloadArray
		clusterCore.Nodes = CoreNodeArray
		clusterCore.Inspections = coreInspections

		clusterNode.Nodes = NodeNodeArray
		clusterNode.Inspections = nodeInspections

		clusterResource.Workloads = ResourceWorkloadArray
		clusterResource.Namespace = ResourceNamespaceArray
		//clusterResource.PersistentVolumeClaim = ResourcePersistentVolumeClaimArray
		clusterResource.Service = ResourceServiceArray
		clusterResource.Ingress = ResourceIngressArray
		clusterResource.Inspections = resourceInspections

		report.Kubernetes[name].ClusterCore = clusterCore
		report.Kubernetes[name].ClusterNode = clusterNode
		report.Kubernetes[name].ClusterResource = clusterResource
	}

	data, err := json.Marshal(report.Kubernetes)
	if err != nil {
		return err
	}

	reportID := common.GetUUID()
	rating := 0
	reportTime := time.Now().Format(time.DateTime)
	err = db.CreateReport(reportID, record.Name, reportTime, string(data), rating)
	if err != nil {
		return err
	}

	record.EndTime = time.Now().Format(time.DateTime)
	record.ReportID = reportID
	err = db.CreateRecord(record)
	if err != nil {
		return err
	}

	return nil
}
