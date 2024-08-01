package core

import (
	"fmt"
	"inspection-server/pkg/apis"
	"inspection-server/pkg/common"
	"inspection-server/pkg/db"
	pdfPrint "inspection-server/pkg/print"
	"inspection-server/pkg/send"
	"time"
)

func Inspection(plan *apis.Plan) error {
	record := apis.NewRecord()
	record.ID = common.GetUUID()
	record.Name = plan.Name
	record.Mode = plan.Mode
	record.State = "巡检中"
	record.TemplateID = plan.TemplateID
	record.NotifyID = plan.NotifyID
	record.PlanID = plan.ID
	record.StartTime = time.Now().Format(time.DateTime)
	err := db.CreateRecord(record)
	if err != nil {
		return err
	}

	plan.State = "巡检中"
	err = db.UpdatePlan(plan)
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

	allGrafanaInspections, err := GetAllGrafanaInspections()
	if err != nil {
		return err
	}

	var sendMessageDetail []string
	for clusterID, client := range clients {
		for _, k := range template.KubernetesConfig {
			if k.ClusterID == clusterID && k.Enable {
				clusterCore := apis.NewClusterCore()
				clusterNode := apis.NewClusterNode()
				clusterResource := apis.NewClusterResource()

				coreInspections := apis.NewInspections()
				nodeInspections := apis.NewInspections()
				resourceInspections := apis.NewInspections()

				NodeNodeArray, nodeInspectionArray, err := GetNodes(client, k.ClusterNodeConfig.NodeConfig)
				if err != nil {
					return err
				}
				nodeInspections = append(nodeInspections, nodeInspectionArray...)

				ResourceWorkloadArray, resourceInspectionArray, err := GetWorkloads(client, k.ClusterResourceConfig.WorkloadConfig)
				if err != nil {
					return err
				}
				resourceInspections = append(resourceInspections, resourceInspectionArray...)

				if k.ClusterResourceConfig.NamespaceConfig.Enable {
					ResourceNamespaceArray, resourceInspectionArray, err := GetNamespaces(client)
					if err != nil {
						return err
					}

					clusterResource.Namespace = ResourceNamespaceArray
					resourceInspections = append(resourceInspections, resourceInspectionArray...)
				}

				if k.ClusterResourceConfig.ServiceConfig.Enable {
					ResourceServiceArray, resourceInspectionArray, err := GetServices(client)
					if err != nil {
						return err
					}

					clusterResource.Service = ResourceServiceArray
					resourceInspections = append(resourceInspections, resourceInspectionArray...)
				}

				if k.ClusterResourceConfig.ServiceConfig.Enable {
					ResourceIngressArray, resourceInspectionArray, err := GetIngress(client)
					if err != nil {
						return err
					}

					clusterResource.Ingress = ResourceIngressArray
					resourceInspections = append(resourceInspections, resourceInspectionArray...)
				}

				clusterNode.Nodes = NodeNodeArray
				clusterResource.Workloads = ResourceWorkloadArray

				if allGrafanaInspections[k.ClusterName] != nil {
					if len(allGrafanaInspections[k.ClusterName].ClusterCoreInspection) > 0 {
						coreInspections = append(coreInspections, allGrafanaInspections[k.ClusterName].ClusterCoreInspection...)
					}

					if len(allGrafanaInspections[k.ClusterName].ClusterNodeInspection) > 0 {
						nodeInspections = append(nodeInspections, allGrafanaInspections[k.ClusterName].ClusterNodeInspection...)
					}

					if len(allGrafanaInspections[k.ClusterName].ClusterResourceInspection) > 0 {
						resourceInspections = append(resourceInspections, allGrafanaInspections[k.ClusterName].ClusterResourceInspection...)
					}
				}

				clusterCore.Inspections = coreInspections
				clusterNode.Inspections = nodeInspections
				clusterResource.Inspections = resourceInspections

				for _, c := range coreInspections {
					if c.Level <= 2 {
						sendMessageDetail = append(sendMessageDetail, fmt.Sprintf("集群 %s 警告: %s", k.ClusterName, c.Title))
					}
				}
				for _, n := range nodeInspections {
					if n.Level <= 2 {
						sendMessageDetail = append(sendMessageDetail, fmt.Sprintf("集群 %s 警告: %s", k.ClusterName, n.Title))
					}
				}
				for _, r := range resourceInspections {
					if r.Level <= 2 {
						sendMessageDetail = append(sendMessageDetail, fmt.Sprintf("集群 %s 警告: %s", k.ClusterName, r.Title))
					}
				}

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
			Rating:     "优",
			ReportTime: time.Now().Format(time.DateTime),
		},
		Kubernetes: kubernetes,
	}
	err = db.CreateReport(report)
	if err != nil {
		return err
	}

	sendMessage := "该巡检报告的健康等级为: " + report.Global.Rating + "\n"
	for _, s := range sendMessageDetail {
		sendMessage = sendMessage + s + "\n"
	}

	if plan.NotifyID != "" {
		notify, err := db.GetNotify(plan.NotifyID)
		if err != nil {
			return err
		}

		p := pdfPrint.NewPrint()
		p.URL = "http://127.0.0.1/#/inspection-record/result-pdf-view/" + report.ID
		p.ReportTime = report.Global.ReportTime
		err = pdfPrint.FullScreenshot(p)
		if err != nil {
			return err
		}

		err = send.Notify(notify.AppID, notify.AppSecret, common.GetReportFileName(p.ReportTime), common.PrintPDFPath+common.GetReportFileName(p.ReportTime), "该测试报告的健康等级为: 优")
		if err != nil {
			return err
		}
	}

	record.EndTime = time.Now().Format(time.DateTime)
	record.Rating = report.Global.Rating
	record.ReportID = report.ID
	record.State = "巡检完成"
	err = db.UpdateRecord(record)
	if err != nil {
		return err
	}

	return nil
}
