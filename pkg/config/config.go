package config

import (
	"context"
	"encoding/json"
	"inspection-server/pkg/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sync"
)

var MutexConfig sync.Mutex

type Config struct {
	Kubernetes map[string]*Kubernetes `json:"kubernetes"`
}

type Kubernetes struct {
	Enable    bool       `json:"enable"`
	Agent     string     `json:"agent"`
	Workloads *Workloads `json:"workloads"`
	Nodes     []*Node    `json:"nodes"`
}

type Workloads struct {
	Deployment  []*WorkloadData `json:"deployment"`
	Statefulset []*WorkloadData `json:"statefulset"`
	Daemonset   []*WorkloadData `json:"daemonset"`
	Job         []*WorkloadData `json:"job"`
	Cronjob     []*WorkloadData `json:"cronjob"`
}

type WorkloadData struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Regexp    string `json:"regexp"`
	Core      bool   `json:"core"`
}

type Node struct {
	Names    []string   `json:"names"`
	Commands []*Command `json:"commands"`
}

type Command struct {
	Description string `json:"description"`
	Command     string `json:"command"`
	Core        bool   `json:"core"`
}

func NewConfig() *Config {
	return &Config{
		Kubernetes: make(map[string]*Kubernetes),
	}
}

func Register() error {
	localKubernetesClient, err := common.GetKubernetesClient(common.LocalCluster)
	if err != nil {
		return err
	}

	clusters, err := localKubernetesClient.DynamicClient.Resource(common.ClusterRes).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	config := NewConfig()
	for _, c := range clusters.Items {
		kubernetesClient, err := common.GetKubernetesClient(c.GetName())
		if err != nil {
			return err
		}

		nodeList, err := kubernetesClient.Clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return err
		}

		var nodeNames []string
		for _, n := range nodeList.Items {
			nodeNames = append(nodeNames, n.GetName())
		}

		workloads := &Workloads{
			Deployment: []*WorkloadData{
				{
					Name:      "cattle-cluster-agent",
					Namespace: "cattle-system",
					Core:      false,
				},
				{
					Name:      "rancher-webhook",
					Namespace: "cattle-system",
					Core:      true,
				},
				{
					Name:      "calico-kube-controllers",
					Namespace: "kube-system",
				},
				{
					Name:      "coredns",
					Namespace: "kube-system",
					Regexp:    "\\[(ERROR|WARNING)\\].*",
				},
				{
					Name:      "coredns-autoscaler",
					Namespace: "kube-system",
				},
				{
					Name:      "calico-kube-controllers",
					Namespace: "metrics-server",
				},
			},
			Daemonset: []*WorkloadData{
				{
					Name:      "inspection-agent",
					Namespace: "cattle-system",
				},
				{
					Name:      "kube-api-auth",
					Namespace: "cattle-system",
				},
				{
					Name:      "cattle-node-agent",
					Namespace: "cattle-system",
				},
				{
					Name:      "nginx-ingress-controller",
					Namespace: "ingress-nginx",
				},
				{
					Name:      "canal",
					Namespace: "kube-system",
				},
			},
		}
		nodes := []*Node{
			{
				Names: nodeNames,
				Commands: []*Command{
					{
						Description: "Kubelet Health Check",
						Command:     "curl -sS http://localhost:10248/healthz",
						Core:        true,
					},
					{
						Description: "API Server Ready Check",
						Command:     "kubectl get --raw='/readyz'",
						Core:        false,
					},
					{
						Description: "Test Error command",
						Command:     "test-error",
						Core:        false,
					},
				},
			},
			{
				Names: []string{},
				Commands: []*Command{
					{
						Description: "Test Error command",
						Command:     "test-error",
					},
				},
			},
		}

		config.Kubernetes[c.GetName()] = &Kubernetes{
			Enable:    true,
			Workloads: workloads,
			Nodes:     nodes,
		}
	}

	err = WriteConfigFile(config)
	if err != nil {
		return err
	}

	return nil
}

func ReadConfigFile() (*Config, error) {
	MutexConfig.Lock()
	defer MutexConfig.Unlock()

	content, err := common.ReadFile(common.ConfigFilePath)
	if err != nil {
		return nil, err
	}

	config := NewConfig()
	err = json.Unmarshal(content, &config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func WriteConfigFile(c *Config) error {
	MutexConfig.Lock()
	defer MutexConfig.Unlock()

	jsonData, err := json.Marshal(c)
	if err != nil {
		return err
	}

	err = common.WriteFile(common.ConfigFilePath, jsonData)
	if err != nil {
		return err
	}

	return nil
}
