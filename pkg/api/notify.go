package api

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"inspection-server/pkg/apis"
	"inspection-server/pkg/common"
	"inspection-server/pkg/db"
	"inspection-server/pkg/send"
	"io"
	"net/http"
)

func GetNotify() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		notifyID := vars["id"]
		notify, err := db.GetNotify(notifyID)
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		notify.AppSecret = ""
		jsonData, err := json.MarshalIndent(notify, "", "\t")
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		rw.Write(jsonData)
	})
}

func ListNotify() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		notifys, err := db.ListNotify()
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		jsonData, err := json.MarshalIndent(notifys, "", "\t")
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		rw.Write(jsonData)
	})
}

func CreateNotify() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		notify := apis.NewNotify()
		body, err := io.ReadAll(req.Body)
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		err = json.Unmarshal(body, notify)
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		notify.ID = common.GetUUID()
		err = db.CreateNotify(notify)
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		rw.Write([]byte("创建完成"))
	})
}

func UpdateNotify() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		notify := apis.NewNotify()
		body, err := io.ReadAll(req.Body)
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		err = json.Unmarshal(body, notify)
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		err = db.UpdateNotify(notify)
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		rw.Write([]byte("更新完成"))
	})
}

func DeleteNotify() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		notifyID := vars["id"]
		tasks, err := db.ListTask()
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		for _, t := range tasks {
			if t.ID == notifyID {
				rw.Write([]byte(fmt.Sprintf("该通知在被巡检任务 %s 使用无法删除", t.Name)))
				return
			}
		}

		err = db.DeleteNotify(notifyID)
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		rw.Write([]byte("删除完成"))
	})
}

func TestNotify() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		notify := apis.NewNotify()
		body, err := io.ReadAll(req.Body)
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		err = json.Unmarshal(body, notify)
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		message := "测试成功"
		err = send.Notify(notify.AppID, notify.AppSecret, common.SendTestPDFName, common.SendTestPDFPath, message)
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		rw.Write([]byte("测试成功"))
	})
}
