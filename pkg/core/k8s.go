package core

import (
	"context"
	"encoding/json"
	"fmt"
	"inspection-server/pkg/apis"
	"inspection-server/pkg/common"
	"io"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/utils/strings/slices"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
)

var (
	warning = "warning"
	success = "success"
)

//func GetGlobal(report *apis.Report) error {
//	report.Global.Warnings = []apis.Warning{
//		{Title: "红色警告", Message: "红色：影响 rancher 或者 集群运行的", Type: 3},
//		{Title: "黄色警告", Message: "黄色：不影响 rancher 或者 集群运行的，但是有风险的", Type: 2},
//		{Title: "灰色警告", Message: "灰色：没有太大风险的，不过最好也可以处理的", Type: 1},
//	}
//
//	var rating int
//	for _, w := range report.Global.Warnings {
//		if w.Type > rating {
//			rating = w.Type
//		}
//	}
//
//	report.Global.ReportTime = time.Now().Format(time.DateTime)
//	report.Global.Rating = rating
//
//	return nil
//}

func GetNodes(client *apis.Client, nodesConfig []*apis.NodeConfig) ([]*apis.Node, []*apis.Inspection, error) {
	nodeNodeArray := apis.NewNodes()
	nodeInspections := apis.NewInspections()

	set := labels.Set(map[string]string{"name": "inspection-agent"})
	podList, err := client.Clientset.CoreV1().Pods(common.InspectionNamespace).List(context.TODO(), metav1.ListOptions{LabelSelector: set.String()})
	if err != nil {
		return nil, nil, err
	}

	for _, pod := range podList.Items {
		for _, n := range nodesConfig {
			if slices.Contains(n.Names, pod.Spec.NodeName) {
				node, err := client.Clientset.CoreV1().Nodes().Get(context.TODO(), pod.Spec.NodeName, metav1.GetOptions{})
				if err != nil {
					return nil, nil, err
				}

				podLimits := getResourceList(node.Annotations["management.cattle.io/pod-limits"])
				podRequests := getResourceList(node.Annotations["management.cattle.io/pod-requests"])

				limitsCPU := podLimits.Cpu().Value()
				limitsMemory := podLimits.Memory().Value()
				requestsCPU := podRequests.Cpu().Value()
				requestsMemory := podRequests.Memory().Value()
				requestsPods := podRequests.Pods().Value()
				allocatableCPU, _ := node.Status.Allocatable.Cpu().AsInt64()
				allocatableMemory, _ := node.Status.Allocatable.Memory().AsInt64()
				allocatablePods, _ := node.Status.Allocatable.Pods().AsInt64()

				//fmt.Println(podLimits)
				//fmt.Println(limitsCPU)
				//fmt.Println(limitsMemory)
				//fmt.Println(requestsCPU)
				//fmt.Println(requestsMemory)
				//fmt.Println(requestsPods)
				//fmt.Println("limitsCPU:")
				//fmt.Println(float64(limitsCPU) / float64(allocatableCPU))
				if float64(limitsCPU)/float64(allocatableCPU) > 0.8 {
					//nodeInspections = append(nodeInspections, apis.NewInspection(fmt.Sprintf("Node %s limits CPU 超过 80%", pod.Spec.NodeName), fmt.Sprintf("limits CPU %d, allocatable CPU %d", limitsCPU, allocatableCPU), 2))
					nodeInspections = append(nodeInspections, apis.NewInspection(fmt.Sprintf("Node %s limits CPU 超过 80%", pod.Spec.NodeName), fmt.Sprintf(""), 2))
				}

				if float64(limitsMemory)/float64(allocatableMemory) > 0.8 {
					nodeInspections = append(nodeInspections, apis.NewInspection(fmt.Sprintf("Node %s limits Memory 超过 80%", pod.Spec.NodeName), fmt.Sprintf(""), 2))
				}

				if float64(requestsCPU)/float64(allocatableCPU) > 0.8 {
					nodeInspections = append(nodeInspections, apis.NewInspection(fmt.Sprintf("Node %s requests CPU 超过 80%", pod.Spec.NodeName), fmt.Sprintf(""), 2))
				}

				if float64(requestsMemory)/float64(allocatableMemory) > 0.8 {
					nodeInspections = append(nodeInspections, apis.NewInspection(fmt.Sprintf("Node %s requests Memory 超过 80%", pod.Spec.NodeName), fmt.Sprintf(""), 2))
				}

				if float64(requestsPods)/float64(allocatablePods) > 0.8 {
					nodeInspections = append(nodeInspections, apis.NewInspection(fmt.Sprintf("Node %s requests Pods超过 80%", pod.Spec.NodeName), fmt.Sprintf(""), 2))
				}

				var commands []string
				for _, c := range n.Commands {
					commands = append(commands, c.Description+": "+c.Command)
				}

				fmt.Println(commands)
				command := "/opt/inspection/inspection.sh"
				stdout, stderr, err := ExecToPodThroughAPI(client.Clientset, client.Config, command, commands, pod.Namespace, pod.Name, "inspection-agent-container")
				if err != nil {
					return nil, nil, err
				}

				var results []apis.CommandCheckResult
				err = json.Unmarshal([]byte(stdout), &results)
				if err != nil {
					return nil, nil, err
				}

				for _, r := range results {
					if r.Error != "" {
						nodeInspections = append(nodeInspections, apis.NewInspection(fmt.Sprintf("Node %s: %s 警告", pod.Spec.NodeName, r.Description), fmt.Sprintf("%s 检查报错 %s", r.Description, r.Error), 2))
					}
				}

				nodeData := &apis.Node{
					Name: pod.Spec.NodeName,
					Resource: &apis.Resource{
						LimitsCPU:         limitsCPU,
						LimitsMemory:      limitsMemory,
						RequestsCPU:       requestsCPU,
						RequestsMemory:    requestsMemory,
						RequestsPods:      requestsPods,
						AllocatableCPU:    allocatableCPU,
						AllocatableMemory: allocatableMemory,
						AllocatablePods:   allocatablePods,
					},
					Commands: &apis.Command{
						Stdout: results,
						Stderr: stderr,
					},
				}

				nodeNodeArray = append(nodeNodeArray, nodeData)
			}
		}
	}

	return nodeNodeArray, nodeInspections, nil
}

