package common

import (
	"fmt"
	"os"
)

var (
	ServerURL   = os.Getenv("SERVER_URL")
	BearerToken = os.Getenv("BEARER_TOKEN")

	MySQL         = os.Getenv("MY_SQL")
	MySQLUser     = os.Getenv("MY_SQL_USER")
	MySQLPassword = os.Getenv("MY_SQL_PASSWORD")
	MySQLHost     = os.Getenv("MY_SQL_HOST")
	MySQLPort     = os.Getenv("MY_SQL_PORT")
	MySQLDB       = os.Getenv("MY_SQL_DB")

	SQLiteName          = "sqlite.db"
	AgentName           = "inspection-agent"
	AgentScriptName     = "inspection-agent-sh"
	InspectionNamespace = "cattle-inspection-system"

	PrintWaitSecond = os.Getenv("PRINT_WAIT_SECOND")

	LocalCluster = "local"

	//WorkDir = "/Users/chenjiandao/jiandao/inspection-server/opt/"
	WorkDir = "/opt/"

	ConfigFilePath = WorkDir + "config/config.yml"

	PrintShotPath = WorkDir + "print/screenshot.png"
	PrintPDFPath  = WorkDir + "print/"
	//PrintPDFName  = "report.pdf"

	WriteKubeconfigPath = WorkDir + "kubeconfig/"

	SendTestPDFPath = WorkDir + SendTestPDFName
	SendTestPDFName = "test.pdf"

	AgentYamlPath = WorkDir + "yaml/"
)

func GetReportFileName(time string) string {
	return fmt.Sprintf("Report(%s).pdf", time)
}
