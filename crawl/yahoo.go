package crawl

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/nzai/stockrecorder/config"
	"github.com/nzai/stockrecorder/io"
)

const (
	gccount           = 16
	retryCount        = 5
	retryDelay        = 10 * time.Minute
	configKeyTimeZone = "timezone"
	//	雅虎财经的历史分时数据没有超过60天的
	lastestDays = 60
)

//	抓取雅虎今日数据
func marketAll(market string) error {
	//	股票编码
	codes := config.GetArray(market, "codes")
	if len(codes) == 0 {
		return errors.New(fmt.Sprintf("市场[%s]的codes配置有误", market))
	}

	chanSend := make(chan int, gccount)
	chanReceive := make(chan int)

	for _, code := range codes {
		//	并发抓取
		go func(c string) {
			err := stockAll(market, c)
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

	close(chanSend)
	close(chanReceive)

	return nil
}

//	抓取股票所有数据
func stockAll(market, code string) error {

	location := time.Local
	timezone := config.GetString(market, configKeyTimeZone, "")
	if timezone != "" {
		loc, err := time.LoadLocation(timezone)
		if err != nil {
			return err
		}
		location = loc
	}
	now := time.Now().In(location)
	today := now.Truncate(time.Hour * 24)

	//	定时器今天
	go func() {
		ticker := time.NewTicker(time.Hour * 24)
		log.Printf("已启动%s的定时抓取任务", code)
		for _ = range ticker.C {
			now = time.Now().In(location)
			today = now.Truncate(time.Hour * 24)

			err := stockToday(code, today)
			if err != nil {
				log.Fatalf("抓取%s在%s的数据出错:%v", code, today.Format("20060102"), err)
			}
		}
	}()

	//	历史
	return stockHistory(code, today)
}

//	抓取今天
func stockToday(code string, today time.Time) error {
	log.Printf("%s在%s分时数据抓取任务-开始", code, today.Format("20060102"))
	//	保存原始数据
	day := today.Add(-time.Hour * 24)
	for try := 0; try < retryCount; try++ {
		//	抓取数据
		err := thatDay(code, day)
		if err != nil {
			log.Fatalf("[%d]抓取%s在%s的数据出错:%v", try, code, today.Format("20060102"), err)
			time.Sleep(retryDelay)
		}

		log.Printf("%s在%s分时数据抓取任务-结束", code, today.Format("20060102"))
		return nil
	}

	return errors.New(fmt.Sprintf("%s在%s分时数据抓取任务失败", code, today.Format("20060102")))
}

//	抓取历史
func stockHistory(code string, today time.Time) error {
	log.Printf("%s历史分时数据抓取任务-开始", code)
	//	保存原始数据
	day := today.Add(-time.Hour * 24)
	for index := 0; index < lastestDays; index++ {

		for try := 0; try < retryCount; try++ {
			//	抓取数据
			err := thatDay(code, day)
			if err == nil {
				//	往前一天递推
				day = day.Add(-time.Hour * 24)
				break
			} else {
				log.Fatalf("[%d]抓取%s在%s的数据出错:%v", try, code, today.Format("20060102"), err)
				time.Sleep(retryDelay)
			}
		}
	}
	log.Printf("%s历史分时数据抓取任务-结束", code)
	return nil
}

//	抓取某一天
func thatDay(code string, day time.Time) error {

	//	文件保存路径
	dataDir := config.GetDataDir()
	fileName := fmt.Sprintf("%s_raw.txt", day.Format("20060102"))
	filePath := filepath.Join(dataDir, code, fileName)

	//	如果文件已存在就忽略
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		//	如果不存在就抓取并保存
		start := day.Truncate(time.Hour * 24)
		end := start.Add(time.Hour * 24)
		html, err := peroid(code, start, end)
		if err != nil {
			return err
		}

		//	写入文件
		return io.WriteString(filePath, html)
	}

	return nil
}

//	抓取一段时间
func peroid(code string, start, end time.Time) (string, error) {

	pattern := "https://finance-yql.media.yahoo.com/v7/finance/chart/%s?period2=%d&period1=%d&interval=1m&indicators=quote&includeTimestamps=true&includePrePost=true&events=div%7Csplit%7Cearn&corsDomain=finance.yahoo.com"
	url := fmt.Sprintf(pattern, code, end.Unix(), start.Unix())

	//	抓取数据
	return io.GetString(url)
}
