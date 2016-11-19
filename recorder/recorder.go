package recorder

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/nzai/stockrecorder/market"
	"github.com/nzai/stockrecorder/source"
	"github.com/nzai/stockrecorder/store"
)

const (
	datePattern = "20060102"
)

// Recorder 股票记录器
type Recorder struct {
	source  source.Source   // 数据源
	store   store.Store     // 存储
	markets []market.Market // 市场
}

// NewRecorder 新建Recorder
func NewRecorder(source source.Source, store store.Store, markets ...market.Market) *Recorder {
	return &Recorder{source, store, markets}
}

// RunAndWait 执行
func (r Recorder) RunAndWait() {
	var wg sync.WaitGroup
	wg.Add(len(r.markets))

	for _, m := range r.markets {
		go func(m market.Market) {
			// 构造记录器
			mr := marketRecorder{r.source, r.store, m}
			// 启动
			mr.RunAndWait()
			wg.Done()
		}(m)
	}

	wg.Wait()
}

// marketRecorder 市场记录器
type marketRecorder struct {
	source        source.Source // 数据源
	store         store.Store   // 存储
	market.Market               // 市场
}

// RunAndWait 启动市场记录器
func (mr marketRecorder) RunAndWait() {

	// 获取市场所在地到明天零点的时间差
	now := time.Now()
	duration, err := mr.durationToNextDay(now)
	if err != nil {
		log.Printf("[%s] 获取市场所在地到明天零点的时间差时发生错误: %v", mr.Name(), err)
		return
	}

	// 抓取历史数据
	go func(_now time.Time, _duration time.Duration) {
		log.Printf("[%s] 获取历史数据开始", mr.Name())
		err = mr.crawlHistoryData(_now, _duration)
		if err != nil {
			log.Printf("[%s] 获取历史数据时发生错误: %v", mr.Name(), err)
		}
		log.Printf("[%s] 获取历史数据结束", mr.Name())
	}(now, duration)

	// 持续抓取每日数据
	for {
		log.Printf("[%s] 定时任务已启动，将于%s后激活下一次任务", mr.Name(), duration.String())
		now = <-time.After(duration)
		log.Printf("[%s] 获取%s的数据开始", mr.Name(), now.Format(datePattern))
		err = mr.crawlYesterdayData(now.Add(-time.Hour * 24))
		if err != nil {
			log.Printf("[%s] 获取%s的数据时发生错误: %v", mr.Name(), now.Format(datePattern), err)
		}
		log.Printf("[%s] 获取%s的数据结束", mr.Name(), now.Format(datePattern))

		// 每天调整延时时长，保证长时间运行的时间精度
		duration, err = mr.durationToNextDay(time.Now())
		if err != nil {
			log.Printf("[%s] 获取市场所在地到明天零点的时间差时发生错误: %v", mr.Name(), err)
			return
		}
	}
}

// durationToNextDay 现在到明天0点的时间间隔
func (mr marketRecorder) durationToNextDay(now time.Time) (time.Duration, error) {

	//	获取市场所在时区
	location, err := time.LoadLocation(mr.Timezone())
	if err != nil {
		return 0, err
	}

	//	市场所处时区当前时间
	marketNow := now.In(location)
	year, month, day := marketNow.AddDate(0, 0, 1).Date()

	// 市场所处时区的下一个0点
	marketTomorrowZero := time.Date(year, month, day, 0, 0, 0, 0, location)

	return marketTomorrowZero.Sub(marketNow), nil
}

// getHistoryData 抓取历史数据
func (mr marketRecorder) crawlHistoryData(now time.Time, dur time.Duration) error {

	// 获取上市公司
	companies, err := mr.Market.Companies()
	if err != nil {
		return err
	}
	log.Printf("[%s] 共有%d家上市公司", mr.Name(), len(companies))

	// 结束日期(昨天0点,不含)
	endDate := now.Add(dur).AddDate(0, 0, -2)
	// 起始日期(含)
	startDate := endDate.Add(-mr.source.Expiration())
	log.Printf("[%s]抓取历史数据起始日期: %s  结束日期: %s", mr.Name(), startDate.Format(datePattern), endDate.Format(datePattern))

	for !startDate.After(endDate) {

		// 避免重复记录
		recorded, err := mr.store.Exists(mr.Market, startDate)
		if err != nil {
			return err
		}

		if recorded {
			// 后移一天
			startDate = startDate.Add(time.Hour * 24)
			continue
		}

		// 抓取那一天的报价
		err = mr.crawl(companies, startDate)
		if err != nil {
			return err
		}

		// 后移一天
		startDate = startDate.Add(time.Hour * 24)
	}

	return nil
}

// startDailyTask 每天0点抓取前一天的数据
func (mr marketRecorder) crawlYesterdayData(yesterday time.Time) error {

	// 避免重复记录
	recorded, err := mr.store.Exists(mr.Market, yesterday)
	if err != nil || recorded {
		return err
	}

	// 获取上市公司
	companies, err := mr.Market.Companies()
	if err != nil {
		return err
	}
	log.Printf("[%s] 共有%d家上市公司", mr.Name(), len(companies))

	// 抓取
	return mr.crawl(companies, yesterday)
}

// crawl 抓取指定日期的市场报价
func (mr marketRecorder) crawl(companies []market.Company, date time.Time) error {

	ch := make(chan bool, mr.source.ParallelMax())
	defer close(ch)

	var wg sync.WaitGroup
	wg.Add(len(companies))

	dailyQuote := market.DailyQuote{
		Market: mr.Market,
		Date:   date,
	}

	for _, company := range companies {

		go func(_market market.Market, _company market.Company, _date time.Time) {
			quote, err := mr.source.Crawl(_market, _company, _date)
			if err != nil {
				if err != source.ErrQuoteInvalid {
					log.Print(err)
				}
			} else {
				dailyQuote.Quotes = append(dailyQuote.Quotes, *quote)
			}

			<-ch
			wg.Done()
		}(mr.Market, company, date)

		// 限流
		ch <- false
	}
	//	阻塞，直到抓取所有
	wg.Wait()

	// 保存
	err := mr.store.Save(dailyQuote)
	if err != nil {
		return fmt.Errorf("[%s] 保存上市公司在%s的分时数据时发生错误: %v", mr.Market.Name(), date.Format(datePattern), err)
	}

	log.Printf("[%s] 上市公司在%s的分时数据已经抓取结束", mr.Market.Name(), date.Format(datePattern))

	return nil
}
