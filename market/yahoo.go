package market

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/nzai/stockrecorder/config"
	"github.com/nzai/stockrecorder/io"
)

//	从雅虎财经获取上市公司分时数据
func DownloadCompanyDaily(marketName, companyCode, queryCode string, day time.Time) error {
	//	文件保存路径
	dataDir := config.GetDataDir()
	fileName := fmt.Sprintf("%s_raw.txt", day.Format("20060102"))
	filePath := filepath.Join(dataDir, marketName, companyCode, fileName)

	//	如果文件已存在就忽略
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		//	如果不存在就抓取并保存
		start := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())
		end := start.Add(time.Hour * 24)

		pattern := "https://finance-yql.media.yahoo.com/v7/finance/chart/%s?period2=%d&period1=%d&interval=1m&indicators=quote&includeTimestamps=true&includePrePost=true&events=div%7Csplit%7Cearn&corsDomain=finance.yahoo.com"
		url := fmt.Sprintf(pattern, queryCode, end.Unix(), start.Unix())

		html, err := io.DownloadStringRetry(url, retryTimes, retryIntervalSeconds)
		if err != nil {
			return err
		}

		//	写入文件
		return io.WriteString(filePath, html)
	}

	return nil
}

type YahooJson struct {
	Chart YahooChart `json:"chart"`
}

type YahooChart struct {
	Result []YahooResult `json:"result"`
	Err    YahooError    `json:"error"`
}

type YahooError struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

type YahooResult struct {
	Meta       YahooMeta       `json:"meta"`
	Timestamp  []int64         `json:"timestamp"`
	Indicators YahooIndicators `json:"indicators"`
}

type YahooMeta struct {
	Currency             string              `json:"currency"`
	Symbol               string              `json:"symbol"`
	ExchangeName         string              `json:"exchangeName"`
	InstrumentType       string              `json:"instrumentType"`
	FirstTradeDate       int64               `json:"firstTradeDate"`
	GMTOffset            int                 `json:"gmtoffset"`
	Timezone             string              `json:"timezone"`
	PreviousClose        float32             `json:"previousClose"`
	Scale                int                 `json:"scale"`
	CurrentTradingPeriod YahooTradingPeroid  `json:"currentTradingPeriod"`
	TradingPeriods       YahooTradingPeroids `json:"tradingPeriods"`
	DataGranularity      string              `json:"dataGranularity"`
	ValidRanges          []string            `json:"validRanges"`
}

type YahooTradingPeroid struct {
	Pre     YahooTradingPeroidSection `json:"pre"`
	Regular YahooTradingPeroidSection `json:"regular"`
	Post    YahooTradingPeroidSection `json:"post"`
}

type YahooTradingPeroids struct {
	Pres     [][]YahooTradingPeroidSection `json:"pre"`
	Regulars [][]YahooTradingPeroidSection `json:"regular"`
	Posts    [][]YahooTradingPeroidSection `json:"post"`
}

type YahooTradingPeroidSection struct {
	Timezone  string `json:"timezone"`
	Start     int64  `json:"start"`
	End       int64  `json:"end"`
	GMTOffset int    `json:"gmtoffset"`
}

type YahooIndicators struct {
	Quotes []YahooQuote `json:"quote"`
}

type YahooQuote struct {
	Open   []float32 `json:"open"`
	Close  []float32 `json:"close"`
	High   []float32 `json:"high"`
	Low    []float32 `json:"low"`
	Volume []int64   `json:"volume"`
}