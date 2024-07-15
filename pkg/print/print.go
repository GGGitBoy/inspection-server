package print

import (
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
	URL string `json:"url"`
}

func NewPrint() *Print {
	return &Print{}
}

func FullScreenshot(print *Print) error {
	if common.PrintWaitSecond != "" {
		num, err := strconv.Atoi(common.PrintWaitSecond)
		if err == nil {
			waitSecond = num
		}
	}

	path, _ := launcher.LookPath()
	u := launcher.New().Bin(path).MustLaunch()
	browser := rod.New().ControlURL(u).MustConnect()
	defer browser.MustClose()

	page := browser.MustPage(print.URL)
	page.MustWaitLoad()

	time.Sleep(time.Duration(waitSecond) * time.Second)
	metrics := page.MustEval(`() => ({
		width: document.body.scrollWidth,
		height: document.body.scrollHeight,
	})`)
	page.MustSetViewport(metrics.Get("width").Int(), metrics.Get("height").Int(), 1, false)
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
		}, 100);
	}`)
	time.Sleep(time.Duration(waitSecond) * time.Second)

	screenshot, err := page.Screenshot(false, nil)
	if err != nil {
		log.Fatalf("Failed to capture screenshot: %v", err)
	}

	err = common.WriteFile(common.PrintShotPath, screenshot)
	if err != nil {
		return err
	}

	err = ToPrintPDF()
	if err != nil {
		return err
	}

	return nil
}

func ToPrintPDF() error {
	imgFile, err := os.Open(common.PrintShotPath)
	if err != nil {
		return err
	}
	defer imgFile.Close()

	// 解码图片
	img, _, err := image.Decode(imgFile)
	if err != nil {
		return err
	}

	// 获取图片的大小（像素）
	imgWidth := img.Bounds().Dx()
	imgHeight := img.Bounds().Dy()

	// A4页面的宽度（210mm）转换为点数（1 point = 1/72 inch, 1 inch = 25.4 mm）
	pageWidth := 595.28 // 210mm in points

	// 计算图片适应页面宽度的缩放比例
	scale := pageWidth / float64(imgWidth)

	// 计算图片在页面上的实际高度
	newHeight := float64(imgHeight) * scale

	// 创建一个新的PDF文档
	pdf := gopdf.GoPdf{}
	rect := &gopdf.Rect{
		W: pageWidth,
		H: newHeight,
	}
	pdf.Start(gopdf.Config{PageSize: *rect})

	pdf.AddPage()

	// 将图片部分添加到当前页
	err = pdf.Image(common.PrintShotPath, 0, 0, rect)
	if err != nil {
		return err
	}

	// 保存PDF文档
	err = pdf.WritePdf(common.PrintPDFPath)
	if err != nil {
		return err
	}

	return nil
}
