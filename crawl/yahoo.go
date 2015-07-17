package crawl

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/nzai/stockrecorder/config"
	"github.com/nzai/stockrecorder/io"
)

const (
	gccount    = 16
	retryCount = 5
	retryDelay = 10 * time.Minute
)

//	抓取雅虎今日数据
func yahooToday(market string, endHour int) error {
	//	股票编码
	codes := config.GetArray(market, "codes")
	if len(codes) == 0 {
		return errors.New(fmt.Sprintf("市场[%s]的codes配置有误", market))
	}

	chanSend := make(chan int, gccount)
	chanReceive := make(chan int)

	for _, code := range codes {
		//	并发抓取
		go func(c string, h int) {

			for try := 0; try < retryCount; try++ {
				//	更新每只股票的历史
				err := yahooTodayStock(c, h)
				if err == nil {
					break
				}

				log.Fatal(err)
				time.Sleep(retryDelay)
			}

			<-chanSend
			chanReceive <- 1
		}(code, endHour)

		chanSend <- 1
	}

	//	阻塞，直到抓取所有
	for _, _ = range codes {
		<-chanReceive
	}

	close(chanSend)
	close(chanReceive)

	return nil
}

//	抓取雅虎今日股票数据
func yahooTodayStock(code string, endHour int) error {

	//	抓取数据
	raw, err := io.GetString(yahooQueryUrl(code, endHour))
	if err != nil {
		return err
	}

	//	保存原始数据
	dataDir := config.GetDataDir()

	fileName := fmt.Sprintf("%s_raw.txt", getCurrentDate(endHour).Format("20060102"))
	filePath := filepath.Join(dataDir, code, fileName)

	return io.WriteString(filePath, raw)
}

//	构造雅虎查询url
func yahooQueryUrl(code string, endHour int) string {

	start := getCurrentDate(endHour)
	end := start.Add(time.Hour * 24)

	pattern := "https://finance-yql.media.yahoo.com/v7/finance/chart/%s?period2=%d&period1=%d&interval=1m&indicators=quote&includeTimestamps=true&includePrePost=true&events=div%7Csplit%7Cearn&corsDomain=finance.yahoo.com"

	return fmt.Sprintf(pattern, code, end.Unix(), start.Unix())
}

//	当前日期：获取当天0点的时间，跨天的返回前一天
func getCurrentDate(endHour int) time.Time {

	now := time.Now().UTC()
	today := now.Add(-time.Hour * time.Duration(endHour))
	return today.Truncate(time.Hour * 24)
}
