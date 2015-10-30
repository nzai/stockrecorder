package market

import (
	"fmt"
	"regexp"
	"sort"
	"time"

	"github.com/nzai/stockrecorder/io"
)

//	香港证券市场
type HongKong struct{}

func (m HongKong) Name() string {
	return "HongKong"
}

func (m HongKong) Timezone() string {
	return "Asia/Hong_Kong"
}

//	更新上市公司列表
func (m HongKong) Companies() ([]Company, error) {

	urls := [...]string{
		"https://www.hkex.com.hk/chi/market/sec_tradinfo/stockcode/eisdeqty_c.htm",
		"https://www.hkex.com.hk/chi/market/sec_tradinfo/stockcode/eisdgems_c.htm",
	}

	companies := make([]Company, 0)
	for _, url := range urls {

		//	尝试从网络获取实时上市公司列表
		html, err := io.DownloadStringRetry(url, retryTimes, retryIntervalSeconds)
		if err != nil {
			return nil, err
		}

		//	解析json
		hk, err := m.parseHtml(html)
		if err != nil {
			return nil, err
		}

		companies = append(companies, hk...)
	}

	//	按Code排序
	sort.Sort(CompanyList(companies))

	return companies, nil
}

//	解析香港证券交易所上市公司
func (m HongKong) parseHtml(html string) ([]Company, error) {

	//  使用正则分析json
	regex := regexp.MustCompile(`>(\d{5})</td>\s*?<td[^>]*?>(<a.*?>)?([^<]+?)(</a>)?</td>`)
	group := regex.FindAllStringSubmatch(html, -1)

	companies := make([]Company, 0)
	for _, section := range group {
		companies = append(companies, Company{Market: m.Name(), Code: section[1], Name: section[3]})
	}

	if len(companies) == 0 {
		return nil, fmt.Errorf("错误的香港证券交易所上市公司列表内容:%s", html)
	}

	return companies, nil
}

//	抓取
func (m HongKong) Crawl(market Market, company Company, day time.Time) error {
	queryCode := company.Code[1:] + ".HK"
	if company.Code[:1] != "0" {
		queryCode = company.Code + ".HK"
	}

	return downloadCompanyDaily(market, company.Code, queryCode, day)
}
