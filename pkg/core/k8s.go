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
	"os"
	"regexp"
	"sync"
	"time"
)

func GetGlobal(report *apis.Report) error {
	report.Global.Warnings = []apis.Warning{
		{Title: "红色警告", Message: "红色：影响 rancher 或者 集群运行的", Type: 3},
		{Title: "黄色警告", Message: "黄色：不影响 rancher 或者 集群运行的，但是有风险的", Type: 2},
		{Title: "灰色警告", Message: "灰色：没有太大风险的，不过最好也可以处理的", Type: 1},
	}

	var rating int
	for _, w := range report.Global.Warnings {
		if w.Type > rating {
			rating = w.Type
		}
	}

	report.Global.ReportTime = time.Now().Format(time.DateTime)
	report.Global.Rating = rating

	return nil
}

func GetNodes(client *apis.Client) ([]*apis.Node, error) {
	var nodeArray []*apis.Node

	set := labels.Set(map[string]string{"name": "inspection-agent"})
	podList, err := client.Clientset.CoreV1().Pods(common.InspectionNamespace).List(context.TODO(), metav1.ListOptions{LabelSelector: set.String()})
	if err != nil {
		return nil, err
	}

	for _, pod := range podList.Items {
		node, err := client.Clientset.CoreV1().Nodes().Get(context.TODO(), pod.Spec.NodeName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		annotations := map[string]string{
			"pod-limits":   node.Annotations["management.cattle.io/pod-limits"],
			"pod-requests": node.Annotations["management.cattle.io/pod-requests"],
		}

		command := "/opt/inspection.sh"
		stdout, stderr, err := ExecToPodThroughAPI(client.Clientset, client.Config, command, pod.Namespace, pod.Name, "inspection-agent-container")
		if err != nil {
			return nil, err
		}

		var results []apis.CommandCheckResult
		err = json.Unmarshal([]byte(stdout), &results)
		if err != nil {
			return nil, err
		}

		nodeArray = append(nodeArray, &apis.Node{
			Name:        pod.Spec.NodeName,
			Annotations: annotations,
			Commands: &apis.Command{
				Stdout: results,
				Stderr: stderr,
			},
		})
	}

	return nodeArray, nil
}

func ExecToPodThroughAPI(clientset *kubernetes.Clientset, config *rest.Config, command string, namespace string, podName string, containerName string) (string, string, error) {
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

func GetWorkloads(name string, client *apis.Client) (*apis.Workload, error) {
	deploymentArray := apis.NewWorkloadDatas()
	daemonSetArray := apis.NewWorkloadDatas()
	statefulSetArray := apis.NewWorkloadDatas()
	jobArray := apis.NewWorkloadDatas()
	//cronJobArray := apis.NewWorkloadDatas()

	globalConfig, err := config.ReadConfigFile()
	if err != nil {
		return nil, err
	}

	if globalConfig.Kubernetes[name].Enable {
		for _, deploy := range globalConfig.Kubernetes[name].Workloads.Deployment {
			deployment, err := client.Clientset.AppsV1().Deployments(deploy.Namespace).Get(context.TODO(), deploy.Name, metav1.GetOptions{})
			if err != nil {
				if k8serrors.IsNotFound(err) {
					continue
				}
				return nil, err
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
				return nil, err
			}

			deploymentArray = append(deploymentArray, &apis.WorkloadData{
				Name:      deployment.Name,
				Namespace: deployment.Namespace,
				Pods:      pods,
				Status: &apis.Status{
					State:     state,
					Condition: condition,
				},
			})
		}

		for _, ds := range globalConfig.Kubernetes[name].Workloads.Daemonset {
			daemonSet, err := client.Clientset.AppsV1().DaemonSets(ds.Namespace).Get(context.TODO(), ds.Name, metav1.GetOptions{})
			if err != nil {
				if k8serrors.IsNotFound(err) {
					continue
				}
				return nil, err
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
				return nil, err
			}

			daemonSetArray = append(daemonSetArray, &apis.WorkloadData{
				Name:      daemonSet.Name,
				Namespace: daemonSet.Namespace,
				Pods:      pods,
				Status: &apis.Status{
					State:     "active",
					Condition: condition,
				},
			})
		}

		for _, sts := range globalConfig.Kubernetes[name].Workloads.Statefulset {
			statefulset, err := client.Clientset.AppsV1().StatefulSets(sts.Namespace).Get(context.TODO(), sts.Name, metav1.GetOptions{})
			if err != nil {
				if k8serrors.IsNotFound(err) {
					continue
				}
				return nil, err
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
				return nil, err
			}

			statefulSetArray = append(statefulSetArray, &apis.WorkloadData{
				Name:      statefulset.Name,
				Namespace: statefulset.Namespace,
				Pods:      pods,
				Status: &apis.Status{
					State:     state,
					Condition: condition,
				},
			})
		}

		for _, j := range globalConfig.Kubernetes[name].Workloads.Job {
			job, err := client.Clientset.BatchV1().Jobs(j.Namespace).Get(context.TODO(), j.Name, metav1.GetOptions{})
			if err != nil {
				if k8serrors.IsNotFound(err) {
					continue
				}
				return nil, err
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
				return nil, err
			}

			jobArray = append(jobArray, &apis.WorkloadData{
				Name:      job.Name,
				Namespace: job.Namespace,
				Pods:      pods,
				Status: &apis.Status{
					State:     state,
					Condition: condition,
				},
			})
		}
	}

	return &apis.Workload{
		Deployment:  deploymentArray,
		Daemonset:   daemonSetArray,
		Statefulset: statefulSetArray,
		Job:         jobArray,
		//Cronjob: cronJobArray,
	}, nil
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
