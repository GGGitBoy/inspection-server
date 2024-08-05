package api

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"inspection-server/pkg/common"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
)

func ListAgent() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		localKubernetesClient, err := common.GetKubernetesClient(common.LocalCluster)
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		clusterList, err := localKubernetesClient.DynamicClient.Resource(common.ClusterRes).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		var listAgent []string
		for _, c := range clusterList.Items {
			kubernetesClient, err := common.GetKubernetesClient(c.GetName())
			_, err = kubernetesClient.Clientset.AppsV1().DaemonSets(common.InspectionNamespace).Get(context.TODO(), common.AgentName, metav1.GetOptions{})
			if err != nil && !k8serrors.IsNotFound(err) {
				logrus.Errorf("Could not get cluster %s inspection-agent : %v\n", c.GetName(), err)
				continue
			}

			listAgent = append(listAgent, c.GetName())
		}

		jsonData, err := json.MarshalIndent(listAgent, "", "\t")
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
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
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		err = kubernetesClient.Clientset.AppsV1().DaemonSets(common.InspectionNamespace).Delete(context.TODO(), common.AgentName, metav1.DeleteOptions{})
		if err != nil && !k8serrors.IsNotFound(err) {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		err = kubernetesClient.Clientset.CoreV1().ConfigMaps(common.InspectionNamespace).Delete(context.TODO(), common.AgentScriptName, metav1.DeleteOptions{})
		if err != nil && !k8serrors.IsNotFound(err) {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		err = kubernetesClient.Clientset.RbacV1().ClusterRoleBindings().Delete(context.TODO(), common.AgentName, metav1.DeleteOptions{})
		if err != nil && !k8serrors.IsNotFound(err) {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		err = kubernetesClient.Clientset.CoreV1().ServiceAccounts(common.InspectionNamespace).Delete(context.TODO(), common.AgentName, metav1.DeleteOptions{})
		if err != nil && !k8serrors.IsNotFound(err) {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		rw.Write([]byte("删除完成"))
	})
}
