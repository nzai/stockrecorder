package main

import (
	"log"

	"github.com/nzai/stockrecorder/config"
	"github.com/nzai/stockrecorder/market"
	"github.com/nzai/stockrecorder/server"
)

func main() {

	defer func() {
		// 捕获panic异常
		log.Print("发生了致命错误")
		if err := recover(); err != nil {
			log.Print("致命错误:", err)
		}
	}()

	//	读取配置文件
	err := config.Init()
	if err != nil {
		log.Fatal("读取配置文件错误: ", err)
		return
	}

	log.Print("启动市场监视任务")

	//	美国股市
	market.Add(market.America{})
	//	中国股市
	market.Add(market.China{})
	//	香港股市
	market.Add(market.HongKong{})

	//	启动监视
	err = market.Monitor()
	if err != nil {
		log.Printf("启动市场监视任务时发生错误: %s", err.Error())
	}

	//	启动http server
	server.Start()
}
