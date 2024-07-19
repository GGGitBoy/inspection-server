package apis

type Record struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Mode      int    `json:"mode"`
	ReportID  string `json:"report_id"`
}

func NewRecords() []*Record {
	return []*Record{}
}

func NewRecord() *Record {
	return &Record{}
}
