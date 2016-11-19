package market

import (
	"fmt"
	"regexp"
	"sort"

	"github.com/guotie/gogb2312"
	"github.com/nzai/go-utility/net"
)

// China 中国证券市场
type China struct{}

// Name 名称
func (m China) Name() string {
	return "China"
}

// Timezone 时区
func (m China) Timezone() string {
	return "Asia/Shanghai"
}

// Companies 上市公司
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

	var companies []Company
	for _, company := range dict {
		companies = append(companies, company)
	}

	//	按Code排序
	sort.Sort(CompanyList(companies))

	return companies, nil
}

// shanghaiCompanies 上海证券交易所上市公司
func (m China) shanghaiCompanies() ([]Company, error) {

	urls := [...]string{
		"http://query.sse.com.cn/security/stock/downloadStockListFile.do?csrcCode=&stockCode=&areaName=&stockType=1",
		"http://query.sse.com.cn/security/stock/downloadStockListFile.do?csrcCode=&stockCode=&areaName=&stockType=2",
	}
	referer := "http://www.sse.com.cn/assortment/stock/list/share/"

	var list []Company
	for _, url := range urls {

		//	尝试从网络获取实时上市公司列表
		text, err := net.DownloadStringRefererRetry(url, referer, retryTimes, retryIntervalSeconds)
		if err != nil {
			return nil, err
		}

		//	解析json
		companies, err := m.parseShanghaiJSON(text)
		if err != nil {
			return nil, err
		}

		list = append(list, companies...)
	}

	return list, nil
}

// parseShanghaiJSON 解析上海证券交易所上市公司
func (m China) parseShanghaiJSON(text string) ([]Company, error) {

	//	深圳证券交易所的查询结果是GBK编码的，需要转成UTF8
	text, err, _, _ := gogb2312.ConvertGB2312String(text)
	if err != nil {
		return nil, err
	}

	//  使用正则分析json
	regex := regexp.MustCompile(`(\d{6})	  (\S+)	  \d{6}	  \S+`)
	group := regex.FindAllStringSubmatch(text, -1)

	var companies []Company
	for _, section := range group {
		companies = append(companies, Company{Code: section[1], Name: section[2]})
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

	var list []Company
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
		companies, err := m.parseShenzhenHTML(html)
		if err != nil {
			return nil, err
		}

		list = append(list, companies...)
	}

	return list, nil
}

// parseShenzhenHTML 解析深圳证券交易所上市公司
func (m China) parseShenzhenHTML(html string) ([]Company, error) {
	//  使用正则分析html
	regex := regexp.MustCompile(`@' align='center'  >(\d{6})</td><td  class='cls-data-td' null align='center'  >([^<]*?)</td>`)
	group := regex.FindAllStringSubmatch(html, -1)

	var companies []Company
	for _, section := range group {
		companies = append(companies, Company{Code: section[1], Name: section[2]})
	}

	if len(companies) == 0 {
		return nil, fmt.Errorf("错误的深圳证券交易所上市公司列表内容:%s", html)
	}

	return companies, nil
}

// YahooQueryCode 雅虎查询代码
func (m China) YahooQueryCode(company Company) string {

	var suffix string
	switch company.Code[:1] {
	case "0":
		suffix = "SZ"
	case "2":
		suffix = "SZ"
	case "3":
		suffix = "SZ"
	case "9":
		suffix = "SS"
	case "6":
		suffix = "SS"
	default:
		suffix = "SS"
	}

	return company.Code + "." + suffix
}
