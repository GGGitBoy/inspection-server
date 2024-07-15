package api

import (
	"encoding/json"
	"inspection-server/pkg/agent"
	"inspection-server/pkg/config"
	"io"
	"log"
	"net/http"
)

func GetConfig() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		configData, err := config.ReadConfigFile()
		if err != nil {
			log.Fatal(err)
		}

		jsonData, err := json.MarshalIndent(configData, "", "\t")
		if err != nil {
			log.Fatal(err)
		}

		rw.Write(jsonData)
	})
}

func UpdateConfig() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		configData := config.NewConfig()
		body, err := io.ReadAll(req.Body)
		if err != nil {
			log.Fatal(err)
		}

		err = json.Unmarshal(body, configData)
		if err != nil {
			log.Fatal(err)
		}

		err = config.WriteConfigFile(configData)
		if err != nil {
			log.Fatal(err)
		}

		err = agent.SyncAgent(configData)
		if err != nil {
			log.Fatal(err)
		}

		rw.Write([]byte("更新完成"))
	})
}
