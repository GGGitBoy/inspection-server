package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/chromedp/chromedp"
)

func main() {

	opts := append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoDefaultBrowserCheck,                        //不检查默认浏览器
		chromedp.Flag("headless", true),                       // 开启窗口模式
		chromedp.Flag("blink-settings", "imagesEnabled=true"), //开启图像界面,重点是开启这个
		chromedp.Flag("ignore-certificate-errors", true),      //忽略错误
		chromedp.Flag("disable-web-security", true),           //禁用网络安全标志
		chromedp.Flag("disable-extensions", true),             //开启插件支持
		chromedp.Flag("disable-default-apps", true),
		chromedp.WindowSize(1920, 1080),    // 设置浏览器分辨率（窗口大小）
		chromedp.Flag("disable-gpu", true), //开启 gpu 渲染
		chromedp.Flag("hide-scrollbars", true),
		chromedp.Flag("mute-audio", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("no-default-browser-check", true),

		chromedp.NoFirstRun, //设置网站不是首次运行
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.164 Safari/537.36"), //设置UserAgent
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	fmt.Println("bbb")
	ctx, cancel := chromedp.NewContext(
		allocCtx,
		chromedp.WithLogf(log.Printf),
	)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 300*time.Second)
	defer cancel()

	fmt.Println("aa")
	var buf []byte
	if err := chromedp.Run(ctx,
		//chromedp.EmulateViewport(1280, 1024),
		chromedp.Navigate("http://54.180.24.161:32572/#/inspection-record/report-pdf-view/0ec748db-2002-469e-839a-414d097448ea"),
		//chromedp.WaitVisible(`body`, chromedp.ByQuery),
		//chromedp.WaitVisible(`#app`, chromedp.ByID),
		// 获取页面高度
		chromedp.Sleep(10*time.Second),
		chromedp.Evaluate(`() => {
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
		}`, nil),
		chromedp.Sleep(10*time.Second),
		//chromedp.Evaluate(`document.documentElement.scrollHeight`, &pageHeight),
		// 设置浏览器窗口大小
		chromedp.EmulateViewport(1920, 2024),
		chromedp.FullScreenshot(&buf, 100),
	); err != nil {
		log.Fatal(err)
	}

	fmt.Println("eeee")

	if err := ioutil.WriteFile("grafana.png", buf, 0644); err != nil {
		log.Fatal(err)

	}
}
