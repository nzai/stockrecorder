package market

import (
	"fmt"
	"time"
)

//	查询
func QueryPeroid60(market, code string, start, end time.Time) ([]Peroid60, error) {

	_market, found := markets[market]
	if !found {
		return nil, fmt.Errorf("[Query]\t未能找到市场%s", market)
	}

	return loadPeroid(_market, code, start, end, "regular")
}
