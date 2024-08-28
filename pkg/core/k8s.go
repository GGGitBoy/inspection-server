package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
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

	Commands = []*apis.CommandConfig{
		{
			Description: "API Server Ready Check",
			Command:     "kubectl get --raw='/readyz'",
		},
		{
			Description: "API Server Live Check",
			Command:     "kubectl get --raw='/livez'",
		},
		{
			Description: "ETCD Ready Check",
			Command:     "kubectl get --raw='/readyz/etcd'",
		},
		{
			Description: "ETCD Live Check",
			Command:     "kubectl get --raw='/livez/etcd'",
		},
	}
)

func GetHealthCheck(client *apis.Client, clusterName string) (*apis.HealthCheck, []*apis.Inspection, error) {
	healthCheck := apis.NewHealthCheck()
	coreInspections := apis.NewInspections()

	set := labels.Set(map[string]string{"name": "inspection-agent"})
	podList, err := client.Clientset.CoreV1().Pods(common.InspectionNamespace).List(context.TODO(), metav1.ListOptions{LabelSelector: set.String()})
	if err != nil {
		log.Printf("Error listing pods in namespace %s: %v", common.InspectionNamespace, err)
		return nil, nil, err
	}

	if len(podList.Items) > 0 {
		var commands []string
		for _, c := range Commands {
			commands = append(commands, c.Description+": "+c.Command)
		}

		command := "/opt/inspection/inspection.sh"
		stdout, stderr, err := ExecToPodThroughAPI(client.Clientset, client.Config, command, commands, podList.Items[0].Namespace, podList.Items[0].Name, "inspection-agent-container")
		if err != nil {
			log.Printf("Error executing command in pod %s: %v", podList.Items[0].Name, err)
			return nil, nil, err
		}

		if stderr != "" {
			log.Printf("Stderr from pod %s: %s", podList.Items[0].Name, stderr)
			return nil, nil, errors.New(stderr)
		}

		var results []apis.CommandCheckResult
		err = json.Unmarshal([]byte(stdout), &results)
		if err != nil {
			log.Printf("Error unmarshalling stdout for pod %s: %v", podList.Items[0].Name, err)
			return nil, nil, err
		}

		for _, r := range results {
			if r.Error != "" {
				coreInspections = append(coreInspections, apis.NewInspection(fmt.Sprintf("cluster %s (%s) failed", clusterName, r.Description), fmt.Sprintf("%s", r.Error), 3))
			}

			switch r.Description {
			case "API Server Ready Check":
				healthCheck.APIServerReady = &r
			case "API Server Live Check":
				healthCheck.APIServerLive = &r
			case "ETCD Ready Check":
				healthCheck.EtcdReady = &r
			case "ETCD Live Check":
				healthCheck.EtcdLive = &r
			}
		}
	}

	return healthCheck, coreInspections, nil
}

