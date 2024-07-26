package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Alerting struct {
	Data *Data `json:"data"`
}

type Data struct {
	RuleGroups []RuleGroup      `json:"groups"`
	Totals     map[string]int64 `json:"totals,omitempty"`
}

// swagger:model
type RuleGroup struct {
	// required: true
	Name string `json:"name"`
	// required: true
	File string `json:"file"`
	// In order to preserve rule ordering, while exposing type (alerting or recording)
	// specific properties, both alerting and recording rules are exposed in the
	// same array.
	// required: true
	Rules  []AlertingRule   `json:"rules"`
	Totals map[string]int64 `json:"totals"`
	// required: true
	Interval       float64   `json:"interval"`
	LastEvaluation time.Time `json:"lastEvaluation"`
	EvaluationTime float64   `json:"evaluationTime"`
}

type AlertingRule struct {
	// State can be "pending", "firing", "inactive".
	// required: true
	State string `json:"state,omitempty"`
	// required: true
	Name string `json:"name,omitempty"`
	// required: true
	Query    string  `json:"query,omitempty"`
	Duration float64 `json:"duration,omitempty"`
	// required: true
	Annotations map[string]string `json:"annotations,omitempty"`

	// required: true
	ActiveAt       *time.Time       `json:"activeAt,omitempty"`
	Alerts         []Alert          `json:"alerts,omitempty"`
	Totals         map[string]int64 `json:"totals,omitempty"`
	TotalsFiltered map[string]int64 `json:"totalsFiltered,omitempty"`
	Rule
}

type Alert struct {
	// required: true
	Labels map[string]string `json:"labels"`
	// required: true
	Annotations map[string]string `json:"annotations"`
	// required: true
	State    string     `json:"state"`
	ActiveAt *time.Time `json:"activeAt"`
	// required: true
	Value string `json:"value"`
}

type Rule struct {
	// required: true
	Name string `json:"name"`
	// required: true
	Query  string            `json:"query"`
	Labels map[string]string `json:"labels,omitempty"`
	// required: true
	Health    string `json:"health"`
	LastError string `json:"lastError,omitempty"`
	// required: true
	Type           string    `json:"type"`
	LastEvaluation time.Time `json:"lastEvaluation"`
	EvaluationTime float64   `json:"evaluationTime"`
}

func NewAlerting() *Alerting {
	return &Alerting{}
}

func main() {
	url := "https://192.168.2.155:8443/api/v1/namespaces/cattle-global-monitoring/services/http:access-grafana:80/proxy/api/prometheus/grafana/api/v1/rules" // 示例 URL

	// 创建一个新的 GET 请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// 设置请求头
	req.Header.Set("User-Agent", "My-Go-App")
	req.Header.Set("Authorization", "Bearer token-8ljpf:j272rljrb4dvf69j2twt25jk8cg95756blbg5s9pwcm2gr4vphs599")

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}

	alerting := NewAlerting()
	err = json.Unmarshal(body, alerting)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}

	da, err := json.Marshal(alerting)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}

	fmt.Println(string(da))
	// 输出响应内容
	fmt.Println("==============")
	fmt.Println("=============")
	fmt.Println(string(body))
}
