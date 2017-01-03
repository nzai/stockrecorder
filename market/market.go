package market

import "errors"

const (
	retryTimes           = 50
	retryIntervalSeconds = 10
)

// Market 市场
type Market interface {
	//	名称
	Name() string
	//	时区
	Timezone() string
	//	获取上市公司列表
	Companies() ([]Company, error)

	// 用于雅虎财经接口的查询代码后缀
	YahooQueryCode(company Company) string
}

var (
	// ErrUnknownMarket 未知的市场
	ErrUnknownMarket = errors.New("未知的市场")
)

// Get 获取市场
func Get(name string) (*Market, error) {

	markets := []Market{America{}, China{}, HongKong{}}
	for _, market := range markets {
		if market.Name() != name {
			continue
		}

		return &market, nil
	}

	return nil, ErrUnknownMarket
}
