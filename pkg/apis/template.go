package apis

type Template struct {
	ID               string              `json:"id"`
	Name             string              `json:"name"`
	KubernetesConfig []*KubernetesConfig `json:"kubernetes"`
}

type KubernetesConfig struct {
	Enable         bool            `json:"enable"`
	ClusterID      string          `json:"cluster_id"`
	ClusterName    string          `json:"cluster_name"`
	WorkloadConfig *WorkloadConfig `json:"workloads"`
	NodeConfig     []*NodeConfig   `json:"nodes"`
}

type WorkloadConfig struct {
	Deployment  []*WorkloadDetailConfig `json:"deployment"`
	Statefulset []*WorkloadDetailConfig `json:"statefulset"`
	Daemonset   []*WorkloadDetailConfig `json:"daemonset"`
	Job         []*WorkloadDetailConfig `json:"job"`
	Cronjob     []*WorkloadDetailConfig `json:"cronjob"`
}

type WorkloadDetailConfig struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Regexp    string `json:"regexp"`
	Core      bool   `json:"core"`
}

type NodeConfig struct {
	Names    []string         `json:"names"`
	Commands []*CommandConfig `json:"commands"`
}

type CommandConfig struct {
	Description string `json:"description"`
	Command     string `json:"command"`
	Core        bool   `json:"core"`
}

func NewTemplate() *Template {
	return &Template{
		KubernetesConfig: []*KubernetesConfig{},
	}
}

func NewTemplates() []*Template {
	return []*Template{}
}

func NewKubernetesConfig() []*KubernetesConfig {
	return []*KubernetesConfig{}
}
