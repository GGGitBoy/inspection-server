package api

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
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
		logrus.Infof("Received request to get task with ID: %s", taskID)

		task, err := db.GetTask(taskID)
		if err != nil {
			logrus.Errorf("Failed to get task with ID %s: %v", taskID, err)
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		jsonData, err := json.MarshalIndent(task, "", "\t")
		if err != nil {
			logrus.Errorf("Failed to marshal task data for task ID %s: %v", taskID, err)
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		_, err = rw.Write(jsonData)
		if err != nil {
			logrus.Errorf("Failed to write response for task ID %s: %v", taskID, err)
			common.HandleError(rw, http.StatusInternalServerError, err)
		} else {
			logrus.Infof("Successfully retrieved and sent task with ID: %s", taskID)
		}
	})
}

func ListTask() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		logrus.Infof("Received request to list all tasks")

		tasks, err := db.ListTask()
		if err != nil {
			logrus.Errorf("Failed to list tasks: %v", err)
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		jsonData, err := json.MarshalIndent(tasks, "", "\t")
		if err != nil {
			logrus.Errorf("Failed to marshal task list data: %v", err)
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		_, err = rw.Write(jsonData)
		if err != nil {
			logrus.Errorf("Failed to write response for task list: %v", err)
			common.HandleError(rw, http.StatusInternalServerError, err)
		} else {
			logrus.Infof("Successfully retrieved and sent task list")
		}
	})
}

func CreateTask() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		logrus.Infof("Received request to create a new task")

		task := apis.NewTask()
		body, err := io.ReadAll(req.Body)
		if err != nil {
			logrus.Errorf("Failed to read request body: %v", err)
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		err = json.Unmarshal(body, task)
		if err != nil {
			logrus.Errorf("Failed to unmarshal task data: %v", err)
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		task.ID = common.GetUUID()
		task.State = "计划中"
		if task.TemplateID == "" {
			logrus.Warnf("Task creation failed: No template associated with the task")
			rw.Write([]byte("该计划没有对应的模版"))
			return
		}

		err = db.CreateTask(task)
		if err != nil {
			logrus.Errorf("Failed to create task with ID %s: %v", task.ID, err)
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		err = schedule.AddSchedule(task)
		if err != nil {
			logrus.Errorf("Failed to schedule task with ID %s: %v", task.ID, err)
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		rw.Write([]byte("创建完成"))
		logrus.Infof("Successfully created and scheduled task with ID: %s", task.ID)
	})
}

func DeleteTask() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		taskID := vars["id"]
		logrus.Infof("Received request to delete task with ID: %s", taskID)

		task, err := db.GetTask(taskID)
		if err != nil {
			logrus.Errorf("Failed to get task with ID %s: %v", taskID, err)
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		if task.State == "巡检中" {
			logrus.Warnf("Task deletion failed: Task with ID %s is currently in progress", taskID)
			rw.Write([]byte("巡检中的计划不能删除"))
			return
		}

		err = schedule.RemoveSchedule(task)
		if err != nil {
			logrus.Errorf("Failed to remove schedule for task with ID %s: %v", taskID, err)
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		err = db.DeleteTask(task.ID)
		if err != nil {
			logrus.Errorf("Failed to delete task with ID %s: %v", taskID, err)
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		err = db.DeleteReport(task.ReportID)
		if err != nil {
			logrus.Errorf("Failed to delete report for task ID %s: %v", taskID, err)
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		rw.Write([]byte("删除完成"))
		logrus.Infof("Successfully deleted task with ID: %s", taskID)
	})
}
