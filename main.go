package main

import (
	"log"
	"runtime/debug"

	"github.com/nzai/stockrecorder/market"
	"github.com/nzai/stockrecorder/recorder"
	"github.com/nzai/stockrecorder/source"
	"github.com/nzai/stockrecorder/store"
)

func main() {

	defer func() {
		// 捕获panic异常
		log.Print("发生了致命错误")
		if err := recover(); err != nil {
			log.Print("致命错误:", err)
		}
		debug.PrintStack()
	}()

	//	读取配置文件
	config, err := parseConfig()
	if err != nil {
		log.Fatal("读取配置文件错误: ", err)
	}

	log.Print("启动市场监视任务")

	// 创建记录器，使用雅虎财经作为数据源，亚马逊S3作为存储，监控美股、A股、港股
	r := recorder.NewRecorder(
		source.YahooFinance{},               // 雅虎财经作为数据源
		store.NewAmazonS3(config.Amazon.S3), // 亚马逊S3作为存储
		market.America{},                    // 记录美股
		market.China{},                      // A股
		market.HongKong{},                   // 港股
	)
	r.RunAndWait()
}