func GetNodes(client *apis.Client, nodesConfig []*apis.NodeConfig) ([]*apis.Node, []*apis.Inspection, error) {
	nodeNodeArray := apis.NewNodes()
	nodeInspections := apis.NewInspections()

	set := labels.Set(map[string]string{"name": "inspection-agent"})
	podList, err := client.Clientset.CoreV1().Pods(common.InspectionNamespace).List(context.TODO(), metav1.ListOptions{LabelSelector: set.String()})
	if err != nil {
		log.Printf("Error listing pods in namespace %s: %v", common.InspectionNamespace, err)
		return nil, nil, err
	}

	for _, pod := range podList.Items {
		for _, n := range nodesConfig {
			if slices.Contains(n.Names, pod.Spec.NodeName) {
				node, err := client.Clientset.CoreV1().Nodes().Get(context.TODO(), pod.Spec.NodeName, metav1.GetOptions{})
				if err != nil {
					log.Printf("Error getting node %s: %v", pod.Spec.NodeName, err)
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

				if float64(limitsCPU)/float64(allocatableCPU) > 0.8 {
					nodeInspections = append(nodeInspections, apis.NewInspection(fmt.Sprintf("Node %s High Limits CPU", pod.Spec.NodeName), fmt.Sprintf("节点 %s limits CPU 超过百分之 80", pod.Spec.NodeName), 2))
					log.Printf("Node %s High Limits CPU: limits CPU %d, allocatable CPU %d", pod.Spec.NodeName, limitsCPU, allocatableCPU)
				}

				if float64(limitsMemory)/float64(allocatableMemory) > 0.8 {
					nodeInspections = append(nodeInspections, apis.NewInspection(fmt.Sprintf("Node %s High Limits Memory", pod.Spec.NodeName), fmt.Sprintf("节点 %s limits Memory 超过百分之 80", pod.Spec.NodeName), 2))
					log.Printf("Node %s High Limits Memory: limits Memory %d, allocatable Memory %d", pod.Spec.NodeName, limitsMemory, allocatableMemory)
				}

				if float64(requestsCPU)/float64(allocatableCPU) > 0.8 {
					nodeInspections = append(nodeInspections, apis.NewInspection(fmt.Sprintf("Node %s High Requests CPU", pod.Spec.NodeName), fmt.Sprintf("节点 %s requests CPU 超过百分之 80", pod.Spec.NodeName), 2))
					log.Printf("Node %s High Requests CPU: requests CPU %d, allocatable CPU %d", pod.Spec.NodeName, requestsCPU, allocatableCPU)
				}

				if float64(requestsMemory)/float64(allocatableMemory) > 0.8 {
					nodeInspections = append(nodeInspections, apis.NewInspection(fmt.Sprintf("Node %s High Requests Memory", pod.Spec.NodeName), fmt.Sprintf("节点 %s requests Memory 超过百分之 80", pod.Spec.NodeName), 2))
					log.Printf("Node %s High Requests Memory: requests Memory %d, allocatable Memory %d", pod.Spec.NodeName, requestsMemory, allocatableMemory)
				}

				if float64(requestsPods)/float64(allocatablePods) > 0.8 {
					nodeInspections = append(nodeInspections, apis.NewInspection(fmt.Sprintf("Node %s High Requests Pods", pod.Spec.NodeName), fmt.Sprintf("节点 %s requests Pods 超过百分之 80", pod.Spec.NodeName), 2))
					log.Printf("Node %s High Requests Pods: requests Pods %d, allocatable Pods %d", pod.Spec.NodeName, requestsPods, allocatablePods)
				}

				var commands []string
				for _, c := range n.Commands {
					commands = append(commands, c.Description+": "+c.Command)
				}

				log.Printf("Commands to execute on node %s: %v", pod.Spec.NodeName, commands)
				command := "/opt/inspection/inspection.sh"
				stdout, stderr, err := ExecToPodThroughAPI(client.Clientset, client.Config, command, commands, pod.Namespace, pod.Name, "inspection-agent-container")
				if err != nil {
					log.Printf("Error executing command in pod %s: %v", pod.Name, err)
					return nil, nil, err
				}

				if stderr != "" {
					log.Printf("Stderr from pod %s: %s", pod.Name, stderr)
				}

				var results []apis.CommandCheckResult
				err = json.Unmarshal([]byte(stdout), &results)
				if err != nil {
					log.Printf("Error unmarshalling stdout for pod %s: %v", pod.Name, err)
					return nil, nil, err
				}

				for _, r := range results {
					if r.Error != "" {
						nodeInspections = append(nodeInspections, apis.NewInspection(fmt.Sprintf("Node %s (%s)", pod.Spec.NodeName, r.Description), fmt.Sprintf("%s", r.Error), 2))
						log.Printf("Node %s inspection failed (%s): %s", pod.Spec.NodeName, r.Description, r.Error)
					}
				}

				nodeData := &apis.Node{
					Name:   pod.Spec.NodeName,
					HostIP: pod.Status.HostIP,
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
	log.Printf("Starting exec to pod: %s, namespace: %s, container: %s", podName, namespace, containerName)
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
	log.Printf("Executing command: %s with additional commands: %v", command, commands)

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		log.Printf("Error creating SPDY executor: %v", err)
		return "", "", err
	}

	var stdout, stderr string
	stdoutWriter := &outputWriter{output: &stdout}
	stderrWriter := &outputWriter{output: &stderr}

	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  os.Stdin,
		Stdout: stdoutWriter,
		Stderr: stderrWriter,
		Tty:    false,
	})
	if err != nil {
		log.Printf("Error executing command: %v", err)
		return stdout, stderr, err
	}

	log.Printf("Command execution completed. Stdout: %s, Stderr: %s", stdout, stderr)
	return stdout, stderr, nil
}

type outputWriter struct {
	output *string
}

func (w *outputWriter) Write(p []byte) (n int, err error) {
	*w.output += string(p)
	return len(p), nil
}

func GetWorkloads(client *apis.Client, workloadConfig *apis.WorkloadConfig) (*apis.Workload, []*apis.Inspection, error) {
	log.Println("Starting workload inspection")

	ResourceWorkloadArray := apis.NewWorkload()
	resourceInspections := apis.NewInspections()

	for _, deploy := range workloadConfig.Deployment {
		log.Printf("Inspecting Deployment: %s in namespace %s", deploy.Name, deploy.Namespace)
		deployment, err := client.Clientset.AppsV1().Deployments(deploy.Namespace).Get(context.TODO(), deploy.Name, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				log.Printf("Deployment %s not found in namespace %s", deploy.Name, deploy.Namespace)
				continue
			}
			log.Printf("Error getting Deployment %s in namespace %s: %v", deploy.Name, deploy.Namespace, err)
			return nil, nil, err
		}

		deployState := warning
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
			log.Printf("Error getting pods for Deployment %s in namespace %s: %v", deploy.Name, deploy.Namespace, err)
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
			resourceInspections = append(resourceInspections, apis.NewInspection(fmt.Sprintf("Deployment %s 警告", deploymentData.Name), fmt.Sprintf("命名空间 %s 下的 Deployment %s 处于非健康状态", deploymentData.Namespace, deploymentData.Name), defaultLevel(deploy.Level)))
		}
	}

	for _, ds := range workloadConfig.Daemonset {
		log.Printf("Inspecting DaemonSet: %s in namespace %s", ds.Name, ds.Namespace)
		daemonSet, err := client.Clientset.AppsV1().DaemonSets(ds.Namespace).Get(context.TODO(), ds.Name, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				log.Printf("DaemonSet %s not found in namespace %s", ds.Name, ds.Namespace)
				continue
			}
			log.Printf("Error getting DaemonSet %s in namespace %s: %v", ds.Name, ds.Namespace, err)
			return nil, nil, err
		}

		dsState := warning
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
			log.Printf("Error getting pods for DaemonSet %s in namespace %s: %v", ds.Name, ds.Namespace, err)
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
			resourceInspections = append(resourceInspections, apis.NewInspection(fmt.Sprintf("Daemonset %s 警告", daemonSetData.Name), fmt.Sprintf("命名空间 %s 下的 Daemonset %s 处于非健康状态", daemonSetData.Namespace, daemonSetData.Name), defaultLevel(ds.Level)))
		}
	}

	for _, sts := range workloadConfig.Statefulset {
		log.Printf("Inspecting StatefulSet: %s in namespace %s", sts.Name, sts.Namespace)
		statefulset, err := client.Clientset.AppsV1().StatefulSets(sts.Namespace).Get(context.TODO(), sts.Name, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				log.Printf("StatefulSet %s not found in namespace %s", sts.Name, sts.Namespace)
				continue
			}
			log.Printf("Error getting StatefulSet %s in namespace %s: %v", sts.Name, sts.Namespace, err)
			return nil, nil, err
		}

		stsState := warning
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
			log.Printf("Error getting pods for StatefulSet %s in namespace %s: %v", sts.Name, sts.Namespace, err)
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
			resourceInspections = append(resourceInspections, apis.NewInspection(fmt.Sprintf("Statefulset %s 警告", statefulSetData.Name), fmt.Sprintf("命名空间 %s 下的 Statefulset %s 处于非健康状态", statefulSetData.Namespace, statefulSetData.Name), defaultLevel(sts.Level)))
		}
	}

	for _, j := range workloadConfig.Job {
		log.Printf("Inspecting Job: %s in namespace %s", j.Name, j.Namespace)
		job, err := client.Clientset.BatchV1().Jobs(j.Namespace).Get(context.TODO(), j.Name, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				log.Printf("Job %s not found in namespace %s", j.Name, j.Namespace)
				continue
			}
			log.Printf("Error getting Job %s in namespace %s: %v", j.Name, j.Namespace, err)
			return nil, nil, err
		}

		jState := warning
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
			log.Printf("Error getting pods for Job %s in namespace %s: %v", j.Name, j.Namespace, err)
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
			resourceInspections = append(resourceInspections, apis.NewInspection(fmt.Sprintf("Job %s 警告", jobData.Name), fmt.Sprintf("命名空间 %s 下的 Job %s 处于非健康状态", jobData.Namespace, jobData.Name), defaultLevel(j.Level)))
		}
	}

	log.Println("Workload inspection completed")
	return ResourceWorkloadArray, resourceInspections, nil
}

