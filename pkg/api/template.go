package api

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"inspection-server/pkg/apis"
	"inspection-server/pkg/common"
	"inspection-server/pkg/db"
	"io"
	"net/http"
)

func GetTemplate() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		templateID := vars["id"]
		template, err := db.GetTemplate(templateID)
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		jsonData, err := json.MarshalIndent(template, "", "\t")
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		rw.Write(jsonData)
	})
}

func ListTemplate() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		templates, err := db.ListTemplate()
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		jsonData, err := json.MarshalIndent(templates, "", "\t")
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		rw.Write(jsonData)
	})
}

func CreateTemplate() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		template := apis.NewTemplate()
		body, err := io.ReadAll(req.Body)
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		err = json.Unmarshal(body, template)
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		template.ID = common.GetUUID()
		err = db.CreateTemplate(template)
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		rw.Write([]byte("创建完成"))
	})
}

func UpdateTemplate() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		template := apis.NewTemplate()
		body, err := io.ReadAll(req.Body)
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		err = json.Unmarshal(body, template)
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		err = db.UpdateTemplate(template)
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		rw.Write([]byte("更新完成"))
	})
}

func DeleteTemplate() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		templateID := vars["id"]
		tasks, err := db.ListTask()
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		for _, t := range tasks {
			if t.TemplateID == templateID {
				rw.Write([]byte(fmt.Sprintf("该通知在被巡检任务 %s 使用无法删除", t.Name)))
				return
			}
		}

		err = db.DeleteTemplate(templateID)
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		rw.Write([]byte("删除完成"))
	})
}
