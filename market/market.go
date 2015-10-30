package market

import (
	"fmt"
	"log"
	"time"
)

const (
	//	雅虎财经的历史分时数据没有超过90天的
	lastestDays          = 90
	companyGCCount       = 64
	retryTimes           = 50
	retryIntervalSeconds = 10
)

//	市场更新
type Market interface {
	//	名称
	Name() string
	//	时区
	Timezone() string
	//	获取上市公司列表
	Companies() ([]Company, error)

	//	抓取任务(每日)
	Crawl(market Market, company Company, day time.Time) error
}

var (
	markets                       = []Market{}
	marketOffset map[string]int64 = make(map[string]int64)
)

//	添加市场
func Add(market Market) {

	markets = append(markets, market)

	log.Printf("市场[%s]已经加入监视列表", market.Name())
}

//	监视市场(所有操作的入口)
func Monitor() error {
	log.Print("启动监视")

	for _, m := range markets {
		//	本地时间
		now := time.Now()
		_, offsetLocal := now.Zone()

		//	获取市场所在时区
		location, err := time.LoadLocation(m.Timezone())
		if err != nil {
			return err
		}

		//	市场所处时区当前时间
		marketNow := now.In(location)
		_, offsetMarket := marketNow.Zone()

		//	计算TimeZoneOffset
		marketOffset[m.Name()] = int64(offsetMarket - offsetLocal)
	}

	//	启动处理队列
	go startProcessQueue()

	//	启动抓取任务
	for _, m := range markets {

		//	启动每日定时任务
		go func(market Market) {
			//	所处时区距明日0点的间隔
			now := marketow(market)
			du := locationYesterdayZero(market).Add(time.Hour * 48).Sub(now)

			log.Printf("[%s]\t定时任务已启动，将于%s后激活首次任务", market.Name(), du.String())
			time.AfterFunc(du, func() {
				//	立即运行一次
				go dailyTask(market)

				//	每天运行一次
				ticker := time.NewTicker(time.Hour * 24)
				for _ = range ticker.C {
					dailyTask(market)
				}
			})

		}(m)

		//	启动历史数据获取任务
		go func(market Market) {
			historyTask(market, locationYesterdayZero(market))
		}(m)
	}

	return nil
}

//	市场所处时区当前时间
func marketow(market Market) time.Time {
	now := time.Now()

	//	获取市场所在时区
	location, err := time.LoadLocation(market.Timezone())
	if err != nil {
		return now
	}

	return now.In(location)
}

//	昨天0点
func locationYesterdayZero(market Market) time.Time {
	now := marketow(market)
	year, month, day := now.Add(-time.Hour * 24).Date()

	return time.Date(year, month, day, 0, 0, 0, 0, now.Location())
}

//	每日定时任务
func dailyTask(market Market) {

	//	昨天零点
	yesterday := locationYesterdayZero(market)
	log.Printf("[%s]\t%s数据获取任务已启动", market.Name(), yesterday.Format("20060102"))

	//	获取市场所有上市公司
	companies, err := getCompanies(market)
	if err != nil {
		log.Printf("[%s]\t获取上市公司失败: %s", market.Name(), err.Error())
		return
	}

	chanSend := make(chan int, companyGCCount)
	chanReceive := make(chan int)
	defer close(chanSend)
	defer close(chanReceive)

	for _, c := range companies {
		//	并发抓取
		go func(company Company) {

			err := market.Crawl(market, company, yesterday)
			if err != nil {
				log.Printf("[%s]\t抓取上市公司[%s]数据失败: %s", market.Name(), company.Name, err.Error())
			}

			<-chanSend
			chanReceive <- 1
		}(c)

		chanSend <- 1
	}

	//	阻塞，直到抓取所有
	for _, _ = range companies {
		<-chanReceive
	}

	log.Printf("[%s]\t%s数据获取任务已结束", market.Name(), yesterday.Format("20060102"))
}

//	历史数据获取任务
func historyTask(market Market, yesterday time.Time) {
	//	获取市场所有上市公司
	companies, err := getCompanies(market)
	if err != nil {
		log.Printf("[%s]\t获取上市公司失败: %s", market.Name(), err.Error())
		return
	}

	log.Printf("[%s]\t开始抓取%d家上市公司在%s之前的历史", market.Name(), len(companies), yesterday.Format("20060102"))

	chanSend := make(chan int, companyGCCount)
	chanReceive := make(chan int)
	defer close(chanSend)
	defer close(chanReceive)

	for _, c := range companies {
		//	并发抓取
		go func(company Company) {

			for index := 0; index < lastestDays; index++ {
				day := yesterday.Add(-time.Hour * 24 * time.Duration(index))
				err := market.Crawl(market, company, day)
				if err != nil {
					log.Printf("[%s]\t抓取[%s]在%s的分时数据出错:%s", market.Name(), company.Code, day.Format("20060102"), err)
				}
			}

			<-chanSend
			chanReceive <- 1
		}(c)

		chanSend <- 1
	}

	//	阻塞，直到抓取所有
	for _, _ = range companies {
		<-chanReceive
	}

	log.Printf("[%s]\t上市公司的历史分时数据已经抓取结束", market.Name())
}

//	抓取市场上市公司信息
func getCompanies(market Market) ([]Company, error) {

	cl := CompanyList{}
	//	尝试更新上市公司列表
	log.Printf("[%s]\t更新上市公司列表-开始", market.Name())
	companies, err := market.Companies()
	if err != nil {

		//	如果更新失败，则尝试从上次的存档文件中读取上市公司列表
		log.Printf("[%s]\t更新上市公司列表失败，尝试从存档读取:%v", market.Name(), err)
		err = cl.Load(market)
		if err != nil {
			return nil, fmt.Errorf("[%s]\t尝试从存档读取上市公司列表-失败:%v", market.Name(), err)
		}

		companies = cl
		log.Printf("[%s]\t尝试从存档读取上市公司列表-成功,共%d家上市公司", market.Name(), len(companies))

		return companies, nil
	}

	//	存档
	cl = CompanyList(companies)
	err = cl.Save(market)
	if err != nil {
		return nil, err
	}

	log.Printf("[%s]\t更新上市公司列表-成功,共%d家上市公司", market.Name(), len(companies))

	return companies, nil
}
