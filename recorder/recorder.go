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
	now := mr.marketNow()
	duration := mr.durationToNextDay(now)

	// 抓取历史数据
	go func(todayZero time.Time) {
		log.Printf("[%s] 获取历史数据开始", mr.Name())
		err := mr.crawlHistoryData(todayZero)
		if err != nil {
			log.Printf("[%s] 获取历史数据时发生错误: %v", mr.Name(), err)
			return
		}
		log.Printf("[%s] 获取历史数据结束", mr.Name())
	}(now.Add(duration).AddDate(0, 0, -1))

	// 持续抓取每日数据
	for {
		log.Printf("[%s] 定时任务已启动，将于%s后激活下一次任务", mr.Name(), duration.String())
		<-time.After(duration)
		yesterday := mr.marketNow().AddDate(0, 0, -1)
		log.Printf("[%s] 获取%s的数据开始", mr.Name(), yesterday.Format(datePattern))
		err := mr.crawlYesterdayData(yesterday)
		if err != nil {
			log.Printf("[%s] 获取%s的数据时发生错误: %v", mr.Name(), yesterday.Format(datePattern), err)
		} else {
			log.Printf("[%s] 获取%s的数据结束", mr.Name(), yesterday.Format(datePattern))
		}

		// 每天调整延时时长，保证长时间运行的时间精度
		duration = mr.durationToNextDay(mr.marketNow())
	}
}

// marketNow 市场所处时区当前时间
func (mr marketRecorder) marketNow() time.Time {
	now := time.Now()

	//	获取市场所在时区
	location, err := time.LoadLocation(mr.Market.Timezone())
	if err != nil {
		return now
	}

	return now.In(location)
}

// durationToNextDay 现在到明天0点的时间间隔
func (mr marketRecorder) durationToNextDay(now time.Time) time.Duration {

	year, month, day := now.AddDate(0, 0, 1).Date()

	// 市场所处时区的下一个0点
	marketTomorrowZero := time.Date(year, month, day, 0, 0, 0, 0, now.Location())

	return marketTomorrowZero.Sub(now)
}

// crawlHistoryData 抓取历史数据
func (mr marketRecorder) crawlHistoryData(todayZero time.Time) error {

	// 起始日期(含)
	date := todayZero.Add(-mr.source.Expiration())
	log.Printf("[%s]抓取历史数据起始日期: %s  结束日期: %s", mr.Name(), date.Format(datePattern), todayZero.Format(datePattern))

	// 获取上市公司
	companies, err := mr.Market.Companies()
	if err != nil {
		return err
	}
	log.Printf("[%s] 共有%d家上市公司", mr.Name(), len(companies))

	for date.Before(todayZero) {

		// 避免重复记录
		exists, err := mr.store.Exists(mr.Market, date)
		if err != nil {
			return err
		}

		if !exists {
			// 抓取那一天的报价
			err = mr.crawl(companies, date)
			if err != nil {
				return err
			}
		}

		// 后移一天
		date = date.AddDate(0, 0, 1)
	}

	return nil
}

// crawlYesterdayData 每天0点抓取前一天的数据
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

	_, offset := date.Zone()

	dailyQuote := market.DailyQuote{
		Market:    mr.Market,
		Date:      date,
		UTCOffset: offset,
	}

	for _, company := range companies {

		go func(_market market.Market, _company market.Company, _date time.Time) {
			quote, err := mr.source.Crawl(_market, _company, _date)
			if err == nil {
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
