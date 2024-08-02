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
	"log"
	"net/http"
	"strings"
)

func GetNotify() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		notifyID := vars["id"]

		notify, err := db.GetNotify(notifyID)
		if err != nil {
			log.Fatal(err)
		}

		notify.AppSecret = ""
		jsonData, err := json.MarshalIndent(notify, "", "\t")
		if err != nil {
			log.Fatal(err)
		}

		rw.Write(jsonData)
	})
}

func ListNotify() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		notifys, err := db.ListNotify()
		if err != nil {
			log.Fatal(err)
		}

		jsonData, err := json.MarshalIndent(notifys, "", "\t")
		if err != nil {
			log.Fatal(err)
		}

		rw.Write(jsonData)
	})
}

func CreateNotify() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		notify := apis.NewNotify()
		body, err := io.ReadAll(req.Body)
		if err != nil {
			log.Fatal(err)
		}

		err = json.Unmarshal(body, notify)
		if err != nil {
			log.Fatal(err)
		}

		notify.ID = common.GetUUID()
		err = db.CreateNotify(notify)
		if err != nil {
			log.Fatal(err)
		}

		rw.Write([]byte("创建完成"))
	})
}

func UpdateNotify() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		notify := apis.NewNotify()
		body, err := io.ReadAll(req.Body)
		if err != nil {
			log.Fatal(err)
		}

		err = json.Unmarshal(body, notify)
		if err != nil {
			log.Fatal(err)
		}

		err = db.UpdateNotify(notify)
		if err != nil {
			log.Fatal(err)
		}

		rw.Write([]byte("更新完成"))
	})
}

func DeleteNotify() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		notifyID := vars["id"]

		plans, err := db.ListPlan()
		if err != nil {
			log.Fatal(err)
		}

		for _, p := range plans {
			if p.ID == notifyID {
				rw.Write([]byte("该通知在被使用"))
				return
			}
		}

		err = db.DeleteNotify(notifyID)
		if err != nil {
			log.Fatal(err)
		}
	})
}

func TestNotify() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		notify := apis.NewNotify()
		body, err := io.ReadAll(req.Body)
		if err != nil {
			log.Fatal(err)
		}

		err = json.Unmarshal(body, notify)
		if err != nil {
			log.Fatal(err)
		}

		//message := `该巡检报告的健康等级为:\n该巡检报告的健康等级为:\n该巡检报告的健康等级为:\n该巡检报告的健康等级为:\n该巡检报告的健康等级为:\n该巡检报告的健康等级为:\n该巡检报告的健康等级为:`

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf(`该巡检报告的健康等级为: %s\n`, "report.Global.Rating"))
		str := sb.String()
		fmt.Println(str)

		message := str

		err = send.Notify(notify.AppID, notify.AppSecret, common.SendTestPDFName, common.SendTestPDFPath, message)
		if err != nil {
			log.Fatal(err)
		}

		rw.Write([]byte("测试完成"))
	})
}
