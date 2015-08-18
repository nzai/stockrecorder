package market

import (
	"log"
	"time"
)

const (
	//	雅虎财经的历史分时数据没有超过90天的
	lastestDays          = 90
	companyGCCount       = 32
	retryTimes           = 50
	retryIntervalSeconds = 60
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

	for _, m := range markets {
		go func(mkt Market) {
			
			//	市场所处时区当前时间
			now := locationNow(mkt)
			
			//	启动每日定时任务
			go func(market Market){
				//	所处时区的明日0点
				tomorrowZero := now.Truncate(time.Hour * 24).Add(time.Hour * 24)
				du := tomorrowZero.Sub(now)
				
				log.Printf("[%s]\t定时任务已启动，将于%d时%d分%d秒后激活首次任务", market.Name(), du.Hours(), du.Minutes(), du.Seconds());
				time.AfterFunc(du, func(){
					
				})
				

			}(mkt)
			
			monitorMarket(mkt)
		}(m) 
	}
}

//	监视市场
func monitorMarket(market Market) {

	//	获取市场所在时区的今天0点
	now := locationNow(market)
	tomorrowZero := now.Truncate(time.Hour * 24).Add(time.Hour * 24)

	//	到明天的时间间隔
	du := tomorrowZero.Sub(now)
	log.Printf("[%s]\t%s后启动首次定时任务", market.Name(), du.String())
	time.AfterFunc(du, func() {
		ticker := time.NewTicker(time.Hour * 24)
		log.Printf("[%s]\t已启动定时抓取任务", market.Name())

		//	立刻运行一次
//		go func() {
//			err := crawlYesterday(market)
//			if err != nil {
//				log.Fatalf("[%s]\t抓取%s的数据出错:%v", market.Name(), today.Format("20060102"), err)
//			}
//		}()

		//	每天运行一次
		for _ = range ticker.C {
			err := crawlYesterday(market)
			if err != nil {
				log.Fatalf("[%s]\t抓取%s的数据出错:%v", market.Name(), now.Format("20060102"), err)
			}
		}
	})

	marketHistory(market)
}

//	所处时区当前时间
func locationNow(market Market) time.Time {
	now := time.Now()
	
	//	获取市场所在时区的今天0点
	location, err := time.LoadLocation(market.Timezone())
	if err != nil {
		return now
	}
	
	return now.In(location)
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
	yesterday := locationNow(market).Truncate(time.Hour * 24).Add(-time.Hour * 24)

	chanSend := make(chan int, companyGCCount)
	chanReceive := make(chan int)

	for _, c := range companies {
		//	并发抓取
		go func(company Company) {

			err := market.Crawl(market, company, yesterday)
			if err != nil {
				log.Fatal(err)
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

	close(chanSend)
	close(chanReceive)

	return nil
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
	yesterday := locationNow(market).Truncate(time.Hour * 24).Add(-time.Hour * 24)

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
		err := market.Crawl(market, company, day)
		if err != nil {
			log.Fatal(err)
		}
	}
}