func ExecToPodThroughAPI(clientset *kubernetes.Clientset, config *rest.Config, command string, commands []string, namespace string, podName string, containerName string) (string, string, error) {
	fmt.Println(podName)
	fmt.Println(namespace)
	req := clientset.CoreV1().RESTClient().
		Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		Param("container", containerName).
		Param("stdin", "true").
		Param("stdout", "true").
		Param("stderr", "true").
		Param("tty", "false").
		Param("command", command)

	for _, c := range commands {
		req.Param("command", c)
	}

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return "", "", err
	}

	var stdout, stderr string
	stdoutWriter := &outputWriter{output: &stdout}
	stderrWriter := &outputWriter{output: &stderr}

	err = exec.StreamWithContext(context.TODO(), remotecommand.StreamOptions{
		Stdin:  os.Stdin,
		Stdout: stdoutWriter,
		Stderr: stderrWriter,
		Tty:    false,
	})

	return stdout, stderr, err
}

type outputWriter struct {
	output *string
}

func (w *outputWriter) Write(p []byte) (n int, err error) {
	*w.output += string(p)
	return len(p), nil
}

func GetWorkloads(client *apis.Client, workloadConfig *apis.WorkloadConfig) (*apis.Workload, []*apis.Inspection, error) {
	ResourceWorkloadArray := apis.NewWorkload()
	resourceInspections := apis.NewInspections()

	deployState := warning
	dsState := warning
	stsState := warning
	jState := warning

	for _, deploy := range workloadConfig.Deployment {
		deployment, err := client.Clientset.AppsV1().Deployments(deploy.Namespace).Get(context.TODO(), deploy.Name, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				continue
			}
			return nil, nil, err
		}

		if isDeploymentAvailable(deployment) {
			deployState = success
		}

		var condition []apis.Condition
		for _, c := range deployment.Status.Conditions {
			condition = append(condition, apis.Condition{
				Type:   string(c.Type),
				Status: string(c.Status),
				Reason: c.Reason,
			})
		}

		set := labels.Set(deployment.Spec.Selector.MatchLabels)
		pods, err := GetPod(deploy.Regexp, deployment.Namespace, set, client.Clientset)
		if err != nil {
			return nil, nil, err
		}

		deploymentData := &apis.WorkloadData{
			Name:      deployment.Name,
			Namespace: deployment.Namespace,
			Pods:      pods,
			Status: &apis.Status{
				State:     deployState,
				Condition: condition,
			},
		}

		ResourceWorkloadArray.Deployment = append(ResourceWorkloadArray.Deployment, deploymentData)
		if deployState == warning {
			resourceInspections = append(resourceInspections, apis.NewInspection(fmt.Sprintf("Deployment %s 警告", deploymentData.Name), fmt.Sprintf("命名空间 %s 下的 Deployment %s 处于非健康状态", deploymentData.Namespace, deploymentData.Name), 2))
		}
	}

	for _, ds := range workloadConfig.Daemonset {
		daemonSet, err := client.Clientset.AppsV1().DaemonSets(ds.Namespace).Get(context.TODO(), ds.Name, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				continue
			}
			return nil, nil, err
		}

		if isDaemonSetAvailable(daemonSet) {
			dsState = success
		}

		var condition []apis.Condition
		for _, c := range daemonSet.Status.Conditions {
			condition = append(condition, apis.Condition{
				Type:   string(c.Type),
				Status: string(c.Status),
				Reason: c.Reason,
			})
		}

		set := labels.Set(daemonSet.Spec.Selector.MatchLabels)
		pods, err := GetPod(ds.Regexp, daemonSet.Namespace, set, client.Clientset)
		if err != nil {
			return nil, nil, err
		}

		daemonSetData := &apis.WorkloadData{
			Name:      daemonSet.Name,
			Namespace: daemonSet.Namespace,
			Pods:      pods,
			Status: &apis.Status{
				State:     dsState,
				Condition: condition,
			},
		}

		ResourceWorkloadArray.Daemonset = append(ResourceWorkloadArray.Daemonset, daemonSetData)
		if dsState == warning {
			resourceInspections = append(resourceInspections, apis.NewInspection(fmt.Sprintf("Daemonset %s 警告", daemonSetData.Name), fmt.Sprintf("命名空间 %s 下的 Daemonset %s 处于非健康状态", daemonSetData.Namespace, daemonSetData.Name), 2))
		}
	}

	for _, sts := range workloadConfig.Statefulset {
		statefulset, err := client.Clientset.AppsV1().StatefulSets(sts.Namespace).Get(context.TODO(), sts.Name, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				continue
			}
			return nil, nil, err
		}

		if isStatefulSetAvailable(statefulset) {
			stsState = success
		}

		var condition []apis.Condition
		for _, c := range statefulset.Status.Conditions {
			condition = append(condition, apis.Condition{
				Type:   string(c.Type),
				Status: string(c.Status),
				Reason: c.Reason,
			})
		}

		set := labels.Set(statefulset.Spec.Selector.MatchLabels)
		pods, err := GetPod(sts.Regexp, statefulset.Namespace, set, client.Clientset)
		if err != nil {
			return nil, nil, err
		}

		statefulSetData := &apis.WorkloadData{
			Name:      statefulset.Name,
			Namespace: statefulset.Namespace,
			Pods:      pods,
			Status: &apis.Status{
				State:     stsState,
				Condition: condition,
			},
		}

		ResourceWorkloadArray.Statefulset = append(ResourceWorkloadArray.Statefulset, statefulSetData)
		if stsState == warning {
			resourceInspections = append(resourceInspections, apis.NewInspection(fmt.Sprintf("Statefulset %s 警告", statefulSetData.Name), fmt.Sprintf("命名空间 %s 下的 Statefulset %s 处于非健康状态", statefulSetData.Namespace, statefulSetData.Name), 2))
		}
	}

	for _, j := range workloadConfig.Job {
		job, err := client.Clientset.BatchV1().Jobs(j.Namespace).Get(context.TODO(), j.Name, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				continue
			}
			return nil, nil, err
		}

		if isJobCompleted(job) {
			jState = success
		}

		var condition []apis.Condition
		for _, c := range job.Status.Conditions {
			condition = append(condition, apis.Condition{
				Type:   string(c.Type),
				Status: string(c.Status),
				Reason: c.Reason,
			})
		}

		set := labels.Set(job.Spec.Selector.MatchLabels)
		pods, err := GetPod(j.Regexp, j.Namespace, set, client.Clientset)
		if err != nil {
			return nil, nil, err
		}

		jobData := &apis.WorkloadData{
			Name:      job.Name,
			Namespace: job.Namespace,
			Pods:      pods,
			Status: &apis.Status{
				State:     jState,
				Condition: condition,
			},
		}

		ResourceWorkloadArray.Job = append(ResourceWorkloadArray.Job, jobData)
		if jState == warning {
			resourceInspections = append(resourceInspections, apis.NewInspection(fmt.Sprintf("Job %s 警告", jobData.Name), fmt.Sprintf("命名空间 %s 下的 Job %s 处于非健康状态", jobData.Namespace, jobData.Name), 2))
		}
	}

	return ResourceWorkloadArray, resourceInspections, nil
}

