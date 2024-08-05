package api

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"inspection-server/pkg/apis"
	"inspection-server/pkg/common"
	"inspection-server/pkg/db"
	"inspection-server/pkg/schedule"
	"io"
	"net/http"
)

func GetTask() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		taskID := vars["id"]
		task, err := db.GetTask(taskID)
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		jsonData, err := json.MarshalIndent(task, "", "\t")
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		rw.Write(jsonData)
	})
}

func ListTask() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		tasks, err := db.ListTask()
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		jsonData, err := json.MarshalIndent(tasks, "", "\t")
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		rw.Write(jsonData)
	})
}

func CreateTask() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		task := apis.NewTask()
		body, err := io.ReadAll(req.Body)
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		err = json.Unmarshal(body, task)
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		task.ID = common.GetUUID()
		task.State = "计划中"
		if task.TemplateID == "" {
			rw.Write([]byte("该计划没有对应的模版"))
			return
		}

		err = db.CreateTask(task)
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		err = schedule.AddSchedule(task)
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		rw.Write([]byte("创建完成"))
	})
}

func DeleteTask() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		taskID := vars["id"]
		task, err := db.GetTask(taskID)
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		if task.State == "巡检中" {
			rw.Write([]byte("巡检中的计划不能删除"))
			return
		}

		err = schedule.RemoveSchedule(task)
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		err = db.Deletetask(task.ID)
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		err = db.DeleteReport(task.ReportID)
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		rw.Write([]byte("删除完成"))
	})
}
