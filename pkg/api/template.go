package api

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
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
		logrus.Infof("Received request to get template with ID: %s", templateID)

		template, err := db.GetTemplate(templateID)
		if err != nil {
			logrus.Errorf("Failed to get template with ID %s: %v", templateID, err)
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		jsonData, err := json.MarshalIndent(template, "", "\t")
		if err != nil {
			logrus.Errorf("Failed to marshal template data for ID %s: %v", templateID, err)
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		_, err = rw.Write(jsonData)
		if err != nil {
			logrus.Errorf("Failed to write response for template ID %s: %v", templateID, err)
			common.HandleError(rw, http.StatusInternalServerError, err)
		} else {
			logrus.Infof("Successfully retrieved and sent template with ID: %s", templateID)
		}
	})
}

func ListTemplate() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		logrus.Infof("Received request to list all templates")

		templates, err := db.ListTemplate()
		if err != nil {
			logrus.Errorf("Failed to list templates: %v", err)
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		jsonData, err := json.MarshalIndent(templates, "", "\t")
		if err != nil {
			logrus.Errorf("Failed to marshal template list data: %v", err)
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		_, err = rw.Write(jsonData)
		if err != nil {
			logrus.Errorf("Failed to write response for template list: %v", err)
			common.HandleError(rw, http.StatusInternalServerError, err)
		} else {
			logrus.Infof("Successfully retrieved and sent template list")
		}
	})
}

func CreateTemplate() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		logrus.Infof("Received request to create a new template")

		template := apis.NewTemplate()
		body, err := io.ReadAll(req.Body)
		if err != nil {
			logrus.Errorf("Failed to read request body: %v", err)
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		err = json.Unmarshal(body, template)
		if err != nil {
			logrus.Errorf("Failed to unmarshal template data: %v", err)
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		templates, err := db.ListTemplate()
		if err != nil {
			logrus.Errorf("Failed to list tasks: %v", err)
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		for _, t := range templates {
			if template.Name == t.Name {
				common.HandleError(rw, http.StatusInternalServerError, fmt.Errorf("该名称已存在"))
				return
			}
		}

		template.ID = common.GetUUID()
		err = db.CreateTemplate(template)
		if err != nil {
			logrus.Errorf("Failed to create template with ID %s: %v", template.ID, err)
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		rw.Write([]byte("创建完成"))
		logrus.Infof("Successfully created template with ID: %s", template.ID)
	})
}

func UpdateTemplate() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		logrus.Infof("Received request to update template")

		template := apis.NewTemplate()
		body, err := io.ReadAll(req.Body)
		if err != nil {
			logrus.Errorf("Failed to read request body: %v", err)
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		err = json.Unmarshal(body, template)
		if err != nil {
			logrus.Errorf("Failed to unmarshal template data: %v", err)
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		err = db.UpdateTemplate(template)
		if err != nil {
			logrus.Errorf("Failed to update template with ID %s: %v", template.ID, err)
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		rw.Write([]byte("更新完成"))
		logrus.Infof("Successfully updated template with ID: %s", template.ID)
	})
}

func DeleteTemplate() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		templateID := vars["id"]
		logrus.Infof("Received request to delete template with ID: %s", templateID)

		tasks, err := db.ListTask()
		if err != nil {
			logrus.Errorf("Failed to list tasks for template deletion: %v", err)
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		for _, t := range tasks {
			if t.TemplateID == templateID {
				logrus.Warnf("Template deletion failed: Template with ID %s is used in task %s", templateID, t.Name)
				rw.Write([]byte(fmt.Sprintf("该通知在被巡检任务 %s 使用无法删除", t.Name)))
				return
			}
		}

		err = db.DeleteTemplate(templateID)
		if err != nil {
			logrus.Errorf("Failed to delete template with ID %s: %v", templateID, err)
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		rw.Write([]byte("删除完成"))
		logrus.Infof("Successfully deleted template with ID: %s", templateID)
	})
}
