package crawl

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"time"

	"github.com/nzai/stockrecorder/config"
	"github.com/nzai/stockrecorder/io"
)

const (
	gccount = 16
)

func today(market string) error {
	//	任务启动时间(hour)
	starthour := config.GetString(market, "starthour", "")
	if starthour == "" {
		return errors.New(fmt.Sprintf("市场[%s]的starthour配置有误", market))
	}

	hour, err := strconv.Atoi(starthour)
	if err != nil {
		return err
	}

	//	股票编码
	codes := config.GetArray(market, "codes")
	if len(codes) == 0 {
		return errors.New(fmt.Sprintf("市场[%s]的codes配置有误", market))
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
		//	每天运行一次
		ticker := time.NewTicker(time.Hour * 24)
		for _ = range ticker.C {
			err := yahooToday(codes)
			if err != nil {
				log.Fatal(err)
			}
		}
	})

	return nil
}

//	抓取雅虎今日数据
func yahooToday(codes []string) error {
	chanSend := make(chan int, gccount)
	chanReceive := make(chan int)

	for _, code := range codes {
		//	并发抓取
		go func(c string) {
			//	更新每只股票的历史
			err := YahooTodayStock(c)
			if err != nil {
				log.Fatal(err)
			}
			<-chanSend
			chanReceive <- 1
		}(code)

		chanSend <- 1
	}

	//	阻塞，直到抓取所有
	for _, _ = range codes {
		<-chanReceive
	}

	return nil
}

//	抓取雅虎今日股票数据
func YahooTodayStock(code string) error {

	//	抓取数据
	raw, err := io.GetString(yahooQueryUrl(code))
	if err != nil {
		return err
	}

	//	保存原始数据
	dataDir, err := config.GetDataDir()
	if err != nil {
		return err
	}

	fileName := fmt.Sprintf("%s_raw.txt", time.Now().Format("20060102"))
	filePath := filepath.Join(dataDir, code, fileName)

	return io.WriteString(filePath, raw)
}

//	过早雅虎查询url
func yahooQueryUrl(code string) string {

	pattern := "https://finance-yql.media.yahoo.com/v7/finance/chart/%s?period2=%d&period1=%d&interval=1m&indicators=quote&includeTimestamps=true&includePrePost=true&events=div%7Csplit%7Cearn&corsDomain=finance.yahoo.com"
	now := time.Now()
	start := now.Truncate(time.Hour * 24)
	end := start.Add(time.Hour * 24)

	return fmt.Sprintf(pattern, code, end.Unix(), start.Unix())
}
