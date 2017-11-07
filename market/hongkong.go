package market

import (
	"bytes"
	"sort"
	"strings"

	"github.com/360EntSecGroup-Skylar/excelize"
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

	url := "http://sc.hkex.com.hk/TuniS/www.hkex.com.hk/chi/services/trading/securities/securitieslists/ListOfSecurities_c.xlsx"

	buffer, err := net.DownloadBufferRetry(url, retryTimes, retryIntervalSeconds)
	if err != nil {
		return nil, err
	}

	companies, err := m.parseXlsx(buffer)
	if err != nil {
		return nil, err
	}

	//	按Code排序
	sort.Sort(CompanyList(companies))

	return companies, nil
}

//	解析香港证券交易所上市公司
func (m HongKong) parseXlsx(buffer []byte) ([]Company, error) {

	file, err := excelize.OpenReader(bytes.NewReader(buffer))
	if err != nil {
		return nil, err
	}

	rows := file.GetRows("ListOfSecuritites")
	var companies []Company
	for _, values := range rows {

		if strings.Trim(values[0], " ") == "" {
			continue
		}

		companies = append(companies, Company{
			Code: values[0],
			Name: values[1],
		})
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
