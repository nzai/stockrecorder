package market

import (
	"fmt"
	"regexp"
	"sort"

	"github.com/nzai/go-utility/net"
)

// HongKong 香港证券市场
type HongKong struct{}

// Name 名称
func (m HongKong) Name() string {
	return "HongKong"
}

// Timezone 所处时区
func (m HongKong) Timezone() string {
	return "Asia/Hong_Kong"
}

// Companies 上市公司
func (m HongKong) Companies() ([]Company, error) {

	urls := [...]string{
		"https://www.hkex.com.hk/chi/market/sec_tradinfo/stockcode/eisdeqty_c.htm",
		"https://www.hkex.com.hk/chi/market/sec_tradinfo/stockcode/eisdgems_c.htm",
	}

	var list []Company
	for _, url := range urls {

		//	尝试从网络获取实时上市公司列表
		html, err := net.DownloadStringRetry(url, retryTimes, retryIntervalSeconds)
		if err != nil {
			return nil, err
		}

		//	解析json
		companies, err := m.parseHTML(html)
		if err != nil {
			return nil, err
		}

		list = append(list, companies...)
	}

	//	按Code排序
	sort.Sort(CompanyList(list))

	return list, nil
}

//	解析香港证券交易所上市公司
func (m HongKong) parseHTML(html string) ([]Company, error) {

	//  使用正则分析json
	regex := regexp.MustCompile(`>(\d{5})</td>\s*?<td[^>]*?>(<a.*?>)?([^<]+?)(</a>)?</td>`)
	group := regex.FindAllStringSubmatch(html, -1)

	var companies []Company
	for _, section := range group {
		companies = append(companies, Company{Code: section[1], Name: section[3]})
	}

	if len(companies) == 0 {
		return nil, fmt.Errorf("错误的香港证券交易所上市公司列表内容:%s", html)
	}

	return companies, nil
}

// YahooQueryCode 雅虎查询代码
func (m HongKong) YahooQueryCode(company Company) string {

	queryCode := company.Code[1:] + ".HK"
	if company.Code[:1] != "0" {
		queryCode = company.Code + ".HK"
	}

	return queryCode
}
