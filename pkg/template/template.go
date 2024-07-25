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

		clusterCoreConfig := apis.NewClusterCoreConfig()
		clusterNodeConfig := apis.NewClusterNodeConfig()
		clusterResourceConfig := apis.NewClusterResourceConfig()

		clusterNodeConfig.NodeConfig = []*apis.NodeConfig{
			{
				Names: nodeNames,
				Commands: []*apis.CommandConfig{
					{
						Description: "API Servedr Ready Check",
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
					{
						Description: "Kubelet Health Check",
						Command:     "curl -sS http://localhost:10248/healthz",
					},
					{
						Description: "KubeProxy Health Check",
						Command:     "curl -sS http://localhost:10256/healthz > /dev/null 2>&1 && echo ok || { curl -sS http://localhost:10256/healthz; }",
					},
					{
						Description: "Containerd Health Check",
						Command:     "crictl pods > /dev/null 2>&1 && echo ok || { crictl pods; }",
					},
					{
						Description: "Docker Health Check",
						Command:     "docker ps > /dev/null 2>&1 && echo ok || { docker ps; }",
					},
					{
						Description: "Test Error command",
						Command:     "test-error",
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

		clusterResourceConfig = &apis.ClusterResourceConfig{
			WorkloadConfig: &apis.WorkloadConfig{
				Deployment: []*apis.WorkloadDetailConfig{
					{
						Name:      "cattle-cluster-agent",
						Namespace: "cattle-system",
					},
					{
						Name:      "rancher-webhook",
						Namespace: "cattle-system",
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
			},
			NamespaceConfig: &apis.NamespaceConfig{
				Enable: true,
			},
			ServiceConfig: &apis.ServiceConfig{
				Enable: true,
			},
			IngressConfig: &apis.IngressConfig{
				Enable: true,
			},
		}

		spec, _, err := unstructured.NestedMap(c.UnstructuredContent(), "spec")
		if err != nil {
			log.Fatalf("Error getting spec: %v", err)
		}

		kubernetesConfig = append(kubernetesConfig, &apis.KubernetesConfig{
			Enable:                true,
			ClusterID:             c.GetName(),
			ClusterName:           spec["displayName"].(string),
			ClusterCoreConfig:     clusterCoreConfig,
			ClusterNodeConfig:     clusterNodeConfig,
			ClusterResourceConfig: clusterResourceConfig,
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