func GetPod(regexpString, namespace string, set labels.Set, clientset *kubernetes.Clientset) ([]*apis.Pod, error) {
	log.Printf("Starting to get pods in namespace %s with labels %s", namespace, set.String())

	pods := apis.NewPods()

	podList, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: set.String()})
	if err != nil {
		log.Printf("Error listing pods in namespace %s: %v", namespace, err)
		return nil, err
	}

	line := int64(10)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, pod := range podList.Items {
		wg.Add(1)
		go func(pod corev1.Pod) {
			defer wg.Done()
			log.Printf("Processing pod: %s", pod.Name)

			getLog := clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{TailLines: &line})
			podLogs, err := getLog.Stream(context.TODO())
			if err != nil {
				log.Printf("Error getting logs for pod %s: %v", pod.Name, err)
				return
			}
			defer podLogs.Close()

			logs, err := io.ReadAll(podLogs)
			if err != nil {
				log.Printf("Error reading logs for pod %s: %v", pod.Name, err)
				return
			}

			var str []string
			if regexpString == "" {
				regexpString = ".*"
			}

			re, err := regexp.Compile(regexpString)
			if err != nil {
				log.Printf("Error compiling regex for pod %s: %v", pod.Name, err)
				return
			}

			str = re.FindAllString(string(logs), -1)
			if str == nil {
				str = []string{}
			}

			mu.Lock()
			pods = append(pods, &apis.Pod{
				Name: pod.Name,
				Log:  str,
			})
			mu.Unlock()
			logrus.Debugf("Processed pod: %s", pod.Name)
		}(pod)
	}
	wg.Wait()

	log.Printf("Completed pod retrieval in namespace %s", namespace)
	return pods, nil
}

