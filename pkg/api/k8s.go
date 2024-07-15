package api

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"inspection-server/pkg/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"net/http"
)

type Resource struct {
	Nodes        []string `json:"nodes"`
	Deployments  []*Data  `json:"deployments"`
	Statefulsets []*Data  `json:"statefulsets"`
	Daemonsets   []*Data  `json:"daemonsets"`
	Jobs         []*Data  `json:"jobs"`
	Cronjobs     []*Data  `json:"cronjobs"`
}

type Data struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

func NewResource() *Resource {
	return &Resource{
		Nodes:        []string{},
		Deployments:  []*Data{},
		Statefulsets: []*Data{},
		Daemonsets:   []*Data{},
		Jobs:         []*Data{},
		Cronjobs:     []*Data{},
	}
}

func GetClusters() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		var clusterNames []string

		localKubernetesClient, err := common.GetKubernetesClient(common.LocalCluster)
		if err != nil {
			log.Fatal(err)
		}

		clusters, err := localKubernetesClient.DynamicClient.Resource(common.ClusterRes).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Fatal(err)
		}

		for _, c := range clusters.Items {
			clusterNames = append(clusterNames, c.GetName())
		}

		jsonData, err := json.MarshalIndent(clusterNames, "", "\t")
		if err != nil {
			log.Fatal(err)
		}

		rw.Write(jsonData)
	})
}

func GetResource() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		resource := NewResource()

		vars := mux.Vars(req)
		clusterID := vars["id"]

		kubernetesClient, err := common.GetKubernetesClient(clusterID)
		if err != nil {
			log.Fatal(err)
		}

		nodes, err := kubernetesClient.Clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Fatal(err)
		}

		for _, r := range nodes.Items {
			resource.Nodes = append(resource.Nodes, r.GetName())
		}

		deployments, err := kubernetesClient.Clientset.AppsV1().Deployments("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Fatal(err)
		}

		for _, r := range deployments.Items {
			resource.Deployments = append(resource.Deployments, &Data{
				Name:      r.GetName(),
				Namespace: r.GetNamespace(),
			})
		}

		daemonSets, err := kubernetesClient.Clientset.AppsV1().DaemonSets("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Fatal(err)
		}

		for _, r := range daemonSets.Items {
			resource.Daemonsets = append(resource.Daemonsets, &Data{
				Name:      r.GetName(),
				Namespace: r.GetNamespace(),
			})
		}

		statefulSets, err := kubernetesClient.Clientset.AppsV1().StatefulSets("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Fatal(err)
		}

		for _, r := range statefulSets.Items {
			resource.Statefulsets = append(resource.Statefulsets, &Data{
				Name:      r.GetName(),
				Namespace: r.GetNamespace(),
			})
		}

		jobs, err := kubernetesClient.Clientset.BatchV1().Jobs("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Fatal(err)
		}

		for _, r := range jobs.Items {
			resource.Jobs = append(resource.Jobs, &Data{
				Name:      r.GetName(),
				Namespace: r.GetNamespace(),
			})
		}

		cronJobs, err := kubernetesClient.Clientset.BatchV1().CronJobs("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Fatal(err)
		}

		for _, r := range cronJobs.Items {
			resource.Cronjobs = append(resource.Cronjobs, &Data{
				Name:      r.GetName(),
				Namespace: r.GetNamespace(),
			})
		}

		jsonData, err := json.MarshalIndent(resource, "", "\t")
		if err != nil {
			log.Fatal(err)
		}

		rw.Write(jsonData)
	})
}
