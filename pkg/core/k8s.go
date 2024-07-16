package core

import (
	"context"
	"encoding/json"
	"fmt"
	"inspection-server/pkg/apis"
	"inspection-server/pkg/common"
	"inspection-server/pkg/config"
	"io"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/utils/strings/slices"
	"os"
	"regexp"
	"sync"
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

func GetNodes(name string, client *apis.Client) ([]*apis.Node, []*apis.Node, []*apis.Inspection, []*apis.Inspection, error) {
	coreNodeArray := apis.NewNodes()
	nodeNodeArray := apis.NewNodes()

	coreInspections := apis.NewInspections()
	nodeInspections := apis.NewInspections()

	globalConfig, err := config.ReadConfigFile()
	if err != nil {
		return nil, nil, nil, nil, err
	}

	set := labels.Set(map[string]string{"name": "inspection-agent"})
	podList, err := client.Clientset.CoreV1().Pods(common.InspectionNamespace).List(context.TODO(), metav1.ListOptions{LabelSelector: set.String()})
	if err != nil {
		return nil, nil, nil, nil, err
	}

	for _, pod := range podList.Items {
		for _, n := range globalConfig.Kubernetes[name].Nodes {
			if slices.Contains(n.Names, pod.Spec.NodeName) {
				node, err := client.Clientset.CoreV1().Nodes().Get(context.TODO(), pod.Spec.NodeName, metav1.GetOptions{})
				if err != nil {
					return nil, nil, nil, nil, err
				}

				annotations := map[string]string{
					"pod-limits":   node.Annotations["management.cattle.io/pod-limits"],
					"pod-requests": node.Annotations["management.cattle.io/pod-requests"],
				}

				var commands []string
				for _, c := range n.Commands {
					commands = append(commands, c.Description+": "+c.Command)
				}

				command := "/opt/inspection.sh"
				stdout, stderr, err := ExecToPodThroughAPI(client.Clientset, client.Config, command, commands, pod.Namespace, pod.Name, "inspection-agent-container")
				if err != nil {
					return nil, nil, nil, nil, err
				}

				var results []apis.CommandCheckResult
				err = json.Unmarshal([]byte(stdout), &results)
				if err != nil {
					return nil, nil, nil, nil, err
				}

				nodeData := &apis.Node{
					Name:        pod.Spec.NodeName,
					Annotations: annotations,
					Commands: &apis.Command{
						Stdout: results,
						Stderr: stderr,
					},
				}

				//if n.Core {
				//	coreNodeArray = append(coreNodeArray, nodeData)
				//} else {
				//	nodeNodeArray =  append(nodeNodeArray, nodeData)
				//}
				nodeNodeArray = append(nodeNodeArray, nodeData)
				nodeInspections = append(nodeInspections, apis.NewInspection("node Inspection", "node Inspection message", "Node", 3, true))
			}
		}
	}

	return coreNodeArray, nodeNodeArray, coreInspections, nodeInspections, nil
}

func ExecToPodThroughAPI(clientset *kubernetes.Clientset, config *rest.Config, command string, commands []string, namespace string, podName string, containerName string) (string, string, error) {
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

func GetWorkloads(name string, client *apis.Client) (*apis.Workload, *apis.Workload, []*apis.Inspection, []*apis.Inspection, error) {
	CoreWorkloadArray := apis.NewWorkload()
	ResourceWorkloadArray := apis.NewWorkload()

	coreInspections := apis.NewInspections()
	resourceInspections := apis.NewInspections()

	globalConfig, err := config.ReadConfigFile()
	if err != nil {
		return nil, nil, nil, nil, err
	}

	if globalConfig.Kubernetes[name].Enable {
		for _, deploy := range globalConfig.Kubernetes[name].Workloads.Deployment {
			deployment, err := client.Clientset.AppsV1().Deployments(deploy.Namespace).Get(context.TODO(), deploy.Name, metav1.GetOptions{})
			if err != nil {
				if k8serrors.IsNotFound(err) {
					continue
				}
				return nil, nil, nil, nil, err
			}

			state := "active"
			var condition []apis.Condition
			for _, c := range deployment.Status.Conditions {
				if c.Status != "True" {
					state = "deactive"
				}
				condition = append(condition, apis.Condition{
					Type:   string(c.Type),
					Status: string(c.Status),
					Reason: c.Reason,
				})
			}

			set := labels.Set(deployment.Spec.Selector.MatchLabels)
			pods, err := GetPod(deploy.Regexp, deployment.Namespace, set, client.Clientset)
			if err != nil {
				return nil, nil, nil, nil, err
			}

			deploymentData := &apis.WorkloadData{
				Name:      deployment.Name,
				Namespace: deployment.Namespace,
				Pods:      pods,
				Status: &apis.Status{
					State:     state,
					Condition: condition,
				},
			}

			if deploy.Core {
				CoreWorkloadArray.Deployment = append(CoreWorkloadArray.Deployment, deploymentData)
				coreInspections = append(coreInspections, apis.NewInspection("core deployment", "core deployment message", "Deployment", 3, true))
			} else {
				ResourceWorkloadArray.Deployment = append(ResourceWorkloadArray.Deployment, deploymentData)
				resourceInspections = append(resourceInspections, apis.NewInspection("resource deployment", "resource deployment message", "Deployment", 1, true))
			}
		}

		for _, ds := range globalConfig.Kubernetes[name].Workloads.Daemonset {
			daemonSet, err := client.Clientset.AppsV1().DaemonSets(ds.Namespace).Get(context.TODO(), ds.Name, metav1.GetOptions{})
			if err != nil {
				if k8serrors.IsNotFound(err) {
					continue
				}
				return nil, nil, nil, nil, err
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
				return nil, nil, nil, nil, err
			}

			daemonSetData := &apis.WorkloadData{
				Name:      daemonSet.Name,
				Namespace: daemonSet.Namespace,
				Pods:      pods,
				Status: &apis.Status{
					State:     "active",
					Condition: condition,
				},
			}

			if ds.Core {
				CoreWorkloadArray.Daemonset = append(CoreWorkloadArray.Daemonset, daemonSetData)
			} else {
				ResourceWorkloadArray.Daemonset = append(ResourceWorkloadArray.Daemonset, daemonSetData)
			}
		}

		for _, sts := range globalConfig.Kubernetes[name].Workloads.Statefulset {
			statefulset, err := client.Clientset.AppsV1().StatefulSets(sts.Namespace).Get(context.TODO(), sts.Name, metav1.GetOptions{})
			if err != nil {
				if k8serrors.IsNotFound(err) {
					continue
				}
				return nil, nil, nil, nil, err
			}

			state := "active"
			var condition []apis.Condition
			for _, c := range statefulset.Status.Conditions {
				if c.Status != "True" {
					state = "deactive"
				}
				condition = append(condition, apis.Condition{
					Type:   string(c.Type),
					Status: string(c.Status),
					Reason: c.Reason,
				})
			}

			set := labels.Set(statefulset.Spec.Selector.MatchLabels)
			pods, err := GetPod(sts.Regexp, statefulset.Namespace, set, client.Clientset)
			if err != nil {
				return nil, nil, nil, nil, err
			}

			statefulSetData := &apis.WorkloadData{
				Name:      statefulset.Name,
				Namespace: statefulset.Namespace,
				Pods:      pods,
				Status: &apis.Status{
					State:     state,
					Condition: condition,
				},
			}

			if sts.Core {
				CoreWorkloadArray.Statefulset = append(CoreWorkloadArray.Statefulset, statefulSetData)
			} else {
				ResourceWorkloadArray.Statefulset = append(ResourceWorkloadArray.Statefulset, statefulSetData)
			}
		}

		for _, j := range globalConfig.Kubernetes[name].Workloads.Job {
			job, err := client.Clientset.BatchV1().Jobs(j.Namespace).Get(context.TODO(), j.Name, metav1.GetOptions{})
			if err != nil {
				if k8serrors.IsNotFound(err) {
					continue
				}
				return nil, nil, nil, nil, err
			}

			state := "active"
			var condition []apis.Condition
			for _, c := range job.Status.Conditions {
				if c.Status != "True" {
					state = "deactive"
				}
				condition = append(condition, apis.Condition{
					Type:   string(c.Type),
					Status: string(c.Status),
					Reason: c.Reason,
				})
			}

			set := labels.Set(job.Spec.Selector.MatchLabels)
			pods, err := GetPod(j.Regexp, j.Namespace, set, client.Clientset)
			if err != nil {
				return nil, nil, nil, nil, err
			}

			jobData := &apis.WorkloadData{
				Name:      job.Name,
				Namespace: job.Namespace,
				Pods:      pods,
				Status: &apis.Status{
					State:     state,
					Condition: condition,
				},
			}

			if j.Core {
				CoreWorkloadArray.Job = append(CoreWorkloadArray.Job, jobData)
			} else {
				ResourceWorkloadArray.Job = append(ResourceWorkloadArray.Job, jobData)
			}
		}
	}

	return CoreWorkloadArray, ResourceWorkloadArray, coreInspections, resourceInspections, nil
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
			defer wg.Done()

			getLog := clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{TailLines: &line})
			podLogs, err := getLog.Stream(context.TODO())
			if err != nil {
				fmt.Errorf("Error getting logs for pod %s: %v\n", pod.Name, err)
				return
			}
			defer podLogs.Close()

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
			mu.Lock()
			pods = append(pods, &apis.Pod{
				Name: pod.Name,
				Log:  str,
			})
			mu.Unlock()
		}(pod)
	}
	wg.Wait()

	return pods, nil
}

