package apis

type Record struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	StartTime  string `json:"start_time"`
	EndTime    string `json:"end_time"`
	Timer      string `json:"timer"`
	Cron       string `json:"cron"`
	State      string `json:"state"`
	Rating     string `json:"rating"`
	ReportID   string `json:"report_id"`
	TemplateID string `json:"template_id"`
	NotifyID   string `json:"notify_id"`
	PlanID     string `json:"plan_id"`
	Mode       int    `json:"mode"`
}

func NewRecords() []*Record {
	return []*Record{}
}

func NewRecord() *Record {
	return &Record{}
}
