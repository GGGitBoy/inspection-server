package common

import (
	"fmt"
	"log"
	"os"
)

var (
	ServerURL   = getEnv("SERVER_URL", "")
	BearerToken = getEnv("BEARER_TOKEN", "")

	MySQL         = getEnv("MY_SQL", "")
	MySQLUser     = getEnv("MY_SQL_USER", "")
	MySQLPassword = getEnv("MY_SQL_PASSWORD", "")
	MySQLHost     = getEnv("MY_SQL_HOST", "")
	MySQLPort     = getEnv("MY_SQL_PORT", "")
	MySQLDB       = getEnv("MY_SQL_DB", "")

	SQLiteName          = "sqlite.db"
	AgentName           = "inspection-agent"
	AgentScriptName     = "inspection-agent-sh"
	InspectionNamespace = "cattle-inspection-system"

	PrintWaitSecond = getEnv("PRINT_WAIT_SECOND", "")

	LocalCluster = "local"

	// WorkDir should be retrieved from an environment variable or configuration file
	WorkDir = "/opt/"

	ConfigFilePath = WorkDir + "config/config.yml"

	PrintShotPath = WorkDir + "print/screenshot.png"
	PrintPDFPath  = WorkDir + "print/"
	// PrintPDFName  = "report.pdf"

	WriteKubeconfigPath = WorkDir + "kubeconfig/"

	SendTestPDFPath = WorkDir + SendTestPDFName
	SendTestPDFName = "test.pdf"

	AgentYamlPath = WorkDir + "yaml/"
)

// getEnv retrieves the environment variable or returns the default value if not set.
func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Printf("Environment variable %s not set, using default value: %s", key, defaultValue)
		return defaultValue
	}
	return value
}

// GetReportFileName generates the report file name using the provided time string.
func GetReportFileName(time string) string {
	fileName := fmt.Sprintf("Report(%s).pdf", time)
	log.Printf("Generated report file name: %s", fileName)
	return fileName
}
