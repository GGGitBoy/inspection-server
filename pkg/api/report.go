package api

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"inspection-server/pkg/common"
	"inspection-server/pkg/db"
	pdfPrint "inspection-server/pkg/print"
	"io"
	"log"
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
			log.Fatal(err)
		}

		jsonData, err := json.MarshalIndent(report, "", "\t")
		if err != nil {
			log.Fatal(err)
		}

		rw.Write(jsonData)
	})
}

func PrintReport() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		p := pdfPrint.NewPrint()
		body, err := io.ReadAll(req.Body)
		if err != nil {
			log.Fatal(err)
		}

		err = json.Unmarshal(body, p)
		if err != nil {
			log.Fatal(err)
		}

		err = pdfPrint.FullScreenshot(p)
		if err != nil {
			log.Fatal(err)
		}

		file, err := os.Open(common.PrintPDFPath)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		fileInfo, err := file.Stat()
		if err != nil {
			log.Fatal(err)
		}

		// 设置响应头
		rw.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(common.PrintPDFPath))
		rw.Header().Set("Content-Type", "application/octet-stream")
		rw.Header().Set("Content-Length", fmt.Sprint(fileInfo.Size()))

		io.Copy(rw, file)
	})
}
