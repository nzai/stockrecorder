package market

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"time"

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

	source := map[string]string{
		"http://www.hkex.com.hk/Market-Data/Securities-Prices/Equities?sc_lang=zh-HK":                      "https://www1.hkex.com.hk/hkexwidget/data/getequityfilter?lang=chi&token=%s&sort=5&order=0&all=1&qid=%d&callback=3322", // 股本證券
		"http://www.hkex.com.hk/Market-Data/Securities-Prices/Exchange-Traded-Products?sc_lang=zh-hk":      "https://www1.hkex.com.hk/hkexwidget/data/getetpfilter?lang=chi&token=%s&sort=2&order=1&all=1&qid=%d&callback=3322",    // 交易所買賣產品
		"http://www.hkex.com.hk/Market-Data/Securities-Prices/Derivative-Warrants?sc_lang=zh-hk":           "https://www1.hkex.com.hk/hkexwidget/data/getdwfilter?lang=chi&token=%s&sort=5&order=0&all=1&qid=%d&callback=3322",     // 衍生權證
		"http://www.hkex.com.hk/Market-Data/Securities-Prices/Callable-Bull-Bear-Contracts?sc_lang=zh-hk":  "https://www1.hkex.com.hk/hkexwidget/data/getcbbcfilter?lang=chi&token=%s&sort=5&order=0&all=1&qid=%d&callback=3322",   // 牛熊證
		"http://www.hkex.com.hk/Market-Data/Securities-Prices/Real-Estate-Investment-Trusts?sc_lang=zh-hk": "https://www1.hkex.com.hk/hkexwidget/data/getreitfilter?lang=chi&token=%s&sort=5&order=0&all=1&qid=%d&callback=3322",   // 房地產投資信託基金
		"http://www.hkex.com.hk/Market-Data/Securities-Prices/Debt-Securities?sc_lang=zh-hk":               "https://www1.hkex.com.hk/hkexwidget/data/getdebtfilter?lang=chi&token=%s&sort=0&order=1&all=1&qid=%d&callback=3322",   // 債務證券
	}

	var companies []Company
	for page, api := range source {
		_companies, err := m.queryCompanies(page, api)
		if err != nil {
			return nil, err
		}

		companies = append(companies, _companies...)
	}

	//	按Code排序
	sort.Sort(CompanyList(companies))

	return companies, nil
}

//	解析香港证券交易所上市公司
func (m HongKong) queryCompanies(page, api string) ([]Company, error) {

	body, err := net.DownloadStringRetry(page, retryTimes, retryIntervalSeconds)
	if err != nil {
		return nil, err
	}

	regexToken := regexp.MustCompile(`\"Base64-AES-Encrypted-Token\";\s*?return \"([^\"]+?)\";`)

	matches := regexToken.FindStringSubmatch(body)
	if len(matches) != 2 {
		return nil, errors.New("获取token失败")
	}

	body, err = net.DownloadStringRetry(fmt.Sprintf(api, matches[1], time.Now().UnixNano()), retryTimes, retryIntervalSeconds)
	if err != nil {
		return nil, err
	}

	regexCode := regexp.MustCompile(`\"ric\":\"(\d{2,5})\.HK\"\S+?\"nm\":\"([^\"]+)\"`)
	group := regexCode.FindAllStringSubmatch(body, -1)

	var companies []Company
	for _, section := range group {
		companies = append(companies, Company{Code: section[1], Name: section[2]})
	}

	return companies, nil
}

// YahooQueryCode 雅虎查询代码
func (m HongKong) YahooQueryCode(company Company) string {

	return company.Code + ".HK"
}
