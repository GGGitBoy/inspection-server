package core

import (
	"inspection-server/pkg/apis"
	"inspection-server/pkg/common"
	"inspection-server/pkg/db"
	"inspection-server/pkg/send"
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

	template, err := db.GetTemplate(plan.TemplateID)
	if err != nil {
		return err
	}

	clients := apis.NewClients()
	err = common.GenerateKubeconfig(clients)
	if err != nil {
		return err
	}

	report := apis.NewReport()
	kubernetes := apis.NewKubernetes()
	for clusterID, client := range clients {
		for _, k := range template.KubernetesConfig {
			if k.ClusterID == clusterID && k.Enable {
				clusterCore := apis.NewClusterCore()
				clusterNode := apis.NewClusterNode()
				clusterResource := apis.NewClusterResource()

				coreInspections := apis.NewInspections()
				nodeInspections := apis.NewInspections()
				resourceInspections := apis.NewInspections()

				CoreWorkloadArray, ResourceWorkloadArray, coreInspectionArray, resourceInspectionArray, err := GetWorkloads(client, k.WorkloadConfig)
				if err != nil {
					return err
				}
				coreInspections = append(coreInspections, coreInspectionArray...)
				resourceInspections = append(resourceInspections, resourceInspectionArray...)

				CoreNodeArray, NodeNodeArray, coreInspectionArray, nodeInspectionArray, err := GetNodes(client, k.NodeConfig)
				if err != nil {
					return err
				}
				coreInspections = append(coreInspections, coreInspectionArray...)
				nodeInspections = append(nodeInspections, nodeInspectionArray...)

				ResourceNamespaceArray, resourceInspectionArray, err := GetNamespaces(client)
				if err != nil {
					return err
				}
				resourceInspections = append(resourceInspections, resourceInspectionArray...)

				//ResourcePersistentVolumeClaimArray, resourceInspectionArray, err := GetPersistentVolumeClaims(name, client)
				//if err != nil {
				//	return err
				//}
				//resourceInspections = append(resourceInspections, resourceInspectionArray...)

				ResourceServiceArray, resourceInspectionArray, err := GetServices(client)
				if err != nil {
					return err
				}
				resourceInspections = append(resourceInspections, resourceInspectionArray...)

				ResourceIngressArray, resourceInspectionArray, err := GetIngress(client)
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

				kubernetes = append(kubernetes, &apis.Kubernetes{
					ClusterID:       k.ClusterID,
					ClusterName:     k.ClusterName,
					ClusterCore:     clusterCore,
					ClusterNode:     clusterNode,
					ClusterResource: clusterResource,
				})
			}
		}
	}

	report = &apis.Report{
		ID: common.GetUUID(),
		Global: &apis.Global{
			Name:       record.Name,
			Rating:     0,
			ReportTime: time.Now().Format(time.DateTime),
		},
		Kubernetes: kubernetes,
	}
	err = db.CreateReport(report)
	if err != nil {
		return err
	}

	record.EndTime = time.Now().Format(time.DateTime)
	record.ReportID = report.ID
	err = db.CreateRecord(record)
	if err != nil {
		return err
	}

	if plan.NotifyID != "" {
		notify, err := db.GetNotify(plan.NotifyID)
		if err != nil {
			return err
		}

		err = send.Notify(notify.AppID, notify.AppSecret)
		if err != nil {
			return err
		}
	}

	return nil
}
