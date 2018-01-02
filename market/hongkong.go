package market

import (
	"errors"
	"fmt"
	"github.com/nzai/go-utility/net"
	"regexp"
	"sort"
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

	url := "http://www.hkex.com.hk/Market-Data/Securities-Prices/Equities?sc_lang=zh-HK"

	body, err := net.DownloadStringRetry(url, retryTimes, retryIntervalSeconds)
	if err != nil {
		return nil, err
	}

	token, err := m.getToken(body)
	if err != nil {
		return nil, err
	}

	companies, err := m.queryCompanies(token)
	if err != nil {
		return nil, err
	}

	//	按Code排序
	sort.Sort(CompanyList(companies))

	return companies, nil
}

// getToken 获取访问api的token
func (m HongKong) getToken(body string) (string, error) {

	regex := regexp.MustCompile(`\"Base64-AES-Encrypted-Token\";\s*?return \"([^\"]+?)\";`)

	matches := regex.FindStringSubmatch(body)
	if len(matches) != 2 {
		return "", errors.New("获取token失败")
	}

	return matches[1], nil
}

//	解析香港证券交易所上市公司
func (m HongKong) queryCompanies(token string) ([]Company, error) {

	url := fmt.Sprintf("https://www1.hkex.com.hk/hkexwidget/data/getequityfilter?lang=chi&sort=5&order=0&all=1&token=%s", token)

	body, err := net.DownloadStringRetry(url, retryTimes, retryIntervalSeconds)
	if err != nil {
		return nil, err
	}

	regex := regexp.MustCompile(`\"ric\":\"(\d{2,5})\.HK\"\S+?\"nm\":\"([^\"]+)\"`)
	group := regex.FindAllStringSubmatch(body, -1)

	var companies []Company
	for _, section := range group {
		companies = append(companies, Company{Code: section[1], Name: section[2]})
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