func GetNamespaces(client *apis.Client) ([]*apis.Namespace, []*apis.Inspection, error) {
	log.Println("Starting to get namespaces")

	resourceInspections := apis.NewInspections()
	namespaces := apis.NewNamespaces()

	namespaceList, err := client.Clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Printf("Error listing namespaces: %v", err)
		return nil, nil, err
	}

	for _, n := range namespaceList.Items {
		logrus.Debugf("Processing namespace: %s", n.Name)

		var emptyResourceQuota, emptyResource bool

		podList, err := client.Clientset.CoreV1().Pods(n.Name).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Printf("Error listing pods in namespace %s: %v", n.Name, err)
			return nil, nil, err
		}

		serviceList, err := client.Clientset.CoreV1().Services(n.Name).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Printf("Error listing services in namespace %s: %v", n.Name, err)
			return nil, nil, err
		}

		deploymentList, err := client.Clientset.AppsV1().Deployments(n.Name).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Printf("Error listing deployments in namespace %s: %v", n.Name, err)
			return nil, nil, err
		}

		replicaSetList, err := client.Clientset.AppsV1().ReplicaSets(n.Name).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Printf("Error listing replica sets in namespace %s: %v", n.Name, err)
			return nil, nil, err
		}

		statefulSetList, err := client.Clientset.AppsV1().StatefulSets(n.Name).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Printf("Error listing stateful sets in namespace %s: %v", n.Name, err)
			return nil, nil, err
		}

		daemonSetList, err := client.Clientset.AppsV1().DaemonSets(n.Name).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Printf("Error listing daemon sets in namespace %s: %v", n.Name, err)
			return nil, nil, err
		}

		jobList, err := client.Clientset.BatchV1().Jobs(n.Name).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Printf("Error listing jobs in namespace %s: %v", n.Name, err)
			return nil, nil, err
		}

		secretList, err := client.Clientset.CoreV1().Secrets(n.Name).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Printf("Error listing secrets in namespace %s: %v", n.Name, err)
			return nil, nil, err
		}

		configMapList, err := client.Clientset.CoreV1().ConfigMaps(n.Name).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Printf("Error listing config maps in namespace %s: %v", n.Name, err)
			return nil, nil, err
		}

		resourceQuotaList, err := client.Clientset.CoreV1().ResourceQuotas(n.Name).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Printf("Error listing resource quotas in namespace %s: %v", n.Name, err)
			return nil, nil, err
		}

		if len(resourceQuotaList.Items) == 0 {
			emptyResourceQuota = true
			resourceInspections = append(resourceInspections, apis.NewInspection(
				fmt.Sprintf("命名空间 %s 没有设置配额", n.Name),
				"未设置资源配额",
				1,
			))
		}

		totalResources := len(podList.Items) + len(serviceList.Items) + len(deploymentList.Items) +
			len(replicaSetList.Items) + len(statefulSetList.Items) + len(daemonSetList.Items) +
			len(jobList.Items) + len(secretList.Items) + (len(configMapList.Items) - 1)

		if totalResources == 0 {
			emptyResource = true
			resourceInspections = append(resourceInspections, apis.NewInspection(
				fmt.Sprintf("命名空间 %s 下资源为空", n.Name),
				"检查对象为 Pod、Service、Deployment、Replicaset、Statefulset、Daemonset、Job、Secret、ConfigMap",
				1,
			))
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
			ConfigMapCount:     len(configMapList.Items) - 1,
		})

		logrus.Debugf("Processed namespace: %s", n.Name)
	}

	log.Println("Completed namespace retrieval")
	return namespaces, resourceInspections, nil
}

