package api

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"inspection-server/pkg/db"
	"log"
	"net/http"
)

func GetRecord() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		recordID := vars["id"]

		record, err := db.GetRecord(recordID)
		if err != nil {
			log.Fatal(err)
		}

		jsonData, err := json.MarshalIndent(record, "", "\t")
		if err != nil {
			log.Fatal(err)
		}

		rw.Write(jsonData)
	})
}

func ListRecord() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		records, err := db.ListRecord()
		if err != nil {
			log.Fatal(err)
		}

		jsonData, err := json.MarshalIndent(records, "", "\t")
		if err != nil {
			log.Fatal(err)
		}

		rw.Write(jsonData)
	})
}

func DeleteRecord() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		recordID := vars["id"]

		record, err := db.GetRecord(recordID)
		if err != nil {
			log.Fatal(err)
		}

		err = db.DeleteRecord(recordID)
		if err != nil {
			log.Fatal(err)
		}

		err = db.DeleteReport(record.ReportID)
		if err != nil {
			log.Fatal(err)
		}
	})
}
