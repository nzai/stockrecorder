package task

import (
	"log"

	"github.com/nzai/stockrecorder/config"
	"github.com/nzai/stockrecorder/market"
)

//	启动任务
func StartTasks() error {

	//	启动抓取任务
	if config.Get().EnableCrawl {
		log.Print("启动抓取任务")
		go func() {
			//	美国股市
			market.Add(market.America{})
			//	中国股市
			market.Add(market.China{})
			//	香港股市
			market.Add(market.HongKong{})

			market.Monitor()
		}()
	}

	//	启动分析任务
	if config.Get().EnableCrawl {
		log.Print("启动分析任务")
		//	go analyse.StartJobs()
	}

	return nil
}
