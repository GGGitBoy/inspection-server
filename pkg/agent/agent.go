package agent

import (
	"bytes"
	"context"
	detector "github.com/rancher/kubernetes-provider-detector"
	detectorProviders "github.com/rancher/kubernetes-provider-detector/providers"
	"github.com/sirupsen/logrus"
	"inspection-server/pkg/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	applyappsv1 "k8s.io/client-go/applyconfigurations/apps/v1"
	applycorev1 "k8s.io/client-go/applyconfigurations/core/v1"
	applyrbacv1 "k8s.io/client-go/applyconfigurations/rbac/v1"
	"k8s.io/client-go/kubernetes"
	"log"
	"os"
	"sigs.k8s.io/yaml"
	"text/template"
)

func Register() error {
	logrus.Infof("register inspection agent\n")
	localKubernetesClient, err := common.GetKubernetesClient(common.LocalCluster)
	if err != nil {
		return err
	}

	clusters, err := localKubernetesClient.DynamicClient.Resource(common.ClusterRes).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, c := range clusters.Items {
		kubernetesClient, err := common.GetKubernetesClient(c.GetName())

		err = CreateAgent(kubernetesClient.Clientset)
		if err != nil {
			return err
		}
	}

	return nil
}

func CreateAgent(clientset *kubernetes.Clientset) error {

	err := ApplyNamespace(clientset)
	if err != nil {
		return err
	}

	err = ApplyServiceAccount(clientset)
	if err != nil {
		return err
	}

	err = ApplyClusterRoleBinding(clientset)
	if err != nil {
		return err
	}

	err = ApplyConfigMap(clientset)
	if err != nil {
		return err
	}

	err = ApplyDaemonSet(clientset)
	if err != nil {
		return err
	}

	return nil
}

func ApplyNamespace(clientset *kubernetes.Clientset) error {
	yamlFile, err := os.ReadFile(common.AgentYamlPath + "namespace.yaml")
	if err != nil {
		return err
	}

	var namespace *applycorev1.NamespaceApplyConfiguration
	err = yaml.Unmarshal(yamlFile, &namespace)
	if err != nil {
		return err
	}

	_, err = clientset.CoreV1().Namespaces().Apply(context.TODO(), namespace, metav1.ApplyOptions{Force: true, FieldManager: "application/apply-patch"})
	if err != nil {
		return err
	}

	return nil
}

func ApplyServiceAccount(clientset *kubernetes.Clientset) error {
	yamlFile, err := os.ReadFile(common.AgentYamlPath + "serviceaccount.yaml")
	if err != nil {
		return err
	}

	var serviceAccount *applycorev1.ServiceAccountApplyConfiguration
	err = yaml.Unmarshal(yamlFile, &serviceAccount)
	if err != nil {
		return err
	}

	_, err = clientset.CoreV1().ServiceAccounts(common.InspectionNamespace).Apply(context.TODO(), serviceAccount, metav1.ApplyOptions{Force: true, FieldManager: "application/apply-patch"})
	if err != nil {
		log.Fatalf("Error applying ServiceAccount: %v", err)
	}

	return nil
}

func ApplyClusterRoleBinding(clientset *kubernetes.Clientset) error {
	yamlFile, err := os.ReadFile(common.AgentYamlPath + "clusterrolebinding.yaml")
	if err != nil {
		return err
	}

	var clusterRoleBinding *applyrbacv1.ClusterRoleBindingApplyConfiguration
	err = yaml.Unmarshal(yamlFile, &clusterRoleBinding)
	if err != nil {
		return err
	}

	_, err = clientset.RbacV1().ClusterRoleBindings().Apply(context.TODO(), clusterRoleBinding, metav1.ApplyOptions{Force: true, FieldManager: "application/apply-patch"})
	if err != nil {
		log.Fatalf("Error applying ClusterRoleBinding: %v", err)
	}

	return nil
}

func ApplyConfigMap(clientset *kubernetes.Clientset) error {
	yamlFile, err := os.ReadFile(common.AgentYamlPath + "configmap.yaml")
	if err != nil {
		return err
	}

	var configMap *applycorev1.ConfigMapApplyConfiguration
	err = yaml.Unmarshal(yamlFile, &configMap)
	if err != nil {
		return err
	}

	_, err = clientset.CoreV1().ConfigMaps(common.InspectionNamespace).Apply(context.TODO(), configMap, metav1.ApplyOptions{Force: true, FieldManager: "application/apply-patch"})
	if err != nil {
		log.Fatalf("Error applying ConfigMap: %v", err)
	}

	return nil
}

func ApplyDaemonSet(clientset *kubernetes.Clientset) error {
	provider, err := detector.DetectProvider(context.TODO(), clientset)
	if err != nil {
		return err
	}

	var setDocker, setContainerd bool
	if provider == detectorProviders.RKE {
		setDocker = true
	} else if provider == detectorProviders.RKE2 {
		setContainerd = true
	} else if provider == detectorProviders.K3s {
		setContainerd = true
	}

	tmpl, err := template.ParseFiles(common.AgentYamlPath + "daemonset.yaml")
	if err != nil {
		return err
	}

	var rendered bytes.Buffer
	err = tmpl.Execute(&rendered, map[string]interface{}{
		"Values": Values{
			SetDocker:     setDocker,
			SetContainerd: setContainerd,
			Provider:      provider,
		},
	})
	if err != nil {
		return err
	}

	var daemonSet *applyappsv1.DaemonSetApplyConfiguration
	err = yaml.Unmarshal(rendered.Bytes(), &daemonSet)
	if err != nil {
		return err
	}

	_, err = clientset.AppsV1().DaemonSets(common.InspectionNamespace).Apply(context.TODO(), daemonSet, metav1.ApplyOptions{Force: true, FieldManager: "application/apply-patch"})
	if err != nil {
		return err
	}

	return nil
}

// Values represents the values to be passed to the template
type Values struct {
	SetDocker     bool
	SetContainerd bool
	Provider      string
}
