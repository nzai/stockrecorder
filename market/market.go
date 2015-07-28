package market

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/nzai/stockrecorder/config"
	"github.com/nzai/stockrecorder/io"
)

const (
	//	雅虎财经的历史分时数据没有超过60天的
	lastestDays       = 60
	companyGCCount    = 32
	retryCount        = 5
	retryDelaySeconds = 600
	companiesFileName = "companies.txt"
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
	Crawl(company Company, day time.Time) error
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
		monitorMarket(market)
	}
}

//	监视市场
func monitorMarket(market Market) {

	//	获取市场所在时区的今天0点
	today := marketToday(market)
	now := time.Now().In(today.Location())

	//	定时器今天
	du := today.Add(time.Hour * 24).Sub(now)
	log.Printf("[%s]\t%s后启动首次定时任务", market.Name(), du.String())
	time.AfterFunc(du, func() {
		ticker := time.NewTicker(time.Hour * 24)
		log.Printf("[%s]\t已启动定时抓取任务", market.Name())

		//	立刻运行一次
		go func() {
			err := crawlMarket(market)
			if err != nil {
				log.Fatalf("[%s]\t抓取%s的数据出错:%v", market.Name(), today.Format("20060102"), err)
			}
		}()

		//	每天运行一次
		for _ = range ticker.C {
			err := crawlMarket(market)
			if err != nil {
				log.Fatalf("[%s]\t抓取%s的数据出错:%v", market.Name(), today.Format("20060102"), err)
			}
		}
	})

	marketHistory(market)
}

//	市场所处时区今日0点
func marketToday(market Market) time.Time {
	//	获取市场所在时区的今天0点
	location, err := time.LoadLocation(market.Timezone())
	if err != nil {
		return time.Now()
	}

	now := time.Now().In(location)

	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
}

//	抓取市场上市公司信息
func crawlMarketCompanies(market Market) ([]Company, error) {

	//	尝试更新上市公司列表
	log.Printf("[%s]\t更新上市公司列表-开始", market.Name())
	companies, err := market.LastestCompanies()
	if err != nil {

		//	如果更新失败，则尝试从上次的存档文件中读取上市公司列表
		log.Printf("[%s]\t更新上市公司列表失败，尝试从存档读取", market.Name())
		companies, err = loadMarketCompanies(market)
		if err != nil {
			log.Printf("[%s]\t尝试从存档读取上市公司列表-失败", market.Name())
			return nil, err
		}

		log.Printf("[%s]\t尝试从存档读取上市公司列表-成功,共%d家上市公司", market.Name(), len(companies))

		return companies, nil
	}

	//	存档
	err = saveMarketCompanies(market, companies)
	if err != nil {
		return nil, err
	}

	log.Printf("[%s]\t更新上市公司列表-成功,共%d家上市公司", market.Name(), len(companies))

	return companies, nil
}

//	上市公司列表保存路径
func marketCompaniesPath(market Market) string {
	return filepath.Join(config.GetDataDir(), market.Name(), companiesFileName)
}

//	保存上市公司列表到文件
func saveMarketCompanies(market Market, companies []Company) error {

	lines := make([]string, 0)
	for _, company := range companies {
		lines = append(lines, fmt.Sprintf("%s\t%s\n", company.Code, company.Name))
	}

	return io.WriteLines(marketCompaniesPath(market), lines)
}

//	从存档读取上市公司列表
func loadMarketCompanies(market Market) ([]Company, error) {
	lines, err := io.ReadLines(marketCompaniesPath(market))
	if err != nil {
		return nil, err
	}

	companies := make([]Company, 0)
	for _, line := range lines {
		parts := strings.Split(line, "\t")
		if len(parts) != 2 {
			return nil, errors.New("错误的上市公司列表格式")
		}

		companies = append(companies, Company{Code: parts[0], Name: parts[1]})
	}

	return companies, nil
}

//	抓取市场前一天的数据
func crawlMarket(market Market) error {

	//	获取市场内的所有上市公司
	companies, err := crawlMarketCompanies(market)
	if err != nil {
		return err
	}

	//	前一天
	yesterday := marketToday(market).Add(-time.Hour * 24)

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
		err := market.Crawl(company, day)
		if err == nil {
			break
		}

		if try > 0 {
			log.Fatalf("[%s]\t抓取%s的数据出错(还有%d次):%v", market.Name(), company.Name, try-1, err)
			time.Sleep(retryDelaySeconds * time.Second)
		} else {
			log.Printf("[%s]\t抓取%s在%s的数据失败,已重试%d次", market.Name(), company.Name, day.Format("20060102"), retryCount)
		}
	}
}

//	市场历史
func marketHistory(market Market) error {
	//	获取市场内的所有上市公司
	companies, err := crawlMarketCompanies(market)
	if err != nil {
		return err
	}

	//	前一天
	yesterday := marketToday(market).Add(-time.Hour * 24)

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
	//	保存原始数据
	for index := 0; index < lastestDays; index++ {
		day = day.Add(-time.Hour * 24)
		crawlCompany(market, company, day)
	}
}
