package market

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"errors"
	"regexp"
	"strconv"

	"github.com/nzai/go-utility/net"
	"github.com/nzai/stockrecorder/quote"
)

// England 英国证券市场
type England struct{}

// Name 名称
func (m England) Name() string {
	return "England"
}

// Timezone 所处时区
func (m England) Timezone() string {
	return "Europe/London"
}

// Companies 上市公司
func (m England) Companies() ([]quote.Company, error) {

	regexPage := regexp.MustCompile(`Page \d+ of (\d+)`)
	regexCompany := regexp.MustCompile(`<td scope=\"row\" class=\"name\">(\w+)</td>(?s).*?>([^<]+?)</a>`)

	url := "http://www.londonstockexchange.com/exchange/prices-and-markets/stocks/prices-search/stock-prices-search.html"
	body, err := net.DownloadStringRetry(url, retryTimes, retryIntervalSeconds)
	if err != nil {
		return nil, err
	}

	matches := regexPage.FindStringSubmatch(body)
	if len(matches) != 2 {
		return nil, errors.New("London的网页内容有变")
	}

	totalPages, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil, err
	}

	groups := regexCompany.FindAllStringSubmatch(body, -1)
	pageSize := len(groups)

	ch := make(chan bool, 32)
	defer close(ch)

	wg := &sync.WaitGroup{}
	wg.Add(totalPages)

	companies := make([]quote.Company, 0, totalPages*pageSize)
	for index := 1; index <= totalPages; index++ {

		ch <- true
		go func(_index int, _wg *sync.WaitGroup) {

			_companies, err := m.queryCompanies(fmt.Sprintf("%s?page=%d", url, _index))
			if err == nil {
				companies = append(companies, _companies...)
			}

			<-ch
			_wg.Done()
		}(index, wg)

		// 限流
		ch <- false
	}

	//	阻塞，直到抓取所有
	wg.Wait()

	//	按Code排序
	sort.Sort(quote.CompanyList(companies))

	return companies, nil
}

//	解析CSV
func (m England) queryCompanies(url string) ([]quote.Company, error) {

	//	尝试从网络获取实时上市公司列表
	body, err := net.DownloadStringRetry(url, retryTimes, retryIntervalSeconds)
	if err != nil {
		return nil, err
	}

	regex := regexp.MustCompile(`<td scope=\"row\" class=\"name\">(\w+)</td>(?s).*?>([^<]+?)</a>`)
	groups := regex.FindAllStringSubmatch(body, -1)

	var companies []quote.Company
	for _, group := range groups {

		companies = append(companies, quote.Company{
			Code: strings.Trim(group[1], " "),
			Name: strings.Trim(group[2], " ")})
	}

	return companies, nil
}

// YahooQueryCode 雅虎查询代码
func (m England) YahooQueryCode(company quote.Company) string {
	return company.Code + ".L"
}
