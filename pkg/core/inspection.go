package core

import (
	"fmt"
	"inspection-server/pkg/apis"
	"inspection-server/pkg/common"
	"inspection-server/pkg/db"
	pdfPrint "inspection-server/pkg/print"
	"inspection-server/pkg/send"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

func Inspection(task *apis.Task) error {
	task.State = "巡检中"
	logrus.Infof("Starting inspection for task ID: %s", task.ID)

	err := db.UpdateTask(task)
	if err != nil {
		logrus.Errorf("Failed to update task state to '巡检中' for task ID %s: %v", task.ID, err)
		return err
	}

	template, err := db.GetTemplate(task.TemplateID)
	if err != nil {
		logrus.Errorf("Failed to get template for task ID %s: %v", task.ID, err)
		return err
	}

	clients := apis.NewClients()
	err = common.GenerateKubeconfig(clients)
	if err != nil {
		logrus.Errorf("Failed to generate kubeconfig: %v", err)
		return err
	}

	report := apis.NewReport()
	kubernetes := apis.NewKubernetes()

	allGrafanaInspections, err := GetAllGrafanaInspections()
	if err != nil {
		logrus.Errorf("Failed to get all Grafana inspections: %v", err)
		return err
	}

	level := 0
	var sendMessageDetail []string
	for clusterID, client := range clients {
		for _, k := range template.KubernetesConfig {
			if k.ClusterID == clusterID && k.Enable {
				sendMessageDetail = append(sendMessageDetail, fmt.Sprintf("集群 %s 巡检警告：", k.ClusterName))
				logrus.Infof("Processing inspections for cluster: %s", k.ClusterName)

				clusterCore := apis.NewClusterCore()
				clusterNode := apis.NewClusterNode()
				clusterResource := apis.NewClusterResource()
				coreInspections := apis.NewInspections()
				nodeInspections := apis.NewInspections()
				resourceInspections := apis.NewInspections()

				healthCheck, coreInspectionArray, err := GetHealthCheck(client, k.ClusterName)
				if err != nil {
					logrus.Errorf("Failed to get health check for cluster %s: %v", clusterID, err)
					return err
				}
				coreInspections = append(coreInspections, coreInspectionArray...)

				NodeNodeArray, nodeInspectionArray, err := GetNodes(client, k.ClusterNodeConfig.NodeConfig)
				if err != nil {
					logrus.Errorf("Failed to get nodes for cluster %s: %v", clusterID, err)
					return err
				}
				nodeInspections = append(nodeInspections, nodeInspectionArray...)

				ResourceWorkloadArray, resourceInspectionArray, err := GetWorkloads(client, k.ClusterResourceConfig.WorkloadConfig)
				if err != nil {
					logrus.Errorf("Failed to get workloads for cluster %s: %v", clusterID, err)
					return err
				}
				resourceInspections = append(resourceInspections, resourceInspectionArray...)

				if k.ClusterResourceConfig.NamespaceConfig.Enable {
					ResourceNamespaceArray, resourceInspectionArray, err := GetNamespaces(client)
					if err != nil {
						logrus.Errorf("Failed to get namespaces for cluster %s: %v", clusterID, err)
						return err
					}

					clusterResource.Namespace = ResourceNamespaceArray
					resourceInspections = append(resourceInspections, resourceInspectionArray...)
				}

				if k.ClusterResourceConfig.ServiceConfig.Enable {
					ResourceServiceArray, resourceInspectionArray, err := GetServices(client)
					if err != nil {
						logrus.Errorf("Failed to get services for cluster %s: %v", clusterID, err)
						return err
					}

					clusterResource.Service = ResourceServiceArray
					resourceInspections = append(resourceInspections, resourceInspectionArray...)
				}

				if k.ClusterResourceConfig.IngressConfig.Enable {
					ResourceIngressArray, resourceInspectionArray, err := GetIngress(client)
					if err != nil {
						logrus.Errorf("Failed to get ingress for cluster %s: %v", clusterID, err)
						return err
					}

					clusterResource.Ingress = ResourceIngressArray
					resourceInspections = append(resourceInspections, resourceInspectionArray...)
				}

				clusterCore.HealthCheck = healthCheck
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
					if c.Level > level {
						level = c.Level
					}
					if c.Level >= 2 {
						sendMessageDetail = append(sendMessageDetail, fmt.Sprintf("%s", c.Title))
					}
				}
				for _, n := range nodeInspections {
					if n.Level > level {
						level = n.Level
					}
					if n.Level >= 2 {
						sendMessageDetail = append(sendMessageDetail, fmt.Sprintf("%s", n.Title))
					}
				}
				for _, r := range resourceInspections {
					if r.Level > level {
						level = r.Level
					}
					if r.Level >= 2 {
						sendMessageDetail = append(sendMessageDetail, fmt.Sprintf("%s", r.Title))
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

	var rating string
	switch level {
	case 0:
		rating = "优"
	case 1:
		rating = "高"
	case 2:
		rating = "中"
	case 3:
		rating = "低"
	default:
		rating = "未知"
	}

	report = &apis.Report{
		ID: common.GetUUID(),
		Global: &apis.Global{
			Name:       task.Name,
			Rating:     rating,
			ReportTime: time.Now().Format(time.DateTime),
		},
		Kubernetes: kubernetes,
	}
	err = db.CreateReport(report)
	if err != nil {
		logrus.Errorf("Failed to create report: %v", err)
		return err
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`该巡检报告的健康等级为: %s\n`, report.Global.Rating))

	for _, s := range sendMessageDetail {
		sb.WriteString(fmt.Sprintf(`%s\n`, s))
	}

	p := pdfPrint.NewPrint()
	p.URL = "http://127.0.0.1/#/inspection/result-pdf-view/" + report.ID
	p.ReportTime = report.Global.ReportTime
	err = pdfPrint.FullScreenshot(p)
	if err != nil {
		logrus.Errorf("Failed to take screenshot for report ID %s: %v", report.ID, err)
		return err
	}

	if task.NotifyID != "" {
		notify, err := db.GetNotify(task.NotifyID)
		if err != nil {
			logrus.Errorf("Failed to get notification details for NotifyID %s: %v", task.NotifyID, err)
			return err
		}

		err = send.Notify(notify.AppID, notify.AppSecret, common.GetReportFileName(p.ReportTime), common.PrintPDFPath+common.GetReportFileName(p.ReportTime), sb.String())
		if err != nil {
			logrus.Errorf("Failed to send notification: %v", err)
			return err
		}
	}

	task.EndTime = time.Now().Format(time.DateTime)
	task.Rating = report.Global.Rating
	task.ReportID = report.ID
	task.State = "巡检完成"
	err = db.UpdateTask(task)
	if err != nil {
		logrus.Errorf("Failed to update task state to '巡检完成' for task ID %s: %v", task.ID, err)
		return err
	}

	logrus.Infof("Inspection completed for task ID: %s", task.ID)
	return nil
}
