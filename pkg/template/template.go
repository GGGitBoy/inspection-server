package template

import (
	"context"
	"inspection-server/pkg/apis"
	"inspection-server/pkg/common"
	"inspection-server/pkg/db"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"log"
)

func Register() error {
	localKubernetesClient, err := common.GetKubernetesClient(common.LocalCluster)
	if err != nil {
		return err
	}

	clusters, err := localKubernetesClient.DynamicClient.Resource(common.ClusterRes).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	template := apis.NewTemplate()
	kubernetesConfig := apis.NewKubernetesConfig()
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

		workloadConfig := &apis.WorkloadConfig{
			Deployment: []*apis.WorkloadDetailConfig{
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
			Daemonset: []*apis.WorkloadDetailConfig{
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
		nodeConfig := []*apis.NodeConfig{
			{
				Names: nodeNames,
				Commands: []*apis.CommandConfig{
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
				Commands: []*apis.CommandConfig{
					{
						Description: "Test Error command",
						Command:     "test-error",
					},
				},
			},
		}

		spec, _, err := unstructured.NestedMap(c.UnstructuredContent(), "spec")
		if err != nil {
			log.Fatalf("Error getting spec: %v", err)
		}

		kubernetesConfig = append(kubernetesConfig, &apis.KubernetesConfig{
			Enable:         true,
			ClusterID:      c.GetName(),
			ClusterName:    spec["displayName"].(string),
			WorkloadConfig: workloadConfig,
			NodeConfig:     nodeConfig,
		})
	}

	template = &apis.Template{
		ID:               common.GetUUID(),
		Name:             "Default",
		KubernetesConfig: kubernetesConfig,
	}

	err = db.CreateTemplate(template)
	if err != nil {
		return err
	}

	return nil
}

//func ReadConfigFile() (*Config, error) {
//	MutexConfig.Lock()
//	defer MutexConfig.Unlock()
//
//	content, err := common.ReadFile(common.ConfigFilePath)
//	if err != nil {
//		return nil, err
//	}
//
//	config := NewConfig()
//	err = json.Unmarshal(content, &config)
//	if err != nil {
//		return nil, err
//	}
//
//	return config, nil
//}
//
//func WriteConfigFile(c *Config) error {
//	MutexConfig.Lock()
//	defer MutexConfig.Unlock()
//
//	jsonData, err := json.Marshal(c)
//	if err != nil {
//		return err
//	}
//
//	err = common.WriteFile(common.ConfigFilePath, jsonData)
//	if err != nil {
//		return err
//	}
//
//	return nil
//}
