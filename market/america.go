package market

import (
	"encoding/csv"
	"fmt"
	"sort"
	"strings"

	"github.com/nzai/go-utility/net"
	"github.com/nzai/stockrecorder/quote"
)

// America 美国证券市场
type America struct{}

// Name 名称
func (m America) Name() string {
	return "America"
}

// Timezone 所处时区
func (m America) Timezone() string {
	return "America/New_York"
}

// Companies 上市公司
func (m America) Companies() ([]quote.Company, error) {

	urls := [...]string{
		"http://www.nasdaq.com/screening/companies-by-industry.aspx?exchange=NASDAQ&render=download",
		"http://www.nasdaq.com/screening/companies-by-industry.aspx?exchange=NYSE&render=download",
		"http://www.nasdaq.com/screening/companies-by-industry.aspx?exchange=AMEX&render=download",
	}

	var list []quote.Company
	for _, url := range urls {

		//	尝试从网络获取实时上市公司列表
		csv, err := net.DownloadStringRetry(url, retryTimes, retryIntervalSeconds)
		if err != nil {
			return nil, err
		}

		//	解析CSV
		companies, err := m.parseCSV(csv)
		if err != nil {
			return nil, err
		}

		list = append(list, companies...)
	}

	//	按Code排序
	sort.Sort(quote.CompanyList(list))

	return list, nil
}

//	解析CSV
func (m America) parseCSV(content string) ([]quote.Company, error) {

	reader := csv.NewReader(strings.NewReader(content))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	dict := make(map[string]bool, 0)
	var companies []quote.Company
	for _, parts := range records[1:] {
		if len(parts) < 2 {
			return nil, fmt.Errorf("错误的美股上市公司CSV格式:%v", parts)
		}

		if strings.Contains(parts[0], "^") {
			continue
		}

		//	去重
		if _, found := dict[parts[0]]; found {
			continue
		}
		dict[parts[0]] = true

		companies = append(companies, quote.Company{
			Code: strings.Trim(parts[0], " "),
			Name: strings.Trim(parts[1], " ")})
	}

	return companies, nil
}

// YahooQueryCode 雅虎查询代码
func (m America) YahooQueryCode(company quote.Company) string {
	return company.Code
}
