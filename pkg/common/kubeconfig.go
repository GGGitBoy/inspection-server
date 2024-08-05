package common

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"inspection-server/pkg/apis"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"net/http"
	"strings"
	"time"
)

var (
	ClusterRes = schema.GroupVersionResource{
		Group:    "management.cattle.io",
		Version:  "v3",
		Resource: "clusters",
	}
)

type KubeConfig struct {
	BaseType string `json:"baseType"`
	Config   string `json:"config"`
	Type     string `json:"apis"`
}

func getClient() http.Client {
	return http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: 20 * time.Second,
	}
}

func getAuthHeader() http.Header {
	authHeader := http.Header{}
	authHeader.Add("Authorization", "Bearer "+BearerToken)
	return authHeader
}

func GenerateKubeconfig(clients map[string]*apis.Client) error {
	localKubernetesClien, err := GetKubernetesClient(LocalCluster)
	if err != nil {
		return err
	}

	clusters, err := localKubernetesClien.DynamicClient.Resource(ClusterRes).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, c := range clusters.Items {
		kubernetesClient, err := GetKubernetesClient(c.GetName())
		if err != nil {
			return err
		}

		clients[c.GetName()] = &apis.Client{
			DynamicClient: kubernetesClient.DynamicClient,
			Clientset:     kubernetesClient.Clientset,
			Config:        kubernetesClient.Config,
		}
	}

	return nil
}

func GetKubernetesClient(name string) (*apis.Client, error) {
	err := WriteKubeconfig(name)
	if err != nil {
		return nil, err
	}

	kubernetesClient, err := GetClient(name)
	if err != nil {
		return nil, err
	}

	return kubernetesClient, nil
}

func WriteKubeconfig(clusterID string) error {
	if FileExists(WriteKubeconfigPath + clusterID) {
		logrus.Infof("cluster %s kubeconfig already exists\n", clusterID)
		return nil
	}

	client := getClient()
	req, err := http.NewRequest(http.MethodPost, ServerURL+"/v3/clusters/"+clusterID+"?action=generateKubeconfig", strings.NewReader(""))
	if err != nil {
		return err
	}
	req.Header = getAuthHeader()

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("GET %s expect 200 status but got %d", req.URL.String(), resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var kubeConfig KubeConfig
	err = json.Unmarshal(body, &kubeConfig)
	if err != nil {
		return err
	}

	err = WriteFile(WriteKubeconfigPath+clusterID, []byte(kubeConfig.Config))
	if err != nil {
		return err
	}

	return err
}

func GetClient(kubeconfigPath string) (*apis.Client, error) {
	config, err := clientcmd.BuildConfigFromFlags("", WriteKubeconfigPath+kubeconfigPath)
	if err != nil {
		return nil, err
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &apis.Client{
		DynamicClient: dynamicClient,
		Clientset:     clientset,
		Config:        config,
	}, err
}
