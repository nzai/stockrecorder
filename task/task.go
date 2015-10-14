package task

import (
	"log"

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

		//	启动监视
		err := market.Monitor()
		if err != nil {
			log.Printf("启动监视任务时发生错误: %s", err.Error())
		}
	}()

	return nil
}
