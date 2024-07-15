package api

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"inspection-server/pkg/common"
	"inspection-server/pkg/db"
	pdfPrint "inspection-server/pkg/print"
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
		p := pdfPrint.NewPrint()
		body, err := io.ReadAll(req.Body)
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		err = json.Unmarshal(body, p)
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		err = pdfPrint.FullScreenshot(p)
		if err != nil {
			common.HandleError(rw, http.StatusInternalServerError, err)
			return
		}

		file, err := os.Open(common.PrintPDFPath + common.GetReportFileName(p.ReportTime))
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

		rw.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(common.GetReportFileName(p.ReportTime)))
		rw.Header().Set("Content-Type", "application/octet-stream")
		rw.Header().Set("Content-Length", fmt.Sprint(fileInfo.Size()))
		io.Copy(rw, file)

		rw.Write([]byte("打印成功"))
	})
}
