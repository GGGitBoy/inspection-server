package main

import (
	"bytes"
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"sigs.k8s.io/yaml"
	"text/template"
)

func aa() {
	// 加载 kubeconfig 文件
	kubeconfig := "/Users/chenjiandao/jiandao/inspection-server/opt/kubeconfig/local"
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("Error loading kubeconfig: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error loading kubeconfig: %v", err)
	}

	// 读取 YAML 文件
	yamlFile, err := os.ReadFile("/Users/chenjiandao/jiandao/inspection-server/scripts/aaaa/namespace.yaml")
	if err != nil {
		log.Fatalf("Error reading YAML file: %v", err)
	}

	// 将 YAML 文件解码为 ConfigMapApplyConfiguration
	var namespace *v1.NamespaceApplyConfiguration
	err = yaml.Unmarshal(yamlFile, &namespace)
	if err != nil {
		log.Fatalf("Error unmarshaling YAML file: %v", err)
	}

	//// 设置 ConfigMapApplyConfiguration 的必要字段
	//configMapApplyConfig.WithKind("ConfigMap").WithAPIVersion("v1")

	_, err = clientset.CoreV1().Namespaces().Apply(context.TODO(), namespace, metav1.ApplyOptions{Force: true, FieldManager: "application/apply-patch"})
	if err != nil {
		log.Fatalf("Error applying ServiceAccount: %v", err)
	}

	fmt.Printf("ConfigMap applied successfully: %v\n", namespace)
}

// Values represents the values to be passed to the template
type Values struct {
	SetNamespace bool
}

func main() {
	// 读取模板文件
	templateFile := "/Users/chenjiandao/jiandao/inspection-server/scripts/aaaa/serviceaccount.yaml"
	tmpl, err := template.ParseFiles(templateFile)
	if err != nil {
		log.Fatalf("Error parsing template file: %v", err)
	}

	// 设置模板变量
	values := Values{
		SetNamespace: true, // or false based on your requirement
	}

	// 渲染模板
	var rendered bytes.Buffer
	err = tmpl.Execute(&rendered, map[string]interface{}{
		"Values": values,
	})
	if err != nil {
		log.Fatalf("Error executing template: %v", err)
	}

	// 输出渲染后的内容
	fmt.Println(rendered.String())
}
