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

	fmt.Println(time.Now().Format(time.DateTime))
	page := browser.MustPage(print.URL)
	page.MustWaitLoad()

	//page.WaitElementsMoreThan()

	//`(s, n) => document.querySelectorAll(s).length >= n`

	time.Sleep(time.Duration(waitSecond) * time.Second)

	// 等待条件并获取 allElements.length
	aaa := page.MustEval(`() => {
		const iframes = document.querySelectorAll('iframe');
		let allElements = [];
		iframes.forEach(iframe => {
			try {
				const iframeDocument = iframe.contentDocument || iframe.contentWindow.document;
				if (iframeDocument) {
					const elements = iframeDocument.querySelectorAll(".css-kvzgb9-panel-content");
					allElements = allElements.concat(Array.from(elements));
				}
			} catch (error) {
				console.warn("Could not access iframe content due to cross-origin restrictions:", error);
			}
		});
		return { length: allElements.length };
	}`)

	// 打印 allElements.length
	fmt.Println(aaa.Get("length").Int())

	//page.MustEval(`() => {
	//	var totalWidth = 0;
	//	var distance = 100;
	//	var timer = setInterval(() => {
	//		var scrollWidth = document.body.scrollWidth;
	//		window.scrollBy(distance, 0);
	//		totalWidth += distance;
	//		if(totalWidth >= scrollWidth){
	//			clearInterval(timer);
	//		}
	//	}, 100);
	//}`)
	fmt.Println(time.Now().Format(time.DateTime))
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
		fmt.Println(err)
	}

	page.MustWaitRequestIdle()
	err = page.WaitElementsMoreThan(".iframe", 2)
	//err = page.Wait(rod.Eval(`() => ({
	//	return document.querySelectorAll(".iframe").length >= 3;
	//})`))
	if err != nil {
		log.Fatalf("Failed to iframe: %v", err)
	}
	//
	//err = page.Wait(rod.Eval(`() => {
	//	const iframes = document.querySelectorAll(".iframe");
	//	let allElements = [];
	//	iframes.forEach(iframe => {
	//		try {
	//			const iframeDocument = iframe.contentDocument || iframe.contentWindow.document;
	//			if (iframeDocument) {
	//				const elements = iframeDocument.querySelectorAll(".css-kvzgb9-panel-content");
	//				allElements = allElements.concat(Array.from(elements));
	//			}
	//		} catch (error) {
	//			console.warn("Could not access iframe content due to cross-origin restrictions:", error);
	//			throw new Error("Accessing iframe content failed");
	//		}
	//	});
	//	return allElements.length >= 44;
	//}`))
	//if err != nil {
	//	log.Fatalf("Failed to evaluate JavaScript: %v", err)
	//}

	fmt.Println(time.Now().Format(time.DateTime))
	//time.Sleep(20 * time.Second)

	bbb := page.MustEval(`() => {
		const iframes = document.querySelectorAll('iframe');
		let allElements = [];
		iframes.forEach(iframe => {
			try {
				const iframeDocument = iframe.contentDocument || iframe.contentWindow.document;
				if (iframeDocument) {
					const elements = iframeDocument.querySelectorAll(".css-kvzgb9-panel-content");
					allElements = allElements.concat(Array.from(elements));
				}
			} catch (error) {
				console.warn("Could not access iframe content due to cross-origin restrictions:", error);
			}
		});
		return { length: allElements.length };
	}`)
	fmt.Println(time.Now().Format(time.DateTime))
	// 打印 allElements.length
	fmt.Println(bbb.Get("length").Int())

	metrics := page.MustEval(`() => ({
		width: document.body.scrollWidth,
		height: document.body.scrollHeight,
	})`)

	fmt.Println(metrics.Get("width").Int())
	fmt.Println(metrics.Get("height").Int())
	fmt.Println("=========")
	page.MustSetViewport(metrics.Get("width").Int(), metrics.Get("height").Int(), 1, false)

	screenshot, err := page.Screenshot(false, nil)
	if err != nil {
		log.Fatalf("Failed to capture screenshot: %v", err)
	}
	fmt.Println(time.Now().Format(time.DateTime))

	err = common.WriteFile(common.PrintShotPath, screenshot)
	if err != nil {
		return err
	}

	err = ToPrintPDF(print)
	if err != nil {
		return err
	}

	return nil
}

func ToPrintPDF(print *Print) error {
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
	err = pdf.WritePdf(common.PrintPDFPath + common.GetReportFileName(print.ReportTime))
	if err != nil {
		return err
	}

	return nil
}