func GetPod(regexpString, namespace string, set labels.Set, clientset *kubernetes.Clientset) ([]*apis.Pod, error) {
	pods := apis.NewPods()

	podList, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: set.String()})
	if err != nil {
		return nil, err
	}

	line := int64(10)
	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, pod := range podList.Items {
		wg.Add(1)
		fmt.Println(pod.Name)
		go func(pod corev1.Pod) {
			fmt.Println("pod name 3")
			defer wg.Done()

			getLog := clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{TailLines: &line})
			podLogs, err := getLog.Stream(context.TODO())
			if err != nil {
				fmt.Errorf("Error getting logs for pod %s: %v\n", pod.Name, err)
				return
			}
			defer podLogs.Close()

			fmt.Println("pod name 4")
			logs, err := io.ReadAll(podLogs)
			if err != nil {
				fmt.Errorf("Error getting logs for pod %s: %v\n", pod.Name, err)
				return
			}

			var str []string
			if regexpString == "" {
				regexpString = ".*"
			}

			re, err := regexp.Compile(regexpString)
			if err != nil {
				fmt.Errorf("Error getting logs for pod %s: %v\n", pod.Name, err)
				return
			}

			str = re.FindAllString(string(logs), -1)
			if str == nil {
				str = []string{}
			}
			fmt.Println("pod name 1")
			mu.Lock()
			pods = append(pods, &apis.Pod{
				Name: pod.Name,
				Log:  str,
			})
			mu.Unlock()
			fmt.Println("pod name 2")
		}(pod)
	}
	wg.Wait()

	fmt.Println(pods)

	return pods, nil
}

