package main

import (
	"fmt"
	pdfPrint "inspection-server/pkg/print"
)

func main() {
	p := pdfPrint.NewPrint()
	p.URL = "http://3.39.195.125:31028/#/inspection/result-pdf-view/5074001a-e897-4be3-95d0-f47ab2aee2b4"
	p.ReportTime = "2024-08-29 23:27:58"
	err := pdfPrint.FullScreenshot(p)
	if err != nil {
		fmt.Println(err)
	}
}
