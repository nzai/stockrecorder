package market

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/nzai/stockrecorder/db"
	"github.com/nzai/stockrecorder/io"
	"gopkg.in/mgo.v2/bson"
)

type Raw60 struct {
	Code    string
	Market  string
	Date    time.Time
	Json    string
	Status  int
	Message string
}

var (
	//	存储队列
	saveQueue chan Raw60      = nil
	dict      map[string]bool = nil
)

//	载入保存过的Raw60
func loadSavedRaw60() error {
	//	连接数据库
	session, err := db.Get()
	if err != nil {
		return fmt.Errorf("[DB]\t获取数据库连接失败:%s", err.Error())
	}
	defer session.Close()

	//	获取所有Raw
	var raws []Raw60
	err = session.DB("stock").C("Raw60").Find(nil).Select(bson.M{"market": 1, "code": 1, "date": 1}).All(&raws)
	if err != nil {
		return fmt.Errorf("[DB]\t获取Raw60失败:%s", err.Error())
	}

	saveQueue = make(chan Raw60)
	dict = make(map[string]bool)
	for _, raw := range raws {
		addDict(raw)
	}

	log.Printf("[DB]\t数据库中已经保存了%d条Raw60记录", len(raws))

	//	保存协程
	go saveRaw60()

	return nil
}

//	添加到字典
func addDict(raw Raw60) {
	dict[raw.Market+raw.Code+strconv.FormatInt(raw.Date.Unix(), 10)] = true
}

//	以队列的方式保存到数据库
func saveRaw60() {
	session, err := db.Get()
	if err != nil {
		log.Printf("[DB]\t获取数据库连接失败:%s", err.Error())
		return
	}
	defer session.Close()

	collection := session.DB("stock").C("Raw60")
	for {
		raw := <-saveQueue
		//	所有新增的记录都是未处理状态
		raw.Status = 0

		rawlist := make([]interface{}, 0)
		rawlist = append(rawlist, raw)

		//	如果队列长度超过1，就批量新增
		queueLength := len(saveQueue)
		if queueLength > 0 {
			//	读取队列
			for index := 0; index < queueLength; index++ {
				raw := <-saveQueue
				//	所有新增的记录都是未处理状态
				raw.Status = 0

				rawlist = append(rawlist, raw)
			}
		}

		var err error
		for times := retryTimes - 1; times >= 0; times-- {
			//	保存到数据库
			err = collection.Insert(rawlist...)
			if err == nil {
				break
			}

			if times > 0 {
				//	延时
				time.Sleep(time.Duration(retryIntervalSeconds) * time.Second)
			}
		}

		if err != nil {
			log.Printf("[DB]\t保存[%s %s %s]出错,已经重试%d次,不再重试:%s", raw.Market, raw.Code, raw.Date.Format("2006-01-02 15:04:05"), retryTimes, err.Error())
		} else {
			for _, ri := range rawlist {
				addDict(ri.(Raw60))
			}
		}
	}
}

//	判断Raw60是否存在
func raw60Exists(marketName, companyCode string, date time.Time) bool {
	_, found := dict[marketName+companyCode+strconv.FormatInt(date.Unix(), 10)]
	return found
}

//	从雅虎财经获取上市公司分时数据
func DownloadCompanyDaily(marketName, companyCode, queryCode string, day time.Time) error {

	//	检查数据库是否解析过,解析过的不再重复解析
	found := raw60Exists(marketName, companyCode, day)
	if found {
		return nil
	}

	//	如果不存在就抓取
	start := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())
	end := start.Add(time.Hour * 24)

	pattern := "https://finance-yql.media.yahoo.com/v7/finance/chart/%s?period2=%d&period1=%d&interval=1m&indicators=quote&includeTimestamps=true&includePrePost=true&events=div%7Csplit%7Cearn&corsDomain=finance.yahoo.com"
	url := fmt.Sprintf(pattern, queryCode, end.Unix(), start.Unix())

	//	查询Yahoo财经接口,返回股票分时数据
	content, err := io.DownloadStringRetry(url, retryTimes, retryIntervalSeconds)
	if err != nil {
		return err
	}

	raw := Raw60{
		Market:  marketName,
		Code:    companyCode,
		Date:    day.UTC(),
		Json:    content,
		Status:  0,
		Message: ""}

	//	保存(加入保存队列)
	saveQueue <- raw

	return nil
}

//type YahooJson struct {
//	Chart YahooChart `json:"chart"`
//}

//type YahooChart struct {
//	Result []YahooResult `json:"result"`
//	Err    *YahooError   `json:"error"`
//}

//type YahooError struct {
//	Code        string `json:"code"`
//	Description string `json:"description"`
//}

//type YahooResult struct {
//	Meta       YahooMeta       `json:"meta"`
//	Timestamp  []int64         `json:"timestamp"`
//	Indicators YahooIndicators `json:"indicators"`
//}

//type YahooMeta struct {
//	Currency             string              `json:"currency"`
//	Symbol               string              `json:"symbol"`
//	ExchangeName         string              `json:"exchangeName"`
//	InstrumentType       string              `json:"instrumentType"`
//	FirstTradeDate       int64               `json:"firstTradeDate"`
//	GMTOffset            int                 `json:"gmtoffset"`
//	Timezone             string              `json:"timezone"`
//	PreviousClose        float32             `json:"previousClose"`
//	Scale                int                 `json:"scale"`
//	CurrentTradingPeriod YahooTradingPeroid  `json:"currentTradingPeriod"`
//	TradingPeriods       YahooTradingPeroids `json:"tradingPeriods"`
//	DataGranularity      string              `json:"dataGranularity"`
//	ValidRanges          []string            `json:"validRanges"`
//}

