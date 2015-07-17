package crawl

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

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
		err := marketJob(market)
		if err != nil {
			log.Fatalf("启动[%s]数据抓取任务失败:%v", market, err)
		}
	}
}

//	抓取市场数据任务
func marketJob(market string) error {
	//	任务启动时间(hour)
	endhour := config.GetString(market, "endhour", "")
	if endhour == "" {
		return errors.New(fmt.Sprintf("市场[%s]的starthour配置有误", market))
	}

	hour, err := strconv.Atoi(endhour)
	if err != nil {
		return err
	}

	now := time.Now()
	startTime := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, now.Location())
	//	现在距离开始的时间间隔
	duration := startTime.Sub(now)
	if now.After(startTime) {
		//	今天的开始时间已经过了
		duration = duration + time.Hour*24
	}

	log.Printf("%s后开始抓取%s的数据", duration.String(), market)

	time.AfterFunc(duration, func() {
		//	到点后立即运行第一次
		go func(m string, h int) {
			//	抓取雅虎今日数据
			err := yahooToday(market, h)
			if err != nil {
				log.Fatal(err)
			}
		}(market, hour)

		//	之后每天运行一次
		ticker := time.NewTicker(time.Hour * 24)
		for _ = range ticker.C {
			//	抓取雅虎今日数据
			err := yahooToday(market, hour)
			if err != nil {
				log.Fatal(err)
			}
		}
	})

	return nil
}
