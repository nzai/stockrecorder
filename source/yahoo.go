package source

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/nzai/go-utility/net"
	"github.com/nzai/stockrecorder/market"
)

// YahooFinance 雅虎财经数据源
type YahooFinance struct{}

// NewYahooFinance 新建雅虎财经数据源
func NewYahooFinance() YahooFinance {
	return YahooFinance{}
}

// Expiration 最早能查到60天前的数据
func (yahoo YahooFinance) Expiration() time.Duration {
	return time.Hour * 24 * 30
}

// Crawl 获取公司每天的报价
func (yahoo YahooFinance) Crawl(_market market.Market, company market.Company, date time.Time) (*market.CompanyDailyQuote, error) {

	// 起止时间
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	end := start.AddDate(0, 0, 1)

	pattern := "https://query2.finance.yahoo.com/v8/finance/chart/%s?period2=%d&period1=%d&interval=1m&indicators=quote&includeTimestamps=true&includePrePost=true&events=div%%7Csplit%%7Cearn&corsDomain=finance.yahoo.com"
	url := fmt.Sprintf(pattern, _market.YahooQueryCode(company), end.Unix(), start.Unix())

	// 查询Yahoo财经接口,返回股票分时数据
	str, err := net.DownloadStringRetry(url, yahoo.RetryCount(), yahoo.RetryInterval())
	if err != nil {
		return nil, err
	}

	// 解析Json
	quote := &YahooQuote{}
	err = json.Unmarshal([]byte(str), &quote)
	if err != nil {
		return nil, err
	}

	// 校验
	err = yahoo.valid(quote)
	if err != nil {
		return nil, err
	}

	// 解析
	return yahoo.parse(_market, company, date, quote)
}

// valid 校验
func (yahoo YahooFinance) valid(quote *YahooQuote) error {

	// 有错
	if quote.Chart.Err != nil {
		return errors.New(quote.Chart.Err.Description)
	}

	// Result为空
	if quote.Chart.Result == nil || len(quote.Chart.Result) == 0 {
		return errors.New("quote.Chart.Result is null")
	}

	// Quotes为空
	if quote.Chart.Result[0].Indicators.Quotes == nil || len(quote.Chart.Result[0].Indicators.Quotes) == 0 {
		return errors.New("quote.Chart.Result[0].Indicators.Quotes is null")
	}

	result, _quote := quote.Chart.Result[0], quote.Chart.Result[0].Indicators.Quotes[0]

	// Quotes数量不正确
	if len(result.Timestamp) != len(_quote.Open) ||
		len(result.Timestamp) != len(_quote.Close) ||
		len(result.Timestamp) != len(_quote.High) ||
		len(result.Timestamp) != len(_quote.Low) ||
		len(result.Timestamp) != len(_quote.Volume) {
		return errors.New("Quotes数量不正确")
	}

	// TradingPeriods数量不正确
	if len(result.Meta.TradingPeriods.Pres) == 0 ||
		len(result.Meta.TradingPeriods.Pres[0]) == 0 ||
		len(result.Meta.TradingPeriods.Posts) == 0 ||
		len(result.Meta.TradingPeriods.Posts[0]) == 0 ||
		len(result.Meta.TradingPeriods.Regulars) == 0 ||
		len(result.Meta.TradingPeriods.Regulars[0]) == 0 {
		return errors.New("TradingPeriods数量不正确")
	}

	return nil
}

