package main

import (
	"fmt"
	pdfPrint "inspection-server/pkg/print"
)

func main() {
	p := pdfPrint.NewPrint()
	p.URL = "http://54.180.103.65:30489/#/inspection-record/result-pdf-view/fe1d8e4c-c290-4763-b4f9-3808dce0871c"
	err := pdfPrint.FullScreenshot(p)
	if err != nil {
		fmt.Println(err)
	}

}
