package api

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"inspection-server/pkg/common"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"net/http"
)

func ListAgent() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		localKubernetesClient, err := common.GetKubernetesClient(common.LocalCluster)
		if err != nil {
			log.Fatal(err)
		}

		clusters, err := localKubernetesClient.DynamicClient.Resource(common.ClusterRes).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Fatal(err)
		}

		var listAgent []string
		for _, c := range clusters.Items {
			kubernetesClient, err := common.GetKubernetesClient(c.GetName())

			_, err = kubernetesClient.Clientset.AppsV1().DaemonSets(common.InspectionNamespace).Get(context.TODO(), "inspection-agent", metav1.GetOptions{})
			if err != nil && !k8serrors.IsNotFound(err) {
				log.Fatal(err)
			}

			listAgent = append(listAgent, c.GetName())
		}

		jsonData, err := json.MarshalIndent(listAgent, "", "\t")
		if err != nil {
			return
		}

		rw.Write(jsonData)
	})
}

func DeleteAgent() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		clusterID := vars["id"]

		kubernetesClient, err := common.GetKubernetesClient(clusterID)
		if err != nil {
			log.Fatal(err)
		}

		err = kubernetesClient.Clientset.AppsV1().DaemonSets(common.InspectionNamespace).Delete(context.TODO(), "inspection-agent", metav1.DeleteOptions{})
		if err != nil && !k8serrors.IsNotFound(err) {
			log.Fatal(err)
		}

		err = kubernetesClient.Clientset.CoreV1().ConfigMaps(common.InspectionNamespace).Delete(context.TODO(), "inspection-agent-sh", metav1.DeleteOptions{})
		if err != nil && !k8serrors.IsNotFound(err) {
			log.Fatal(err)
		}

		err = kubernetesClient.Clientset.RbacV1().ClusterRoleBindings().Delete(context.TODO(), "inspection-agent", metav1.DeleteOptions{})
		if err != nil && !k8serrors.IsNotFound(err) {
			log.Fatal(err)
		}

		err = kubernetesClient.Clientset.CoreV1().ServiceAccounts(common.InspectionNamespace).Delete(context.TODO(), "inspection-agent", metav1.DeleteOptions{})
		if err != nil && !k8serrors.IsNotFound(err) {
			log.Fatal(err)
		}
	})
}