func GetNamespaces(name string, client *apis.Client) ([]*apis.Namespace, []*apis.Inspection, error) {
	resourceInspections := apis.NewInspections()

	namespaces := apis.NewNamespaces()
	namespaces = append(namespaces, &apis.Namespace{
		Name:               "default",
		EmptyResourceQuota: false,
	})
	resourceInspections = append(resourceInspections, apis.NewInspection("Namespace Inspection", "Namespace Inspection message", "Namespace", 1, true))
	return namespaces, resourceInspections, nil
}

func GetPersistentVolumeClaims(name string, client *apis.Client) ([]*apis.PersistentVolumeClaim, []*apis.Inspection, error) {
	resourceInspections := apis.NewInspections()

	persistentVolumeClaims := apis.NewPersistentVolumeClaims()
	persistentVolumeClaims = append(persistentVolumeClaims, &apis.PersistentVolumeClaim{
		Name:  "default",
		State: "bound",
	})
	resourceInspections = append(resourceInspections, apis.NewInspection("PersistentVolumeClaim Inspection", "PersistentVolumeClaim Inspection message", "PersistentVolumeClaim", 1, true))
	return persistentVolumeClaims, resourceInspections, nil
}

func GetServices(name string, client *apis.Client) ([]*apis.Service, []*apis.Inspection, error) {
	resourceInspections := apis.NewInspections()

	services := apis.NewServices()
	services = append(services, &apis.Service{
		Name:           "default",
		Namespace:      "default",
		EmptyEndpoints: false,
	})
	resourceInspections = append(resourceInspections, apis.NewInspection("Service Inspection", "Service Inspection message", "Service", 1, true))
	return services, resourceInspections, nil
}

func GetIngress(name string, client *apis.Client) ([]*apis.Ingress, []*apis.Inspection, error) {
	resourceInspections := apis.NewInspections()

	ingress := apis.NewIngress()
	ingress = append(ingress, &apis.Ingress{
		Name:          "default",
		DuplicatePath: false,
	})
	resourceInspections = append(resourceInspections, apis.NewInspection("Ingress Inspection", "Ingress Inspection message", "Ingress", 1, true))
	return ingress, resourceInspections, nil
}
