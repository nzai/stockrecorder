package task

import (
	"log"

	"github.com/nzai/stockrecorder/analyse"
	"github.com/nzai/stockrecorder/market"
)

//	启动任务
func StartTasks() error {

	go func() {
		log.Print("启动抓取任务")

		//	美国股市
		market.Add(market.America{})
		//	中国股市
		market.Add(market.China{})
		//	香港股市
		market.Add(market.HongKong{})

		market.Monitor()
	}()

	go func() {
		log.Print("启动分析任务")
		
		//	分析历史数据
		analyse.AnalyseHistory()
	}()

	return nil
}
