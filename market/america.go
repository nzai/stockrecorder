package market

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/nzai/stockrecorder/config"
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
func (m America) LastestCompanies() ([]Company, error) {

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

		companies = append(companies, Company{Code: strings.Trim(parts[0], " "), Name: strings.Trim(parts[1], " ")})
	}

	return companies, nil
}

//	抓取
func (m America) Crawl(market Market, company Company, day time.Time) error {
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

		pattern := "https://finance-yql.media.yahoo.com/v7/finance/chart/%s?period2=%d&period1=%d&interval=1m&indicators=quote&includeTimestamps=true&includePrePost=true&events=div%7Csplit%7Cearn&corsDomain=finance.yahoo.com"
		url := fmt.Sprintf(pattern, company.Code, end.Unix(), start.Unix())

		html, err := io.GetString(url)
		if err != nil {
			return err
		}

		//	写入文件
		return io.WriteString(filePath, html)
	}

	return nil
}
