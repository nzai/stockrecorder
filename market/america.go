package market

import (
	"encoding/csv"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/nzai/stockrecorder/io"
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
func (m America) UpdateCompanies() ([]Company, error) {

	urls := [...]string{
		"http://www.nasdaq.com/screening/companies-by-industry.aspx?exchange=NASDAQ&render=download",
		"http://www.nasdaq.com/screening/companies-by-industry.aspx?exchange=NYSE&render=download",
		"http://www.nasdaq.com/screening/companies-by-industry.aspx?exchange=AMEX&render=download",
	}

	var err error
	list := make([]Company, 0)
	for _, url := range urls {

		//	尝试从网络获取实时上市公司列表
		for try := 0; try < retryCount; try++ {
			csv, err := io.GetString(url)
			if err == nil {
				companies, err := m.parseCSV(csv)
				if err == nil {
					list = append(list, companies...)
					break
				}
			}

			time.Sleep(retryDelaySeconds * time.Second)
		}
	}

	if err != nil {
		return nil, err
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
			return nil, errors.New(fmt.Sprintf("错误的美股上市公司CSV格式:%v", parts))
		}

		companies = append(companies, Company{Code: parts[0], Name: parts[1]})
	}

	return companies, nil
}
