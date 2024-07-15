package main

import (
	"fmt"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"io/ioutil"
	"log"
	"time"
)

func main() {
	// 启动浏览器并连接到它
	path, _ := launcher.LookPath()
	u := launcher.New().Bin(path).MustLaunch()
	browser := rod.New().ControlURL(u).MustConnect()
	defer browser.MustClose()

	page := browser.MustPage("http://54.180.112.220:30144/#/inspection-record/report-pdf-view/d491fea4-26a0-45a5-9285-b6f196d1750a")
	page.MustWaitLoad()

	time.Sleep(10 * time.Second)
	//// 获取 document.body.scrollHeight 的值
	//scrollHeight, err := page.Eval(`document.body.scrollHeight`)
	//if err != nil {
	//	fmt.Println("Error getting scrollHeight:", err)
	//	return
	//}

	// 获取页面内容的尺寸
	metrics := page.MustEval(`() => ({
		width: document.body.scrollWidth,
		height: document.body.scrollHeight,
	})`)

	width := metrics.Get("width").Int()
	height := metrics.Get("height").Int()

	fmt.Println(width)
	fmt.Println(height)
	// 设置视窗尺寸
	page.MustSetViewport(width, height, 1, false)

	// 确保所有懒加载内容都已加载
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

	//page.MustWaitElementsMoreThan(".iframe", 0)
	fmt.Println("aaa")
	time.Sleep(10 * time.Second)

	//page.MustSetWindow(0, 0, 1280, 1280)
	//page.MustSetViewport(1280, 1280, 1, false)

	// 等待页面加载完成
	// 等待特定元素达到一定数量

	//page.MustWaitElementsMoreThan(".scrollbar-view", 0)

	// 截取整个页面截图
	screenshot, err := page.Screenshot(false, nil)
	if err != nil {
		log.Fatalf("Failed to capture screenshot: %v", err)
	}

	err = ioutil.WriteFile("screenshot.png", screenshot, 0755)
	if err != nil {
		log.Fatalf("Failed to save screenshot: %v", err)
	}

	log.Println("Screenshot saved to screenshot.png")
}
