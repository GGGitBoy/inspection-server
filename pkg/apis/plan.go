package apis

type Plan struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Timer      string `json:"timer"`
	Cron       string `json:"cron"`
	Mode       int    `json:"mode"`
	State      string `json:"state"`
	TemplateID string `json:"template_id"`
}

func NewPlans() []*Plan {
	return []*Plan{}
}

func NewPlan() *Plan {
	return &Plan{}
}
