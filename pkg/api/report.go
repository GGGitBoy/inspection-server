package api

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"inspection-server/pkg/common"
	"inspection-server/pkg/db"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func GetReport() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		reportID := vars["id"]
		report, err := db.GetReport(reportID)
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		jsonData, err := json.MarshalIndent(report, "", "\t")
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		rw.Write(jsonData)
	})
}

func PrintReport() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		reportID := vars["id"]
		report, err := db.GetReport(reportID)
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		//url := "http://127.0.0.1/#/inspection/result-pdf-view/" + reportID
		//
		//p := pdfPrint.NewPrint()
		//p.URL = url
		//p.ReportTime = report.Global.ReportTime
		//err = pdfPrint.FullScreenshot(p)
		//if err != nil {
		//	common.HandleError(rw, http.StatusInternalServerError, err)
		//	return
		//}

		file, err := os.Open(common.PrintPDFPath + common.GetReportFileName(report.Global.ReportTime))
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}
		defer file.Close()

		fileInfo, err := file.Stat()
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		rw.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(common.GetReportFileName(report.Global.ReportTime)))
		rw.Header().Set("Content-Type", "application/octet-stream")
		rw.Header().Set("Content-Length", fmt.Sprint(fileInfo.Size()))
		_, err = io.Copy(rw, file)
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		rw.Write([]byte("打印成功"))
	})
}
