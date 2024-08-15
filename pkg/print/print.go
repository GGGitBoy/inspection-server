package print

import (
	"fmt"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/signintech/gopdf"
	"image"
	"inspection-server/pkg/common"
	"log"
	"os"
	"strconv"
	"time"
)

var waitSecond = 10

type Print struct {
	URL        string `json:"url"`
	ReportTime string `json:"report_time"`
}

func NewPrint() *Print {
	return &Print{}
}

func FullScreenshot(print *Print) error {
	time.Sleep(2 * time.Second)
	if common.PrintWaitSecond != "" {
		num, err := strconv.Atoi(common.PrintWaitSecond)
		if err != nil {
			log.Printf("Invalid PrintWaitSecond value, using default: %v", err)
		} else {
			waitSecond = num
		}
	}

	path, ok := launcher.LookPath()
	if !ok {
		return fmt.Errorf("Failed to find browser path")
	}
	u := launcher.New().Bin(path).MustLaunch()
	browser := rod.New().ControlURL(u).MustConnect()
	defer browser.MustClose()

	log.Println("Starting page load")
	page := browser.MustPage(print.URL)
	page.MustWaitLoad()

	time.Sleep(time.Duration(waitSecond) * time.Second)

	page.MustEval(`() => {
		var totalHeight = 0;
		var distance = 100;
		var timer = setInterval(() => {
			var scrollHeight = document.body.scrollHeight;
			window.scrollBy(0, distance);
			totalHeight += distance;
			if(totalHeight >= scrollHeight){
				clearInterval(timer);
			}
		}, 1000);
	}`)

	err := page.Wait(rod.Eval(`() => document.body.scrollHeight <= (window.scrollY + window.innerHeight)`))
	if err != nil {
		log.Printf("Error while waiting for page scroll completion: %v", err)
		return err
	}

	metrics := page.MustEval(`() => ({
		width: document.body.scrollWidth,
		height: document.body.scrollHeight,
	})`)

	log.Printf("Page dimensions: width=%d, height=%d", metrics.Get("width").Int(), metrics.Get("height").Int())

	page.MustSetViewport(metrics.Get("width").Int(), metrics.Get("height").Int(), 1, false)

	screenshot, err := page.Screenshot(false, nil)
	if err != nil {
		log.Fatalf("Failed to capture screenshot: %v", err)
		return err
	}
	log.Println("Screenshot captured successfully")

	err = common.WriteFile(common.PrintShotPath, screenshot)
	if err != nil {
		log.Fatalf("Failed to save screenshot: %v", err)
		return err
	}

	err = ToPrintPDF(print)
	if err != nil {
		log.Fatalf("Failed to generate PDF: %v", err)
		return err
	}

	return nil
}

func ToPrintPDF(print *Print) error {
	imgFile, err := os.Open(common.PrintShotPath)
	if err != nil {
		log.Fatalf("Failed to open screenshot file: %v", err)
		return err
	}
	defer imgFile.Close()

	img, _, err := image.Decode(imgFile)
	if err != nil {
		log.Fatalf("Failed to decode image: %v", err)
		return err
	}

	imgWidth := img.Bounds().Dx()
	imgHeight := img.Bounds().Dy()

	pageWidth := 595.28
	scale := pageWidth / float64(imgWidth)
	newHeight := float64(imgHeight) * scale

	pdf := gopdf.GoPdf{}
	rect := &gopdf.Rect{
		W: pageWidth,
		H: newHeight,
	}
	pdf.Start(gopdf.Config{PageSize: *rect})
	pdf.AddPage()

	err = pdf.Image(common.PrintShotPath, 0, 0, rect)
	if err != nil {
		log.Fatalf("Failed to add image to PDF: %v", err)
		return err
	}

	err = pdf.WritePdf(common.PrintPDFPath + common.GetReportFileName(print.ReportTime))
	if err != nil {
		log.Fatalf("Failed to save PDF: %v", err)
		return err
	}

	log.Println("PDF generated successfully")
	return nil
}
