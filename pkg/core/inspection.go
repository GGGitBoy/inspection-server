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
)

func Inspection(task *apis.Task) (error, strings.Builder) {
	var errMessage strings.Builder

	task.State = "巡检中"
	fmt.Println(task.ID)
	fmt.Println(task.State)
	err := db.UpdateTask(task)
	if err != nil {
		return err, errMessage
	}

	template, err := db.GetTemplate(task.TemplateID)
	if err != nil {
		return err, errMessage
	}

	clients := apis.NewClients()
	err = common.GenerateKubeconfig(clients)
	if err != nil {
		return err, errMessage
	}

	report := apis.NewReport()
	kubernetes := apis.NewKubernetes()

	allGrafanaInspections, err := GetAllGrafanaInspections()
	if err != nil {
		errMessage.WriteString(fmt.Sprintf("获取图表告警失败: %v\n", err))
		return err, errMessage
	}

	level := 0
	var sendMessageDetail []string
	for clusterID, client := range clients {
		for _, k := range template.KubernetesConfig {
			if k.ClusterID == clusterID && k.Enable {
				sendMessageDetail = append(sendMessageDetail, fmt.Sprintf("集群 %s 巡检警告：", k.ClusterName))

				clusterCore := apis.NewClusterCore()
				clusterNode := apis.NewClusterNode()
				clusterResource := apis.NewClusterResource()
				coreInspections := apis.NewInspections()
				nodeInspections := apis.NewInspections()
				resourceInspections := apis.NewInspections()

				NodeNodeArray, nodeInspectionArray, err := GetNodes(client, k.ClusterNodeConfig.NodeConfig)
				if err != nil {
					errMessage.WriteString(fmt.Sprintf("获取集群 %s 节点相关巡检信息时失败: %v\n", clusterID, err))
				}
				nodeInspections = append(nodeInspections, nodeInspectionArray...)

				ResourceWorkloadArray, resourceInspectionArray, err := GetWorkloads(client, k.ClusterResourceConfig.WorkloadConfig)
				if err != nil {
					errMessage.WriteString(fmt.Sprintf("获取集群 %s 工作负载相关巡检信息时失败: %v\n", clusterID, err))
				}
				resourceInspections = append(resourceInspections, resourceInspectionArray...)

				if k.ClusterResourceConfig.NamespaceConfig.Enable {
					ResourceNamespaceArray, resourceInspectionArray, err := GetNamespaces(client)
					if err != nil {
						errMessage.WriteString(fmt.Sprintf("获取集群 %s 命名空间相关巡检信息时失败: %v\n", clusterID, err))
					}

					clusterResource.Namespace = ResourceNamespaceArray
					resourceInspections = append(resourceInspections, resourceInspectionArray...)
				}

				if k.ClusterResourceConfig.ServiceConfig.Enable {
					ResourceServiceArray, resourceInspectionArray, err := GetServices(client)
					if err != nil {
						errMessage.WriteString(fmt.Sprintf("获取集群 %s Service 相关巡检信息时失败: %v\n", clusterID, err))
					}

					clusterResource.Service = ResourceServiceArray
					resourceInspections = append(resourceInspections, resourceInspectionArray...)
				}

				if k.ClusterResourceConfig.IngressConfig.Enable {
					ResourceIngressArray, resourceInspectionArray, err := GetIngress(client)
					if err != nil {
						errMessage.WriteString(fmt.Sprintf("获取集群 %s Ingress 相关巡检信息时失败: %v\n", clusterID, err))
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
	if level == 0 {
		rating = "优"
	} else if level == 1 {
		rating = "高"
	} else if level == 2 {
		rating = "中"
	} else if level == 3 {
		rating = "低"
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
		return err, errMessage
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`该巡检报告的健康等级为: %s\n`, report.Global.Rating))

	for _, s := range sendMessageDetail {
		sb.WriteString(fmt.Sprintf(`%s\n`, s))
	}

	if task.NotifyID != "" {
		notify, err := db.GetNotify(task.NotifyID)
		if err != nil {
			return err, errMessage
		}

		p := pdfPrint.NewPrint()
		p.URL = "http://127.0.0.1/#/inspection/result-pdf-view/" + report.ID
		p.ReportTime = report.Global.ReportTime
		err = pdfPrint.FullScreenshot(p)
		if err != nil {
			return err, errMessage
		}

		err = send.Notify(notify.AppID, notify.AppSecret, common.GetReportFileName(p.ReportTime), common.PrintPDFPath+common.GetReportFileName(p.ReportTime), sb.String())
		if err != nil {
			return err, errMessage
		}
	}

	task.EndTime = time.Now().Format(time.DateTime)
	task.Rating = report.Global.Rating
	task.ReportID = report.ID
	task.State = "巡检完成"
	task.ErrMessage = errMessage.String()
	err = db.UpdateTask(task)
	if err != nil {
		return err, errMessage
	}

	return nil, errMessage
}
