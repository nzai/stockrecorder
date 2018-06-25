package recorder

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/nzai/stockrecorder/quote"

	"github.com/nzai/stockrecorder/provider"
	"github.com/nzai/stockrecorder/source"
	"github.com/nzai/stockrecorder/store"
)

const (
	datePattern = "20060102"
)

// Recorder 股票记录器
type Recorder struct {
	source    source.Source           // 数据源
	store     store.Store             // 存储
	providers []provider.DataProvider // 提供者
}

// NewRecorder 新建Recorder
func NewRecorder(source source.Source, store store.Store, providers ...provider.DataProvider) *Recorder {
	return &Recorder{source, store, providers}
}

// Run 执行
func (r Recorder) Run() *sync.WaitGroup {
	wg := new(sync.WaitGroup)
	wg.Add(len(r.providers))

	for _, p := range r.providers {
		go func(p provider.DataProvider, _wg *sync.WaitGroup) {
			// 启动交易所记录器
			newMarketRecorder(r.source, r.store, p.Exchange(), p).Run().Wait()
			_wg.Done()
		}(p, wg)
	}

	return wg
}

// marketRecorder 市场记录器
type marketRecorder struct {
	source   source.Source         // 数据源
	store    store.Store           // 存储
	exchange *quote.Exchange       // 交易所
	provider provider.DataProvider // 提供者
}

// newMarketRecorder 新建市场记录器
func newMarketRecorder(source source.Source, store store.Store, exchange *quote.Exchange, provider provider.DataProvider) *marketRecorder {
	return &marketRecorder{source, store, exchange, provider}
}

// RunAndWait 启动市场记录器
func (r marketRecorder) Run() *sync.WaitGroup {
	wg := new(sync.WaitGroup)
	wg.Add(2)

	// 获取市场所在地到明天零点的时间差
	now := r.marketNow()
	duration := r.durationToNextDay(now)

	// 抓取历史数据
	go func(todayZero time.Time, _wg *sync.WaitGroup) {
		log.Printf("[%s] 获取历史数据开始", r.exchange.Name)
		err := r.crawlHistoryData(todayZero)
		if err != nil {
			log.Printf("[%s] 获取历史数据时发生错误: %v", r.exchange.Name, err)
			return
		}
		log.Printf("[%s] 获取历史数据结束", r.exchange.Name)
		_wg.Done()
	}(now.Add(duration).AddDate(0, 0, -1), wg)

	// 持续抓取每日数据
	go func() {
		log.Printf("[%s] 定时任务已启动，将于%s后激活下一次任务", r.exchange.Name, duration.String())
		for {
			<-time.After(duration)
			yesterday := r.marketNow().AddDate(0, 0, -1)
			log.Printf("[%s] 获取%s的数据开始", r.exchange.Name, yesterday.Format(datePattern))
			err := r.crawlYesterdayData(yesterday)
			if err != nil {
				log.Printf("[%s] 获取%s的数据时发生错误: %v", r.exchange.Name, yesterday.Format(datePattern), err)
			} else {
				log.Printf("[%s] 获取%s的数据结束", r.exchange.Name, yesterday.Format(datePattern))
			}

			// 每天调整延时时长，保证长时间运行的时间精度
			duration = r.durationToNextDay(r.marketNow())
		}

		// wg.Done()
	}()

	return wg
}

// marketNow 市场所处时区当前时间
func (r marketRecorder) marketNow() time.Time {
	return time.Now().In(r.exchange.Location)
}

// durationToNextDay 现在到明天0点的时间间隔
func (r marketRecorder) durationToNextDay(now time.Time) time.Duration {

	year, month, day := now.AddDate(0, 0, 1).Date()

	// 市场所处时区的下一个0点
	marketTomorrowZero := time.Date(year, month, day, 0, 0, 0, 0, now.Location())

	return marketTomorrowZero.Sub(now)
}

// crawlHistoryData 抓取历史数据
func (r marketRecorder) crawlHistoryData(todayZero time.Time) error {

	// 起始日期(含)
	date := todayZero.Add(-r.source.Expiration())
	log.Printf("[%s]抓取历史数据起始日期: %s  结束日期: %s", r.exchange.Name, date.Format(datePattern), todayZero.Format(datePattern))

	// 获取上市公司
	companies, err := r.provider.Companies()
	if err != nil {
		return err
	}
	log.Printf("[%s] 共有%d家上市公司", r.exchange.Name, len(companies))

	for date.Before(todayZero) {

		// 避免重复记录
		exists, err := r.store.Exists(r.exchange, date)
		if err != nil {
			return err
		}

		if !exists {
			// 抓取那一天的报价
			err = r.crawl(companies, date)
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
func (r marketRecorder) crawlYesterdayData(yesterday time.Time) error {

	// 避免重复记录
	recorded, err := r.store.Exists(r.exchange, yesterday)
	if err != nil || recorded {
		return err
	}

	// 获取上市公司
	companies, err := r.provider.Companies()
	if err != nil {
		return err
	}
	log.Printf("[%s] 共有%d家上市公司", r.exchange.Name, len(companies))

	// 抓取昨日数据
	return r.crawl(companies, yesterday)
}

// crawl 抓取指定日期的市场报价
func (r marketRecorder) crawl(companies []*quote.Company, date time.Time) error {

	ch := make(chan bool, r.source.ParallelMax())
	defer close(ch)

	var wg sync.WaitGroup
	wg.Add(len(companies))

	mutex := new(sync.Mutex)
	cdqs := make(map[string]*quote.CompanyDailyQuote, len(companies))
	for _, company := range companies {

		go func(exchange *quote.Exchange, _company *quote.Company, _date time.Time, _mutex *sync.Mutex, quotes map[string]*quote.CompanyDailyQuote) {
			cdq, err := r.source.Crawl(exchange, _company, _date)
			if err == nil {
				mutex.Lock()
				quotes[cdq.Code] = cdq
				mutex.Unlock()
			}

			<-ch
			wg.Done()
		}(r.exchange, company, date, mutex, cdqs)

		// 限流
		ch <- false
	}
	//	阻塞，直到抓取所有
	wg.Wait()

	edq := &quote.ExchangeDailyQuote{
		Exchange:  r.exchange,
		Date:      date,
		Companies: cdqs,
	}

	// 保存
	err := r.store.Save(edq)
	if err != nil {
		return fmt.Errorf("[%s] 保存上市公司在%s的分时数据时发生错误: %v", r.exchange.Name, date.Format(datePattern), err)
	}

	log.Printf("[%s] 上市公司在%s的分时数据已经抓取结束", r.exchange.Name, date.Format(datePattern))

	return nil
}
