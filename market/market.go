package market

import (
	"log"
	"time"
)

const (
	//	雅虎财经的历史分时数据没有超过90天的
	lastestDays       = 90
	companyGCCount    = 32
	retryCount        = 100
	retryDelaySeconds = 30
)

//	市场更新
type Market interface {
	//	名称
	Name() string
	//	时区
	Timezone() string
	//	获取上市公司列表
	LastestCompanies() ([]Company, error)

	//	抓取任务(每日)
	Crawl(market Market, company Company, day time.Time) error
}

var markets = []Market{}

//	添加市场
func Add(market Market) {
	markets = append(markets, market)

	log.Printf("市场[%s]已经加入监视列表", market.Name())
}

//	监视市场(所有操作的入口)
func Monitor() {
	log.Print("启动监视")

	for _, market := range markets {
		go monitorMarket(market)
	}
}

//	监视市场
func monitorMarket(market Market) {

	//	获取市场所在时区的今天0点
	today := today(market)
	now := time.Now().In(today.Location())

	//	定时器今天
	du := today.Add(time.Hour * 24).Sub(now)
	log.Printf("[%s]\t%s后启动首次定时任务", market.Name(), du.String())
	time.AfterFunc(du, func() {
		ticker := time.NewTicker(time.Hour * 24)
		log.Printf("[%s]\t已启动定时抓取任务", market.Name())

		//	立刻运行一次
		go func() {
			err := crawlYesterday(market)
			if err != nil {
				log.Fatalf("[%s]\t抓取%s的数据出错:%v", market.Name(), today.Format("20060102"), err)
			}
		}()

		//	每天运行一次
		for _ = range ticker.C {
			err := crawlYesterday(market)
			if err != nil {
				log.Fatalf("[%s]\t抓取%s的数据出错:%v", market.Name(), today.Format("20060102"), err)
			}
		}
	})

	marketHistory(market)
}

//	市场所处时区今日0点
func today(market Market) time.Time {
	//	获取市场所在时区的今天0点
	location, err := time.LoadLocation(market.Timezone())
	if err != nil {
		return time.Now()
	}

	now := time.Now().In(location)

	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
}

//	抓取市场上市公司信息
func companies(market Market) ([]Company, error) {

	cl := CompanyList{}
	//	尝试更新上市公司列表
	log.Printf("[%s]\t更新上市公司列表-开始", market.Name())
	companies, err := market.LastestCompanies()
	if err != nil {

		//	如果更新失败，则尝试从上次的存档文件中读取上市公司列表
		log.Printf("[%s]\t更新上市公司列表失败，尝试从存档读取", market.Name())
		err = cl.Load(market)
		if err != nil {
			log.Printf("[%s]\t尝试从存档读取上市公司列表-失败", market.Name())
			return nil, err
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

//	抓取市场前一天的数据
func crawlYesterday(market Market) error {

	//	获取市场内的所有上市公司
	companies, err := companies(market)
	if err != nil {
		return err
	}

	//	前一天
	yesterday := today(market).Add(-time.Hour * 24)

	chanSend := make(chan int, companyGCCount)
	chanReceive := make(chan int)

	for _, c := range companies {
		//	并发抓取
		go func(company Company) {

			crawlCompany(market, company, yesterday)

			<-chanSend
			chanReceive <- 1
		}(c)

		chanSend <- 1
	}

	//	阻塞，直到抓取所有
	for _, _ = range companies {
		<-chanReceive
	}

	close(chanSend)
	close(chanReceive)

	return nil
}

//	抓取上市公司某日数据
func crawlCompany(market Market, company Company, day time.Time) {

	for try := retryCount; try > 0; try-- {
		//	抓取数据
		err := market.Crawl(market, company, day)
		if err == nil {
			break
		}

		if try > 0 {
			log.Fatalf("[%s]\t抓取%s的数据出错[%d](%d秒后重试):%v", market.Name(), company.Name, try-1, retryDelaySeconds, err)
			t := time.NewTimer(time.Second * retryDelaySeconds)
			<-t.C
			log.Fatalf("[%s]\t抓取%s的数据", market.Name(), company.Name)
		} else {
			log.Printf("[%s]\t抓取%s在%s的数据失败,已重试%d次", market.Name(), company.Name, day.Format("20060102"), retryCount)
		}
	}
}

//	市场历史
func marketHistory(market Market) error {
	//	获取市场内的所有上市公司
	companies, err := companies(market)
	if err != nil {
		return err
	}

	log.Printf("[%s]\t开始抓取%d家上市公司的历史", market.Name(), len(companies))

	//	前一天
	yesterday := today(market)

	chanSend := make(chan int, companyGCCount)
	chanReceive := make(chan int)

	for _, c := range companies {
		//	并发抓取
		go func(company Company) {

			marketCompanyHistory(market, company, yesterday)

			<-chanSend
			chanReceive <- 1
		}(c)

		chanSend <- 1
	}

	//	阻塞，直到抓取所有
	for _, _ = range companies {
		<-chanReceive
	}

	close(chanSend)
	close(chanReceive)

	return nil
}

//	公司历史
func marketCompanyHistory(market Market, company Company, day time.Time) {
	//	查询之前一段时间的数据
	for index := 0; index < lastestDays; index++ {
		day = day.Add(-time.Hour * 24)
		crawlCompany(market, company, day)
	}
}