//type YahooTradingPeroid struct {
//	Pre     YahooTradingPeroidSection `json:"pre"`
//	Regular YahooTradingPeroidSection `json:"regular"`
//	Post    YahooTradingPeroidSection `json:"post"`
//}

//type YahooTradingPeroids struct {
//	Pres     [][]YahooTradingPeroidSection `json:"pre"`
//	Regulars [][]YahooTradingPeroidSection `json:"regular"`
//	Posts    [][]YahooTradingPeroidSection `json:"post"`
//}

//type YahooTradingPeroidSection struct {
//	Timezone  string `json:"timezone"`
//	Start     int64  `json:"start"`
//	End       int64  `json:"end"`
//	GMTOffset int    `json:"gmtoffset"`
//}

//type YahooIndicators struct {
//	Quotes []YahooQuote `json:"quote"`
//}

//type YahooQuote struct {
//	Open   []float32 `json:"open"`
//	Close  []float32 `json:"close"`
//	High   []float32 `json:"high"`
//	Low    []float32 `json:"low"`
//	Volume []int64   `json:"volume"`
//}

//type Raw60 struct {
//	Code    string
//	Market  string
//	Date    time.Time
//	Json    string
//	Status  int
//	Message string
//}

////	解析雅虎Json
//func ParseDailyYahooJson(marketName, companyCode string, date time.Time, buffer []byte) (*db.DailyAnalyzeResult, error) {

//	yj := &YahooJson{}
//	err := json.Unmarshal(buffer, &yj)
//	if err != nil {
//		return nil, fmt.Errorf("解析雅虎Json发生错误: %s", err)
//	}

//	result := &db.DailyAnalyzeResult{
//		DailyResult: db.DailyResult{
//			Code:    companyCode,
//			Market:  marketName,
//			Date:    date,
//			Error:   false,
//			Message: ""},
//		Pre:     make([]db.Peroid60, 0),
//		Regular: make([]db.Peroid60, 0),
//		Post:    make([]db.Peroid60, 0)}

//	//	检查数据
//	err = validateDailyYahooJson(yj)
//	if err != nil {
//		result.DailyResult.Error = true
//		result.DailyResult.Message = err.Error()

//		return result, nil
//	}

//	periods, quote := yj.Chart.Result[0].Meta.TradingPeriods, yj.Chart.Result[0].Indicators.Quotes[0]
//	for index, ts := range yj.Chart.Result[0].Timestamp {

//		p := db.Peroid60{
//			Code:   companyCode,
//			Market: marketName,
//			Start:  time.Unix(ts, 0),
//			End:    time.Unix(ts+60, 0),
//			Open:   quote.Open[index],
//			Close:  quote.Close[index],
//			High:   quote.High[index],
//			Low:    quote.Low[index],
//			Volume: quote.Volume[index]}

//		//	Pre, Regular, Post
//		if ts >= periods.Pres[0][0].Start && ts < periods.Pres[0][0].End {
//			result.Pre = append(result.Pre, p)
//		} else if ts >= periods.Regulars[0][0].Start && ts < periods.Regulars[0][0].End {
//			result.Regular = append(result.Regular, p)
//		} else if ts >= periods.Posts[0][0].Start && ts < periods.Posts[0][0].End {
//			result.Post = append(result.Regular, p)
//		}
//	}

//	return result, nil
//}

////	验证雅虎Json
//func validateDailyYahooJson(yj *YahooJson) error {

//	if yj.Chart.Err != nil {
//		return fmt.Errorf("[%s]%s", yj.Chart.Err.Code, yj.Chart.Err.Description)
//	}

//	if yj.Chart.Result == nil || len(yj.Chart.Result) == 0 {
//		return fmt.Errorf("Result为空")
//	}

//	if yj.Chart.Result[0].Indicators.Quotes == nil || len(yj.Chart.Result[0].Indicators.Quotes) == 0 {
//		return fmt.Errorf("Quotes为空")
//	}

//	result, quote := yj.Chart.Result[0], yj.Chart.Result[0].Indicators.Quotes[0]
//	if len(result.Timestamp) != len(quote.Open) ||
//		len(result.Timestamp) != len(quote.Close) ||
//		len(result.Timestamp) != len(quote.High) ||
//		len(result.Timestamp) != len(quote.Low) ||
//		len(result.Timestamp) != len(quote.Volume) {
//		return fmt.Errorf("Quotes数量不正确")
//	}

//	if len(result.Meta.TradingPeriods.Pres) == 0 ||
//		len(result.Meta.TradingPeriods.Pres[0]) == 0 ||
//		len(result.Meta.TradingPeriods.Posts) == 0 ||
//		len(result.Meta.TradingPeriods.Posts[0]) == 0 ||
//		len(result.Meta.TradingPeriods.Regulars) == 0 ||
//		len(result.Meta.TradingPeriods.Regulars[0]) == 0 {
//		return fmt.Errorf("TradingPeriods数量不正确")
//	}
//	return nil
//}

//	保存到文件
//func saveDaily(marketName, companyCode string, day time.Time, buffer []byte) error {

//	//	文件保存路径
//	fileName := fmt.Sprintf("%s_raw.txt", day.Format("20060102"))
//	filePath := filepath.Join(config.Get().DataDir, marketName, companyCode, fileName)

//	//	不覆盖原文件
//	_, err := os.Stat(filePath)
//	if os.IsNotExist(err) {
//		return io.WriteBytes(filePath, buffer)
//	}

//	return nil
//}
