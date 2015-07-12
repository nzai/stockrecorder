package crawl

import (
	"log"

	"github.com/nzai/stockrecorder/config"
)

func StartJobs() {
	log.Print("启动数据抓取任务")

	markets := config.GetArray("market", "markets")
	if len(markets) == 0 {
		return
	}

	for _, market := range markets {
		err := crawlMarket(market)
		if err != nil {
			log.Fatalf("启动[%s]数据抓取任务失败:%v", market, err)
		}
	}
}

//	抓取市场数据
func crawlMarket(market string) error {
	//	抓取今天的数据
	err := today(market)
	if err != nil {
		return err
	}

	//	抓取历史数据
	err = history(market)
	if err != nil {
		return err
	}

	return nil
}
