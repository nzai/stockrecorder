package task

import (
	"log"

	"github.com/nzai/stockrecorder/crawl"
	"github.com/nzai/stockrecorder/market"
)

//	启动任务
func StartTasks() error {
	log.Print("启动任务")

	go func() {
		market.Add(market.America{})
		
		market.Monitor()
	}()

	//	启动抓取任务
	go crawl.Start()

	//	启动分析任务
	//	go analyse.StartJobs()

	return nil
}
