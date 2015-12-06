package market

import (
	"encoding/csv"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/nzai/go-utility/net"
)

//	美股市场
type America struct{}

//	获取市场

func (m America) Name() string {
	return "America"
}

func (m America) Timezone() string {
	return "America/New_York"
}

//	更新上市公司列表
func (m America) Companies() ([]Company, error) {

	urls := [...]string{
		"http://www.nasdaq.com/screening/companies-by-industry.aspx?exchange=NASDAQ&render=download",
		"http://www.nasdaq.com/screening/companies-by-industry.aspx?exchange=NYSE&render=download",
		"http://www.nasdaq.com/screening/companies-by-industry.aspx?exchange=AMEX&render=download",
	}

	list := make([]Company, 0)
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
	sort.Sort(CompanyList(list))

	return list, nil
}

//	解析CSV
func (m America) parseCSV(content string) ([]Company, error) {

	reader := csv.NewReader(strings.NewReader(content))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	companies := make([]Company, 0)
	for _, parts := range records[1:] {
		if len(parts) < 2 {
			return nil, fmt.Errorf("错误的美股上市公司CSV格式:%v", parts)
		}

		if strings.Contains(parts[0], "^") {
			continue
		}

		companies = append(companies, Company{Market: m.Name(),
			Code: strings.Trim(parts[0], " "),
			Name: strings.Trim(parts[1], " ")})
	}

	return companies, nil
}

//	抓取
func (m America) Crawl(code string, day time.Time) (string, error) {
	return downloadCompanyDaily(m, code, code, day)
}
