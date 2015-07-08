package stock

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nzai/stockrecorder/config"
	"github.com/nzai/stockrecorder/io"
)

func GetStocks() []string {
	value := config.GetString("stock", "codes", "AAPL")
	stocks := strings.Split(value, ",")

	return stocks
}

func GetToday() error {

	log.Print("更新股指任务-启动")
	codes := GetStocks()
	channels := make(chan int, len(codes))

	for _, code := range codes {
		go func(c string) {
			err := getStockToday(c)
			if err != nil {
				log.Fatal(err)
			}

			channels <- 0
		}(code)
	}

	for _, _ = range codes {
		<-channels
	}
	log.Print("更新股指任务-结束")
	return nil
}

func getStockToday(code string) error {

	pattern := "https://finance-yql.media.yahoo.com/v7/finance/chart/%s?period2=%d&period1=%d&interval=1m&indicators=quote&includeTimestamps=true&includePrePost=true&events=div%7Csplit%7Cearn&corsDomain=finance.yahoo.com"
	start, end := getTimeRange(code)
	url := fmt.Sprintf(pattern, code, end.Unix(), start.Unix())
	html, err := io.GetString(url)
	if err != nil {
		return err
	}

	log.Print(html)

	return nil
}

//	获取股票的交易起始时间
func getTimeRange(code string) (time.Time, time.Time) {

	now := time.Now().UTC()
	if strings.HasSuffix(code, "SS") {
		//	A股每天0930-1500
		return time.Date(now.Year(), now.Month(), now.Day(), 1, 30, 0, 0, now.Location()),
			time.Date(now.Year(), now.Month(), now.Day(), 7, 00, 0, 0, now.Location())
	}

	//	美股每天2230-0400
	return time.Date(now.Year(), now.Month(), now.Day(), 13, 30, 0, 0, now.Location()),
		time.Date(now.Year(), now.Month(), now.Day(), 20, 00, 0, 0, now.Location())
}

func parseHtml(html string) ([]string, error) {

}

func substr(html)
