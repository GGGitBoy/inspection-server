package api

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"inspection-server/pkg/apis"
	"inspection-server/pkg/common"
	"inspection-server/pkg/db"
	"inspection-server/pkg/schedule"
	"io"
	"log"
	"net/http"
)

func GetPlan() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		planID := vars["id"]

		plan, err := db.GetPlan(planID)
		if err != nil {
			log.Fatal(err)
		}

		jsonData, err := json.MarshalIndent(plan, "", "\t")
		if err != nil {
			log.Fatal(err)
		}

		rw.Write(jsonData)
	})
}

func ListPlan() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		plans, err := db.ListPlan()
		if err != nil {
			log.Fatal(err)
		}

		jsonData, err := json.MarshalIndent(plans, "", "\t")
		if err != nil {
			return
		}

		rw.Write(jsonData)
	})
}

func CreatePlan() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		plan := apis.NewPlan()
		body, err := io.ReadAll(req.Body)
		if err != nil {
			log.Fatal(err)
		}

		err = json.Unmarshal(body, plan)
		if err != nil {
			log.Fatal(err)
		}

		plan.ID = common.GetUUID()
		plan.State = "计划中"
		if plan.Mode == 0 {
			plan.Cron = ""
			plan.Timer = ""
		} else if plan.Mode == 1 {
			plan.Cron = ""
			if plan.Timer == "" {
				log.Fatal(err)
				return
			}
		} else if plan.Mode == 2 {
			plan.Timer = ""
			if plan.Cron == "" {
				log.Fatal(err)
				return
			}
		}

		if plan.Name == "" {
			plan.Name = "巡检计划"
		}

		if plan.TemplateID == "" {
			templates, err := db.ListTemplate()
			if err != nil {
				log.Fatal(err)
			}

			if len(templates) == 0 {
				rw.Write([]byte("该计划没有对应的模版"))
				return
			}

			plan.TemplateID = templates[0].ID
		}

		err = db.CreatePlan(plan)
		if err != nil {
			log.Fatal(err)
		}

		err = schedule.AddSchedule(plan)
		if err != nil {
			log.Fatal(err)
		}
	})
}

func DeletePlan() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		planID := vars["id"]

		plan, err := db.GetPlan(planID)
		if err != nil {
			log.Fatal(err)
		}

		if plan.State == "巡检中" {
			rw.Write([]byte("巡检中的计划不能删除"))
			return
		}

		err = schedule.RemoveSchedule(plan)
		if err != nil {
			log.Fatal(err)
		}

		err = db.DeletePlan(plan.ID)
		if err != nil {
			log.Fatal(err)
		}
	})
}