func GetServices(client *apis.Client) ([]*apis.Service, []*apis.Inspection, error) {
	log.Println("Starting to get services")

	resourceInspections := apis.NewInspections()
	services := apis.NewServices()

	serviceList, err := client.Clientset.CoreV1().Services("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Printf("Error listing services: %v", err)
		return nil, nil, err
	}

	for _, s := range serviceList.Items {
		logrus.Debugf("Processing service: %s/%s", s.Namespace, s.Name)
		endpoints, err := client.Clientset.CoreV1().Endpoints(s.Namespace).Get(context.TODO(), s.Name, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				log.Printf("Service %s/%s does not have corresponding endpoints", s.Namespace, s.Name)
				resourceInspections = append(resourceInspections, apis.NewInspection(
					fmt.Sprintf("命名空间 %s 下 Service %s 找不到对应 endpoint", s.Namespace, s.Name),
					"对应的 Endpoints 未找到",
					1,
				))
				continue
			}
			log.Printf("Error getting endpoints for service %s/%s: %v", s.Namespace, s.Name, err)
			return nil, nil, err
		}

		var emptyEndpoints bool
		if len(endpoints.Subsets) == 0 {
			emptyEndpoints = true
			resourceInspections = append(resourceInspections, apis.NewInspection(
				fmt.Sprintf("命名空间 %s 下 Service %s 对应 Endpoints 没有 Subsets", s.Namespace, s.Name),
				"对应的 Endpoints 没有 Subsets",
				1,
			))
		}

		services = append(services, &apis.Service{
			Name:           s.Name,
			Namespace:      s.Namespace,
			EmptyEndpoints: emptyEndpoints,
		})
	}

	log.Println("Completed getting services")
	return services, resourceInspections, nil
}
func GetIngress(client *apis.Client) ([]*apis.Ingress, []*apis.Inspection, error) {
	log.Println("Starting to get ingresses")

	resourceInspections := apis.NewInspections()
	ingress := apis.NewIngress()

	ingressList, err := client.Clientset.NetworkingV1().Ingresses("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Printf("Error listing ingresses: %v", err)
		return nil, nil, err
	}

	ingressMap := make(map[string][]string)
	for _, i := range ingressList.Items {
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
	for key, ingressNames := range ingressMap {
		if len(ingressNames) > 1 {
			for _, ingressName := range ingressNames {
				duplicateIngress[ingressName] = 1
			}
			log.Printf("Found duplicate ingress with same path: %s, Ingress list: %v", key, ingressNames)
		}
	}

	if len(duplicateIngress) > 0 {
		var result []string
		for namespaceName := range duplicateIngress {
			parts := strings.Split(namespaceName, "/")
			for index, i := range ingress {
				if parts[0] == i.Namespace && parts[1] == i.Name {
					ingress[index] = &apis.Ingress{
						Name:          i.Name,
						Namespace:     i.Namespace,
						DuplicatePath: true,
					}
				}
			}

			result = append(result, namespaceName)
		}

		resourceInspections = append(resourceInspections, apis.NewInspection(
			fmt.Sprintf("Ingress %s 存在重复的 Path", strings.Join(result, ", ")),
			"存在重复的路径",
			1,
		))
	}

	log.Println("Completed getting ingresses")
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

func defaultLevel(level int) int {
	if level != 0 {
		return level
	}

	return 2
}
