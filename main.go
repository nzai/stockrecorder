package main

import (
	"log"
	"runtime/debug"

	"github.com/nzai/stockrecorder/config"
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
	conf, err := config.Parse()
	if err != nil {
		log.Fatal("读取配置文件错误: ", err)
	}

	log.Print("启动市场监视任务")

	// 创建记录器，使用雅虎财经作为数据源，亚马逊S3作为存储
	r := recorder.NewRecorder(conf,
		source.YahooFinance{},
		store.AmazonS3{},
		market.America{},
		market.China{},
		market.HongKong{},
	)
	r.RunAndWait()
}
