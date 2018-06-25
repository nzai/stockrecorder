package source

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/nzai/go-utility/net"
	"github.com/nzai/stockrecorder/quote"
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
func (yahoo YahooFinance) Crawl(exchange *quote.Exchange, company *quote.Company, date time.Time) (*quote.CompanyDailyQuote, error) {

	// 起止时间
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	end := start.AddDate(0, 0, 1)

	pattern := "https://query2.finance.yahoo.com/v8/finance/chart/%s%s?period2=%d&period1=%d&interval=1m&indicators=quote&includeTimestamps=true&includePrePost=true&events=div%%7Csplit%%7Cearn&corsDomain=finance.yahoo.com"
	url := fmt.Sprintf(pattern, company.Code, exchange.YahooSuffix, end.Unix(), start.Unix())

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
	return yahoo.parse(company, quote)
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
func (yahoo YahooFinance) parse(company *quote.Company, yq *YahooQuote) (*quote.CompanyDailyQuote, error) {

	cdq := &quote.CompanyDailyQuote{
		Company: company,
		Pre:     make([]quote.Quote, 0),
		Regular: make([]quote.Quote, 0),
		Post:    make([]quote.Quote, 0),
	}

	periods, qs := yq.Chart.Result[0].Meta.TradingPeriods, yq.Chart.Result[0].Indicators.Quotes[0]
	for index, ts := range yq.Chart.Result[0].Timestamp {

		//	如果全为0就忽略
		if qs.Open[index] == 0 && qs.Close[index] == 0 && qs.High[index] == 0 && qs.Low[index] == 0 && qs.Volume[index] == 0 {
			continue
		}

		q := quote.Quote{
			Timestamp: uint64(ts),
			Open:      qs.Open[index],
			Close:     qs.Close[index],
			High:      qs.High[index],
			Low:       qs.Low[index],
			Volume:    uint64(qs.Volume[index]),
		}

		//	Pre, Regular, Post
		if ts >= periods.Pres[0][0].Start && ts < periods.Pres[0][0].End {
			cdq.Pre = append(cdq.Pre, q)
		} else if ts >= periods.Regulars[0][0].Start && ts < periods.Regulars[0][0].End {
			cdq.Pre = append(cdq.Regular, q)
		} else if ts >= periods.Posts[0][0].Start && ts < periods.Posts[0][0].End {
			cdq.Pre = append(cdq.Post, q)
		} else {
			continue
		}
	}

	return cdq, nil
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
					Pre     YahooPeroid `json:"pre"`
					Regular YahooPeroid `json:"regular"`
					Post    YahooPeroid `json:"post"`
				} `json:"currentTradingPeriod"`
				TradingPeriods struct {
					Pres     [][]YahooPeroid `json:"pre"`
					Regulars [][]YahooPeroid `json:"regular"`
					Posts    [][]YahooPeroid `json:"post"`
				} `json:"tradingPeriods"`
				DataGranularity string   `json:"dataGranularity"`
				ValidRanges     []string `json:"validRanges"`
			} `json:"meta"`
			Timestamp []uint64 `json:"timestamp"`
			Events    struct {
				Dividends map[uint64]YahooDividend `json:"dividends"`
				Splits    map[uint64]YahooSplits   `json:"splits"`
			} `json:"events"`
			Indicators struct {
				Quotes []struct {
					Open   []float32 `json:"open"`
					Close  []float32 `json:"close"`
					High   []float32 `json:"high"`
					Low    []float32 `json:"low"`
					Volume []uint64  `json:"volume"`
				} `json:"quote"`
			} `json:"indicators"`
		} `json:"result"`
		Err *struct {
			Code        string `json:"code"`
			Description string `json:"description"`
		} `json:"error"`
	} `json:"chart"`
}

// YahooPeroid 时间段
type YahooPeroid struct {
	Timezone  string `json:"timezone"`
	Start     uint64 `json:"start"`
	End       uint64 `json:"end"`
	GMTOffset int64  `json:"gmtoffset"`
}

// YahooDividend 股息
type YahooDividend struct {
	Amount float32 `json:"amount"`
	Date   uint64  `json:"date"`
}

// YahooSplits 拆股
type YahooSplits struct {
	Date        uint64 `json:"date"`
	Numerator   int    `json:"numerator"`
	Denominator int    `json:"denominator"`
	Ratio       string `json:"splitRatio"`
}
