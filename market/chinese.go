package market

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	"github.com/guotie/gogb2312"
	"github.com/nzai/stockrecorder/config"
	"github.com/nzai/stockrecorder/io"
)

//	美股市场
type Chinese struct{}

func (m Chinese) Name() string {
	return "Chinese"
}

func (m Chinese) Timezone() string {
	return "Asia/Shanghai"
}

//	所有中国上市公司,在雅虎财经查询分时数据时都要带上后缀
var chineseSuffix map[string]string

func init() {
	chineseSuffix = map[string]string{
		"6": "SS",
		"9": "SS",
		"0": "SZ",
		"2": "SZ",
		"3": "SZ",
	}
}

//	更新上市公司列表
func (m Chinese) LastestCompanies() ([]Company, error) {

	companies := make([]Company, 0)

	//	上海证券交易所
	sh, err := m.shanghaiCompanies()
	if err != nil {
		return nil, err
	}
	companies = append(companies, sh...)

	//	深圳证券交易所
	sz, err := m.shenzhenCompanies()
	if err != nil {
		return nil, err
	}
	companies = append(companies, sz...)

	//	按Code排序
	sort.Sort(CompanyList(companies))

	return companies, nil
}

//	上海证券交易所上市公司
func (m Chinese) shanghaiCompanies() ([]Company, error) {

	urls := [...]string{
		"http://query.sse.com.cn/commonQuery.do?isPagination=false&sqlId=COMMON_SSE_ZQPZ_GPLB_MCJS_SSAG_L",
		"http://query.sse.com.cn/commonQuery.do?isPagination=false&sqlId=COMMON_SSE_ZQPZ_GPLB_MCJS_SSBG_L",
	}
	referer := "http://www.sse.com.cn/assortment/stock/list/name/"

	var err error
	list := make([]Company, 0)
	for _, url := range urls {

		//	尝试从网络获取实时上市公司列表
		for try := retryCount; try > 0; try-- {
			json, err := io.GetStringReferer(url, referer)
			if err == nil {
				companies, err := m.parseShanghaiJson(json)
				if err == nil {
					list = append(list, companies...)
					break
				}
			}

			if try == 0 {
				break
			}

			log.Fatalf("获取上海证券交易所上市公司出错[%d](%d秒钟后重试):%v", try-1, retryDelaySeconds, err)
			time.Sleep(time.Duration(retryDelaySeconds) * time.Second)
		}
	}

	if err != nil {
		return nil, err
	}

	return list, nil
}

//	解析上海证券交易所上市公司
func (m Chinese) parseShanghaiJson(json string) ([]Company, error) {

	//  使用正则分析json
	regex := regexp.MustCompile(`"PRODUCTNAME":"([^"]*?)","PRODUCTID":"(\d{6})"`)
	group := regex.FindAllStringSubmatch(json, -1)

	companies := make([]Company, 0)
	for _, section := range group {
		//log.Printf("%s\t%s\n", section[2], section[1])
		companies = append(companies, Company{Code: section[2], Name: section[1]})
	}

	if len(companies) == 0 {
		return nil, errors.New(fmt.Sprintf("错误的上海证券交易所上市公司列表内容:%s", json))
	}

	return companies, nil
}

//	深圳证券交易所上市公司
func (m Chinese) shenzhenCompanies() ([]Company, error) {
	urls := [...]string{
		"http://www.szse.cn/szseWeb/FrontController.szse?ACTIONID=8&CATALOGID=1110&tab1PAGENUM=1&ENCODE=1&TABKEY=tab1",
	}

	var err error
	list := make([]Company, 0)
	for _, url := range urls {

		//	尝试从网络获取实时上市公司列表
		for try := retryCount; try > 0; try-- {
			html, err := io.GetString(url)
			if err == nil {
				//	深圳证券交易所的查询结果是GBK编码的，需要转成UTF8
				html, err, _, _ = gogb2312.ConvertGB2312String(html)
				companies, err := m.parseShenzhenHtml(html)
				if err == nil {
					list = append(list, companies...)
					break
				}
			}

			if try == 0 {
				break
			}

			log.Fatalf("获取深圳证券交易所上市公司出错[%d](%d秒钟后重试):%v", try-1, retryDelaySeconds, err)
			time.Sleep(time.Duration(retryDelaySeconds) * time.Second)
		}
	}

	if err != nil {
		return nil, err
	}

	return list, nil
}

//	解析深圳证券交易所上市公司
func (m Chinese) parseShenzhenHtml(html string) ([]Company, error) {
	//  使用正则分析html
	regex := regexp.MustCompile(`align='center' >(\d{6})</td><td  class='cls-data-td'  align='center' >([^<]*?)</td>`)
	group := regex.FindAllStringSubmatch(html, -1)

	companies := make([]Company, 0)
	for _, section := range group {
		companies = append(companies, Company{Code: section[1], Name: section[2]})
	}

	if len(companies) == 0 {
		return nil, errors.New(fmt.Sprintf("错误的深圳证券交易所上市公司列表内容:%s", html))
	}

	return companies, nil
}

//	抓取
func (m Chinese) Crawl(market Market, company Company, day time.Time) error {
	//	文件保存路径
	dataDir := config.GetDataDir()
	fileName := fmt.Sprintf("%s_raw.txt", day.Format("20060102"))
	filePath := filepath.Join(dataDir, market.Name(), company.Code, fileName)

	//	如果文件已存在就忽略
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		//	如果不存在就抓取并保存
		start := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())
		end := start.Add(time.Hour * 24)

		suffix, found := chineseSuffix[company.Code[:1]]
		if !found {
			suffix = "SS"
		}

		pattern := "https://finance-yql.media.yahoo.com/v7/finance/chart/%s.%s?period2=%d&period1=%d&interval=1m&indicators=quote&includeTimestamps=true&includePrePost=true&events=div%7Csplit%7Cearn&corsDomain=finance.yahoo.com"
		url := fmt.Sprintf(pattern, company.Code, suffix, end.Unix(), start.Unix())

		html, err := io.GetString(url)
		if err != nil {
			return err
		}

		//	写入文件
		return io.WriteString(filePath, html)
	}

	return nil
}
