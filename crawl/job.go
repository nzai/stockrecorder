package crawl

import (
	"log"

	"github.com/nzai/stockrecorder/config"
)

//	启动
func Start() {

	markets := config.GetArray("market", "markets")
	if len(markets) == 0 {
		//	未定义市场就直接返回
		return
	}

	for _, market := range markets {
		//	抓取市场数据
		err := marketAll(market)
		if err != nil {
			log.Fatalf("启动[%s]数据抓取任务失败:%v", market, err)
		}
	}
}
