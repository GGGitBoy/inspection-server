package apis

import (
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

type Inspection struct {
	Clients map[string]*Client `json:"clients"`
	Reports []*Report          `json:"core"`
	Plans   []*Plan            `json:"plans"`
	Records []*Record          `json:"record"`
}

type Client struct {
	DynamicClient *dynamic.DynamicClient `json:"dynamic_client"`
	Clientset     *kubernetes.Clientset  `json:"clientset"`
	Config        *restclient.Config     `json:"config"`
}

type Report struct {
	ID         string                 `json:"id"`
	Global     *Global                `json:"global"`
	Kubernetes map[string]*Kubernetes `json:"kubernetes"`
}

type Global struct {
	Rating     int       `json:"rating"`
	ReportTime string    `json:"report_time"`
	Warnings   []Warning `json:"warnings"`
}

type Warning struct {
	Title   string `json:"title"`
	Message string `json:"message"`
	Type    int    `json:"type"`
}

type Kubernetes struct {
	Workloads             *Workload                `json:"workloads"`
	Nodes                 []*Node                  `json:"nodes"`
	Namespace             []*Namespace             `json:"namespace"`
	PersistentVolumeClaim []*PersistentVolumeClaim `json:"persistent_volume_claim"`
	Service               []*Service               `json:"service"`
	Ingress               []*Ingress               `json:"ingress"`
}

type MetaData struct {
}

type Workload struct {
	Deployment  []*WorkloadData `json:"deployment"`
	Statefulset []*WorkloadData `json:"statefulset"`
	Daemonset   []*WorkloadData `json:"daemonset"`
	Job         []*WorkloadData `json:"job"`
	Cronjob     []*WorkloadData `json:"cronjob"`
}

type WorkloadData struct {
	Name      string  `json:"name"`
	Namespace string  `json:"namespace"`
	Pods      []*Pod  `json:"pods"`
	Status    *Status `json:"status"`
}

type Node struct {
	Name        string            `json:"name"`
	Annotations map[string]string `json:"annotations"`
	Commands    *Command          `json:"commands"`
}

type Namespace struct {
	Name               string `json:"name"`
	EmptyResourceQuota bool   `json:"empty_resource_quota"`
}

type PersistentVolumeClaim struct {
	Name  string `json:"name"`
	State string `json:"state"`
}

type Service struct {
	Name           string `json:"name"`
	Namespace      string `json:"namespace"`
	EmptyEndpoints bool   `json:"empty_endpoints"`
}

type Ingress struct {
	Name          string `json:"name"`
	DuplicatePath bool   `json:"duplicate_path"`
}

type Pod struct {
	Name string   `json:"name"`
	Log  []string `json:"log"`
}

type Status struct {
	State     string      `json:"state"`
	Condition []Condition `json:"condition"`
}

type Condition struct {
	Type   string `json:"type"`
	Status string `json:"status"`
	Reason string `json:"reason"`
}

type Command struct {
	Stdout []CommandCheckResult `json:"stdout"`
	Stderr string               `json:"stderr"`
}

type CommandCheckResult struct {
	Description string `json:"description"`
	Command     string `json:"command"`
	Response    string `json:"response"`
	Error       string `json:"error"`
}

type Plan struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Timer string `json:"timer"`
	Cron  string `json:"cron"`
	Mode  int    `json:"mode"`
	State string `json:"state"`
}

type Record struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Mode      int    `json:"mode"`
	ReportID  string `json:"report_id"`
}

func NewClients() map[string]*Client {
	return make(map[string]*Client)
}

func NewReport() *Report {
	return &Report{
		Global:     &Global{},
		Kubernetes: make(map[string]*Kubernetes),
	}
}

func NewKubernetes() *Kubernetes {
	return &Kubernetes{
		Workloads: &Workload{
			Deployment:  []*WorkloadData{},
			Statefulset: []*WorkloadData{},
			Daemonset:   []*WorkloadData{},
			Job:         []*WorkloadData{},
			Cronjob:     []*WorkloadData{},
		},
		Nodes:                 []*Node{},
		Namespace:             []*Namespace{},
		PersistentVolumeClaim: []*PersistentVolumeClaim{},
		Service:               []*Service{},
		Ingress:               []*Ingress{},
	}
}

func NewPlan() *Plan {
	return &Plan{}
}

func NewRecord() *Record {
	return &Record{}
}

func NewPlans() []*Plan {
	return []*Plan{}
}

func NewRecords() []*Record {
	return []*Record{}
}

func NewPods() []*Pod {
	return []*Pod{}
}

func NewWorkloadDatas() []*WorkloadData {
	return []*WorkloadData{}
}
