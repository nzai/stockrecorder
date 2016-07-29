package market

import (
	"fmt"
	"regexp"
	"sort"
	"time"

	"github.com/guotie/gogb2312"
	"github.com/nzai/go-utility/net"
)

//	中国证券市场
type China struct{}

func (m China) Name() string {
	return "China"
}

func (m China) Timezone() string {
	return "Asia/Shanghai"
}

//	更新上市公司列表
func (m China) Companies() ([]Company, error) {

	dict := make(map[string]Company, 0)

	//	上海证券交易所
	sh, err := m.shanghaiCompanies()
	if err != nil {
		return nil, err
	}
	for _, company := range sh {
		//	去重
		if _, found := dict[company.Code]; found {
			continue
		}

		dict[company.Code] = company
	}

	//	深圳证券交易所
	sz, err := m.shenzhenCompanies()
	if err != nil {
		return nil, err
	}
	for _, company := range sz {
		//	去重
		if _, found := dict[company.Code]; found {
			continue
		}

		dict[company.Code] = company
	}

	companies := make([]Company, 0)
	for _, company := range dict {
		companies = append(companies, company)
	}

	//	按Code排序
	sort.Sort(CompanyList(companies))

	return companies, nil
}

//	上海证券交易所上市公司
func (m China) shanghaiCompanies() ([]Company, error) {

	urls := [...]string{
		"http://query.sse.com.cn/security/stock/downloadStockListFile.do?csrcCode=&stockCode=&areaName=&stockType=1",
		"http://query.sse.com.cn/security/stock/downloadStockListFile.do?csrcCode=&stockCode=&areaName=&stockType=2",
	}
	referer := "http://www.sse.com.cn/assortment/stock/list/share/"

	list := make([]Company, 0)
	for _, url := range urls {

		//	尝试从网络获取实时上市公司列表
		text, err := net.DownloadStringRefererRetry(url, referer, retryTimes, retryIntervalSeconds)
		if err != nil {
			return nil, err
		}

		//	解析json
		companies, err := m.parseShanghaiJson(text)
		if err != nil {
			return nil, err
		}

		list = append(list, companies...)
	}

	return list, nil
}

//	解析上海证券交易所上市公司
func (m China) parseShanghaiJson(text string) ([]Company, error) {

	//	深圳证券交易所的查询结果是GBK编码的，需要转成UTF8
	text, err, _, _ := gogb2312.ConvertGB2312String(text)
	if err != nil {
		return nil, err
	}

	//  使用正则分析json
	regex := regexp.MustCompile(`(\d{6})	  (\S+)	  \d{6}	  \S+`)
	group := regex.FindAllStringSubmatch(text, -1)

	companies := make([]Company, 0)
	for _, section := range group {
		companies = append(companies, Company{Market: m.Name(), Code: section[1], Name: section[2]})
	}

	if len(companies) == 0 {
		return nil, fmt.Errorf("错误的上海证券交易所上市公司列表内容:%s", text)
	}

	return companies, nil
}

//	深圳证券交易所上市公司
func (m China) shenzhenCompanies() ([]Company, error) {
	urls := [...]string{
		"http://www.szse.cn/szseWeb/ShowReport.szse?SHOWTYPE=EXCEL&CATALOGID=1110&tab1PAGENUM=1&ENCODE=1&TABKEY=tab1",
	}

	list := make([]Company, 0)
	for _, url := range urls {

		//	尝试从网络获取实时上市公司列表
		html, err := net.DownloadStringRetry(url, retryTimes, retryIntervalSeconds)
		if err != nil {
			return nil, err
		}

		//	深圳证券交易所的查询结果是GBK编码的，需要转成UTF8
		html, err, _, _ = gogb2312.ConvertGB2312String(html)
		if err != nil {
			return nil, err
		}

		//	解析Html
		companies, err := m.parseShenzhenHtml(html)
		if err != nil {
			return nil, err
		}

		list = append(list, companies...)
	}

	return list, nil
}

//	解析深圳证券交易所上市公司
func (m China) parseShenzhenHtml(html string) ([]Company, error) {
	//  使用正则分析html
	regex := regexp.MustCompile(`null align='center' >(\d{6})</td><td  class='cls-data-td' null align='center' >([^<]*?)</td>`)
	group := regex.FindAllStringSubmatch(html, -1)

	companies := make([]Company, 0)
	for _, section := range group {
		companies = append(companies, Company{Market: m.Name(), Code: section[1], Name: section[2]})
	}

	if len(companies) == 0 {
		return nil, fmt.Errorf("错误的深圳证券交易所上市公司列表内容:%s", html)
	}

	return companies, nil
}

//	所有中国上市公司,在雅虎财经查询分时数据时都要带上后缀
var chineseSuffix map[string]string = map[string]string{
	"6": "SS",
	"9": "SS",
	"0": "SZ",
	"2": "SZ",
	"3": "SZ",
}

//	抓取
func (m China) Crawl(code string, day time.Time) (string, error) {

	suffix, found := chineseSuffix[code[:1]]
	if !found {
		suffix = "SS"
	}

	return downloadCompanyDaily(m, code, code+"."+suffix, day)
}