func GetNamespaces(client *apis.Client) ([]*apis.Namespace, []*apis.Inspection, error) {
	resourceInspections := apis.NewInspections()
	namespaces := apis.NewNamespaces()

	namespaceList, err := client.Clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, nil, err
	}

	for _, n := range namespaceList.Items {
		var emptyResourceQuota, emptyResource bool

		podList, err := client.Clientset.CoreV1().Pods(n.Name).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, nil, err
		}

		serviceList, err := client.Clientset.CoreV1().Services(n.Name).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, nil, err
		}

		deploymentList, err := client.Clientset.AppsV1().Deployments(n.Name).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, nil, err
		}

		replicaSetList, err := client.Clientset.AppsV1().ReplicaSets(n.Name).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, nil, err
		}

		statefulSetList, err := client.Clientset.AppsV1().StatefulSets(n.Name).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, nil, err
		}

		daemonSetList, err := client.Clientset.AppsV1().DaemonSets(n.Name).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, nil, err
		}

		jobList, err := client.Clientset.BatchV1().Jobs(n.Name).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, nil, err
		}

		secretList, err := client.Clientset.CoreV1().Secrets(n.Name).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, nil, err
		}

		configMapList, err := client.Clientset.CoreV1().ConfigMaps(n.Name).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, nil, err
		}

		resourceQuotaList, err := client.Clientset.CoreV1().ResourceQuotas(n.GetName()).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, nil, err
		}

		if len(resourceQuotaList.Items) == 0 {
			emptyResourceQuota = true
			resourceInspections = append(resourceInspections, apis.NewInspection(fmt.Sprintf("命名空间 %s 没有设置配额", n.Name), fmt.Sprintf(""), 1))
		}

		if (len(podList.Items) + len(serviceList.Items) + len(deploymentList.Items) + len(replicaSetList.Items) + len(statefulSetList.Items) + len(daemonSetList.Items) + len(jobList.Items) + len(secretList.Items) + len(configMapList.Items)) == 0 {
			emptyResource = true
			resourceInspections = append(resourceInspections, apis.NewInspection(fmt.Sprintf("命名空间 %s 下资源为空", n.Name), fmt.Sprintf("检查对象为 Pod、Service、Deployment、Replicaset、Statefulset、Daemonset、Job、Secret、ConfigMap"), 1))
		}

		namespaces = append(namespaces, &apis.Namespace{
			Name:               n.Name,
			EmptyResourceQuota: emptyResourceQuota,
			EmptyResource:      emptyResource,
			PodCount:           len(podList.Items),
			ServiceCount:       len(serviceList.Items),
			DeploymentCount:    len(deploymentList.Items),
			ReplicasetCount:    len(replicaSetList.Items),
			StatefulsetCount:   len(statefulSetList.Items),
			DaemonsetCount:     len(daemonSetList.Items),
			JobCount:           len(jobList.Items),
			SecretCount:        len(secretList.Items),
			ConfigMapCount:     len(configMapList.Items),
		})
	}

	return namespaces, resourceInspections, nil
}

