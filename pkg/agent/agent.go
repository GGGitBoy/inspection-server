package agent

import (
	"context"
	"fmt"
	"inspection-server/pkg/common"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	applyappsv1 "k8s.io/client-go/applyconfigurations/apps/v1"
	applycorev1 "k8s.io/client-go/applyconfigurations/core/v1"
	applymetav1 "k8s.io/client-go/applyconfigurations/meta/v1"
	applyrbacv1 "k8s.io/client-go/applyconfigurations/rbac/v1"
	"k8s.io/client-go/kubernetes"
	"log"
)

func SyncAgent() error {
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

func DeleteAgent(clientset *kubernetes.Clientset) error {
	err := clientset.AppsV1().DaemonSets(common.InspectionNamespace).Delete(context.TODO(), "inspection-agent", metav1.DeleteOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return err
	}

	return nil
}

func CreateAgent(clientset *kubernetes.Clientset) error {
	namespace := getNamespace()
	_, err := clientset.CoreV1().Namespaces().Apply(context.TODO(), namespace, metav1.ApplyOptions{Force: true, FieldManager: "application/apply-patch"})
	if err != nil {
		log.Fatalf("Error applying ServiceAccount: %v", err)
	}

	serviceAccount := getServiceAccount()
	_, err = clientset.CoreV1().ServiceAccounts(common.InspectionNamespace).Apply(context.TODO(), serviceAccount, metav1.ApplyOptions{Force: true, FieldManager: "application/apply-patch"})
	if err != nil {
		log.Fatalf("Error applying ServiceAccount: %v", err)
	}

	clusterRoleBinding := getClusterRoleBinding()
	_, err = clientset.RbacV1().ClusterRoleBindings().Apply(context.TODO(), clusterRoleBinding, metav1.ApplyOptions{Force: true, FieldManager: "application/apply-patch"})
	if err != nil {
		log.Fatalf("Error applying ClusterRoleBinding: %v", err)
	}

	configMap := getConfigMap()
	_, err = clientset.CoreV1().ConfigMaps(common.InspectionNamespace).Apply(context.TODO(), configMap, metav1.ApplyOptions{Force: true, FieldManager: "application/apply-patch"})
	if err != nil {
		log.Fatalf("Error applying ConfigMap: %v", err)
	}

	daemonSet := getDaemonSet()
	result, err := clientset.AppsV1().DaemonSets(common.InspectionNamespace).Apply(context.TODO(), daemonSet, metav1.ApplyOptions{Force: true, FieldManager: "application/apply-patch"})
	if err != nil {
		log.Fatalf("Error creating DaemonSet: %s", err.Error())
	}

	fmt.Printf("Created DaemonSet %q.\n", result.GetObjectMeta().GetName())

	return nil
}

func Register() error {
	err := SyncAgent()
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func getNamespace() *applycorev1.NamespaceApplyConfiguration {
	return applycorev1.Namespace(common.InspectionNamespace)
}

func getServiceAccount() *applycorev1.ServiceAccountApplyConfiguration {
	return applycorev1.ServiceAccount("inspection-agent", common.InspectionNamespace)
}

func getClusterRoleBinding() *applyrbacv1.ClusterRoleBindingApplyConfiguration {
	return applyrbacv1.ClusterRoleBinding("inspection-agent").
		WithRoleRef(applyrbacv1.RoleRef().
			WithAPIGroup("rbac.authorization.k8s.io").
			WithKind("ClusterRole").
			WithName("cluster-admin")).
		WithSubjects(
			applyrbacv1.Subject().
				WithKind("ServiceAccount").
				WithName("inspection-agent").
				WithNamespace(common.InspectionNamespace),
		)
}

func getConfigMap() *applycorev1.ConfigMapApplyConfiguration {
	return applycorev1.ConfigMap("inspection-agent-sh", common.InspectionNamespace).
		WithData(map[string]string{
			"inspection.sh": `#!/bin/bash

commands=("$@")

results=()

for command in "${commands[@]}"; do
	description=$(echo "$command" | cut -d ':' -f 1)
	command_desc=$(echo "$command" | cut -d ':' -f 2-)

	result=$(eval "${command_desc}" 2>&1)
	status=$?

	if [ $status -ne 0 ]; then
		result="{\"description\": \"${description}\", \"command\": \"${command_desc}\", \"error\": \"$(echo "$result" | tail -n 1)\"}"
	else
		result="{\"description\": \"${description}\", \"command\": \"${command_desc}\", \"response\": \"${result}\"}"
	fi

	results+=("$result")
done

echo -n "["
for ((i=0; i<${#results[@]}; i++)); do
	echo -n "${results[i]}"
	if [ $i -lt $((${#results[@]} - 1)) ]; then
		echo -n ","
	fi
done
echo -n "]"
`,
		})
}

func getDaemonSet() *applyappsv1.DaemonSetApplyConfiguration {
	return applyappsv1.DaemonSet("inspection-agent", common.InspectionNamespace).
		WithSpec(applyappsv1.DaemonSetSpec().
			WithSelector(applymetav1.LabelSelector().
				WithMatchLabels(map[string]string{"name": "inspection-agent"})).
			WithTemplate(applycorev1.PodTemplateSpec().
				WithLabels(map[string]string{"name": "inspection-agent"}).
				WithSpec(applycorev1.PodSpec().
					WithContainers(applycorev1.Container().
						WithName("inspection-agent-container").
						//WithImage("dockerrrboy/inspection-agent").
						WithImage("alpine:latest").
						WithSecurityContext(
							applycorev1.SecurityContext().
								WithAllowPrivilegeEscalation(true).
								WithPrivileged(true)).
						WithStdin(true).
						WithTTY(false).
						WithVolumeMounts(
							applycorev1.VolumeMount().
								WithName("inspection").
								WithMountPath("/inspection"),
							//applycorev1.VolumeMount().
							//	WithName("docker-socket").
							//	WithMountPath("/var/run/docker.sock"),
							//applycorev1.VolumeMount().
							//	WithName("docker-bin").
							//	WithMountPath("/usr/bin/docker"),
							//applycorev1.VolumeMount().
							//	WithName("docker-lib").
							//	WithMountPath("/var/lib/docker"),
							//applycorev1.VolumeMount().
							//	WithName("docker-etc").
							//	WithMountPath("/etc/docker"),
							applycorev1.VolumeMount().
								WithName("inspection-agent-sh").
								WithMountPath("/opt/inspection"))).
					WithHostNetwork(true).
					WithServiceAccountName("inspection-agent").
					WithVolumes(
						applycorev1.Volume().
							WithName("inspection").
							WithHostPath(applycorev1.HostPathVolumeSource().
								WithPath("/")),
						//applycorev1.Volume().
						//	WithName("docker-socket").
						//	WithHostPath(applycorev1.HostPathVolumeSource().
						//		WithPath("/var/run/docker.sock")),
						//applycorev1.Volume().
						//	WithName("docker-bin").
						//	WithHostPath(applycorev1.HostPathVolumeSource().
						//		WithPath("/usr/bin/docker")),
						//applycorev1.Volume().
						//	WithName("docker-lib").
						//	WithHostPath(applycorev1.HostPathVolumeSource().
						//		WithPath("/var/lib/docker")),
						//applycorev1.Volume().
						//	WithName("docker-etc").
						//	WithHostPath(applycorev1.HostPathVolumeSource().
						//		WithPath("/etc/docker")),
						applycorev1.Volume().
							WithName("inspection-agent-sh").
							WithConfigMap(applycorev1.ConfigMapVolumeSource().
								WithName("inspection-agent-sh").
								WithOptional(false).
								WithDefaultMode(448)),
					),
				),
			),
		)
}
