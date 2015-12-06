package market

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nzai/go-utility/net"
)

type YahooJson struct {
	Chart YahooChart `json:"chart"`
}

type YahooChart struct {
	Result []YahooResult `json:"result"`
	Err    *YahooError   `json:"error"`
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

type Peroid60 struct {
	Market string
	Code   string
	Time   time.Time
	Open   float32
	Close  float32
	High   float32
	Low    float32
	Volume int64
}

type ParseResult struct {
	Success bool
	Message string
	Pre     []Peroid60
	Regular []Peroid60
	Post    []Peroid60
}

//	从雅虎财经获取上市公司分时数据
func downloadCompanyDaily(market Market, code, queryCode string, date time.Time) (string, error) {

	//	如果不存在就抓取
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	end := start.Add(time.Hour * 24)

	pattern := "https://finance-yql.media.yahoo.com/v7/finance/chart/%s?period2=%d&period1=%d&interval=1m&indicators=quote&includeTimestamps=true&includePrePost=true&events=div%7Csplit%7Cearn&corsDomain=finance.yahoo.com"
	url := fmt.Sprintf(pattern, queryCode, end.Unix(), start.Unix())

	//	查询Yahoo财经接口,返回股票分时数据
	return net.DownloadStringRetry(url, retryTimes, retryIntervalSeconds)
}

//	处理雅虎Json
func processDailyYahooJson(market Market, code string, date time.Time, buffer []byte) (*ParseResult, error) {

	//	解析Json
	yj := &YahooJson{}
	err := json.Unmarshal(buffer, &yj)
	if err != nil {
		return nil, err
	}

	//	检查数据
	err = validateDailyYahooJson(yj)
	if err != nil {
		return &ParseResult{false, err.Error(), nil, nil, nil}, nil
	}

	//	服务所在时区与市场所在时区的时间差(秒)
	timezoneOffset := marketOffset[market.Name()]

	pre := make([]Peroid60, 0)
	regular := make([]Peroid60, 0)
	post := make([]Peroid60, 0)

	periods, quote := yj.Chart.Result[0].Meta.TradingPeriods, yj.Chart.Result[0].Indicators.Quotes[0]
	for index, ts := range yj.Chart.Result[0].Timestamp {

		p := Peroid60{
			Code:   code,
			Market: market.Name(),
			Time:   time.Unix(ts+timezoneOffset, 0),
			Open:   quote.Open[index],
			Close:  quote.Close[index],
			High:   quote.High[index],
			Low:    quote.Low[index],
			Volume: quote.Volume[index]}

		//	如果全为0就忽略
		if p.Open == 0 && p.Close == 0 && p.High == 0 && p.Low == 0 && p.Volume == 0 {
			continue
		}

		//	Pre, Regular, Post
		if ts >= periods.Pres[0][0].Start && ts < periods.Pres[0][0].End {
			pre = append(pre, p)
		} else if ts >= periods.Regulars[0][0].Start && ts < periods.Regulars[0][0].End {
			regular = append(regular, p)
		} else if ts >= periods.Posts[0][0].Start && ts < periods.Posts[0][0].End {
			post = append(post, p)
		}
	}

	return &ParseResult{true, "", pre, regular, post}, nil
}

//	验证雅虎Json
func validateDailyYahooJson(yj *YahooJson) error {

	if yj.Chart.Err != nil {
		return fmt.Errorf("[%s]%s", yj.Chart.Err.Code, yj.Chart.Err.Description)
	}

	if yj.Chart.Result == nil || len(yj.Chart.Result) == 0 {
		return fmt.Errorf("Result为空")
	}

	if yj.Chart.Result[0].Indicators.Quotes == nil || len(yj.Chart.Result[0].Indicators.Quotes) == 0 {
		return fmt.Errorf("Quotes为空")
	}

	result, quote := yj.Chart.Result[0], yj.Chart.Result[0].Indicators.Quotes[0]
	if len(result.Timestamp) != len(quote.Open) ||
		len(result.Timestamp) != len(quote.Close) ||
		len(result.Timestamp) != len(quote.High) ||
		len(result.Timestamp) != len(quote.Low) ||
		len(result.Timestamp) != len(quote.Volume) {
		return fmt.Errorf("Quotes数量不正确")
	}

	if len(result.Meta.TradingPeriods.Pres) == 0 ||
		len(result.Meta.TradingPeriods.Pres[0]) == 0 ||
		len(result.Meta.TradingPeriods.Posts) == 0 ||
		len(result.Meta.TradingPeriods.Posts[0]) == 0 ||
		len(result.Meta.TradingPeriods.Regulars) == 0 ||
		len(result.Meta.TradingPeriods.Regulars[0]) == 0 {
		return fmt.Errorf("TradingPeriods数量不正确")
	}

	return nil
}
