package api

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"inspection-server/pkg/apis"
	"inspection-server/pkg/common"
	"inspection-server/pkg/db"
	"io"
	"log"
	"net/http"
)

func GetTemplate() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		templateID := vars["id"]

		template, err := db.GetTemplate(templateID)
		if err != nil {
			log.Fatal(err)
		}

		jsonData, err := json.MarshalIndent(template, "", "\t")
		if err != nil {
			log.Fatal(err)
		}

		rw.Write(jsonData)
	})
}

func ListTemplate() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		templates, err := db.ListTemplate()
		if err != nil {
			log.Fatal(err)
		}

		jsonData, err := json.MarshalIndent(templates, "", "\t")
		if err != nil {
			log.Fatal(err)
		}

		rw.Write(jsonData)
	})
}

func CreateTemplate() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		template := apis.NewTemplate()
		body, err := io.ReadAll(req.Body)
		if err != nil {
			log.Fatal(err)
		}

		err = json.Unmarshal(body, template)
		if err != nil {
			log.Fatal(err)
		}

		template.ID = common.GetUUID()
		err = db.CreateTemplate(template)
		if err != nil {
			log.Fatal(err)
		}

		rw.Write([]byte("创建完成"))
	})
}

func UpdateTemplate() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		template := apis.NewTemplate()
		body, err := io.ReadAll(req.Body)
		if err != nil {
			log.Fatal(err)
		}

		err = json.Unmarshal(body, template)
		if err != nil {
			log.Fatal(err)
		}

		err = db.UpdateTemplate(template)
		if err != nil {
			log.Fatal(err)
		}

		rw.Write([]byte("更新完成"))
	})
}

func DeleteTemplate() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		templateID := vars["id"]

		plans, err := db.ListPlan()
		if err != nil {
			log.Fatal(err)
		}

		for _, p := range plans {
			if p.TemplateID == templateID {
				rw.Write([]byte("该模版在被使用"))
				return
			}
		}

		err = db.DeleteTemplate(templateID)
		if err != nil {
			log.Fatal(err)
		}
	})
}