func GetServices(client *apis.Client) ([]*apis.Service, []*apis.Inspection, error) {
	resourceInspections := apis.NewInspections()
	services := apis.NewServices()

	// 获取所有命名空间下的所有 Services
	serviceList, err := client.Clientset.CoreV1().Services("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, nil, err
	}

	for _, s := range serviceList.Items {
		// 获取对应的 Endpoints
		endpoints, err := client.Clientset.CoreV1().Endpoints(s.Namespace).Get(context.TODO(), s.Name, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				resourceInspections = append(resourceInspections, apis.NewInspection(fmt.Sprintf("命名空间 %s 下 Service %s 找不到对应 endpoint", s.Namespace, s.Name), fmt.Sprintf(""), 1))
				continue
			}
			return nil, nil, err
		}

		// 检查 Endpoints 是否为空
		var emptyEndpoints bool
		if len(endpoints.Subsets) == 0 {
			emptyEndpoints = true
			resourceInspections = append(resourceInspections, apis.NewInspection(fmt.Sprintf("命名空间 %s 下 Service %s 对应 Endpoints 没有 Subsets", s.Namespace, s.Name), fmt.Sprintf(""), 1))
		}

		services = append(services, &apis.Service{
			Name:           s.Name,
			Namespace:      s.Namespace,
			EmptyEndpoints: emptyEndpoints,
		})
	}

	return services, resourceInspections, nil
}

func GetIngress(client *apis.Client) ([]*apis.Ingress, []*apis.Inspection, error) {
	resourceInspections := apis.NewInspections()
	ingress := apis.NewIngress()

	// 获取所有命名空间下的所有 Ingress
	ingresseList, err := client.Clientset.NetworkingV1().Ingresses("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Error listing ingresses: %v", err)
	}

	// 使用 map 来记录 host+path 与 ingress 名称的映射
	ingressMap := make(map[string][]string)
	for _, i := range ingresseList.Items {
		for _, rule := range i.Spec.Rules {
			host := rule.Host
			for _, path := range rule.HTTP.Paths {
				key := host + path.Path
				ingressMap[key] = append(ingressMap[key], fmt.Sprintf("%s/%s", i.Namespace, i.Name))
			}
		}

		ingress = append(ingress, &apis.Ingress{
			Name:          i.Name,
			Namespace:     i.Namespace,
			DuplicatePath: false,
		})
	}

	duplicateIngress := make(map[string]int)
	// 检查是否有重名且 Path 路径相同的 Ingress
	for key, ingressNames := range ingressMap {
		if len(ingressNames) > 1 {
			for _, ingressName := range ingressNames {
				duplicateIngress[ingressName] = 1
			}
			fmt.Printf("发现重名且 Path 路径相同的 Ingress: %s, Ingress 列表: %v\n", key, ingressNames)
		}
	}

	if len(duplicateIngress) > 0 {
		var result []string
		for NamespaceName, _ := range duplicateIngress {
			parts := strings.Split(NamespaceName, "/")
			for index, i := range ingress {
				if parts[0] == i.Namespace && parts[1] == i.Name {
					ingress[index] = &apis.Ingress{
						Name:          i.Name,
						Namespace:     i.Namespace,
						DuplicatePath: true,
					}
				}
			}

			result = append(result, NamespaceName)
		}

		resourceInspections = append(resourceInspections, apis.NewInspection(fmt.Sprintf("Ingress %s 存在重复的 Path", strings.Join(result, ", ")), fmt.Sprintf(""), 1))
	}

	return ingress, resourceInspections, nil
}

func getResourceList(val string) corev1.ResourceList {
	if val == "" {
		return nil
	}
	result := corev1.ResourceList{}
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		return corev1.ResourceList{}
	}
	return result
}

func isDeploymentAvailable(deployment *appsv1.Deployment) bool {
	for _, condition := range deployment.Status.Conditions {
		if (condition.Type == "Failed" && condition.Status == "False") || condition.Reason == "Error" {
			return false
		}
	}
	return deployment.Status.AvailableReplicas >= *deployment.Spec.Replicas
}

func isDaemonSetAvailable(daemonset *appsv1.DaemonSet) bool {
	return daemonset.Status.NumberAvailable >= daemonset.Status.DesiredNumberScheduled
}

func isStatefulSetAvailable(statefulset *appsv1.StatefulSet) bool {
	return statefulset.Status.ReadyReplicas >= *statefulset.Spec.Replicas
}

func isJobCompleted(job *batchv1.Job) bool {
	return job.Status.Succeeded >= *job.Spec.Completions
}