// parse 解析结果
func (yahoo YahooFinance) parse(_market market.Market, company market.Company, date time.Time, quote *YahooQuote) (*market.CompanyDailyQuote, error) {

	companyDailyQuote := market.CompanyDailyQuote{Company: company}

	periods, _quote := quote.Chart.Result[0].Meta.TradingPeriods, quote.Chart.Result[0].Indicators.Quotes[0]
	for index, ts := range quote.Chart.Result[0].Timestamp {

		//	如果全为0就忽略
		if _quote.Open[index] == 0 && _quote.Close[index] == 0 && _quote.High[index] == 0 && _quote.Low[index] == 0 && _quote.Volume[index] == 0 {
			continue
		}

		var series *market.QuoteSeries

		//	Pre, Regular, Post
		if ts >= periods.Pres[0][0].Start && ts < periods.Pres[0][0].End {
			series = &companyDailyQuote.Pre
		} else if ts >= periods.Regulars[0][0].Start && ts < periods.Regulars[0][0].End {
			series = &companyDailyQuote.Regular
		} else if ts >= periods.Posts[0][0].Start && ts < periods.Posts[0][0].End {
			series = &companyDailyQuote.Post
		} else {
			continue
		}

		series.Count++
		series.Timestamp = append(series.Timestamp, uint32(ts))
		series.Open = append(series.Open, uint32(_quote.Open[index]*100))
		series.Close = append(series.Close, uint32(_quote.Close[index]*100))
		series.Max = append(series.Max, uint32(_quote.High[index]*100))
		series.Min = append(series.Min, uint32(_quote.Low[index]*100))
		series.Volume = append(series.Volume, uint32(_quote.Volume[index]))
	}

	return &companyDailyQuote, nil
}

// ParallelMax 最大并发数
func (yahoo YahooFinance) ParallelMax() int {
	return 32
}

// RetryCount 失败重试次数
func (yahoo YahooFinance) RetryCount() int {
	return 5
}

// RetryInterval 失败重试时间间隔
func (yahoo YahooFinance) RetryInterval() time.Duration {
	return time.Second * 10
}

// YahooQuote 雅虎财经返回的json
type YahooQuote struct {
	Chart struct {
		Result []struct {
			Meta struct {
				Currency             string  `json:"currency"`
				Symbol               string  `json:"symbol"`
				ExchangeName         string  `json:"exchangeName"`
				InstrumentType       string  `json:"instrumentType"`
				FirstTradeDate       int64   `json:"firstTradeDate"`
				GMTOffset            int64   `json:"gmtoffset"`
				Timezone             string  `json:"timezone"`
				PreviousClose        float32 `json:"previousClose"`
				Scale                int     `json:"scale"`
				CurrentTradingPeriod struct {
					Pre struct {
						Timezone  string `json:"timezone"`
						Start     int64  `json:"start"`
						End       int64  `json:"end"`
						GMTOffset int64  `json:"gmtoffset"`
					} `json:"pre"`
					Regular struct {
						Timezone  string `json:"timezone"`
						Start     int64  `json:"start"`
						End       int64  `json:"end"`
						GMTOffset int64  `json:"gmtoffset"`
					} `json:"regular"`
					Post struct {
						Timezone  string `json:"timezone"`
						Start     int64  `json:"start"`
						End       int64  `json:"end"`
						GMTOffset int64  `json:"gmtoffset"`
					} `json:"post"`
				} `json:"currentTradingPeriod"`
				TradingPeriods struct {
					Pres [][]struct {
						Timezone  string `json:"timezone"`
						Start     int64  `json:"start"`
						End       int64  `json:"end"`
						GMTOffset int64  `json:"gmtoffset"`
					} `json:"pre"`
					Regulars [][]struct {
						Timezone  string `json:"timezone"`
						Start     int64  `json:"start"`
						End       int64  `json:"end"`
						GMTOffset int64  `json:"gmtoffset"`
					} `json:"regular"`
					Posts [][]struct {
						Timezone  string `json:"timezone"`
						Start     int64  `json:"start"`
						End       int64  `json:"end"`
						GMTOffset int64  `json:"gmtoffset"`
					} `json:"post"`
				} `json:"tradingPeriods"`
				DataGranularity string   `json:"dataGranularity"`
				ValidRanges     []string `json:"validRanges"`
			} `json:"meta"`
			Timestamp  []int64 `json:"timestamp"`
			Indicators struct {
				Quotes []struct {
					Open   []float32 `json:"open"`
					Close  []float32 `json:"close"`
					High   []float32 `json:"high"`
					Low    []float32 `json:"low"`
					Volume []int64   `json:"volume"`
				} `json:"quote"`
			} `json:"indicators"`
		} `json:"result"`
		Err *struct {
			Code        string `json:"code"`
			Description string `json:"description"`
		} `json:"error"`
	} `json:"chart"`
}
