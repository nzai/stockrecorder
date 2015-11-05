package market

import (
	"fmt"
	"strings"
	"time"
)

//	查询
func QueryPeroid60(market, code string, start, end time.Time) ([]Peroid60, error) {

	_market, found := markets[market]
	if !found {
		return nil, fmt.Errorf("[Query]\t未能找到市场%s", market)
	}

	_code := strings.ToUpper(code)

	peroids := make([]Peroid60, 0)
	for date := start; end == date || end.After(date); date = date.Add(time.Hour * 24) {

		if !isExists(_market, _code, date, regularSuffix) {
			continue
		}

		//	读取
		dayPeroids, err := loadPeroid60(_market, _code, date)
		if err != nil {
			return nil, err
		}

		if len(dayPeroids) > 0 {
			peroids = append(peroids, dayPeroids...)
		}
	}

	// todo
	return peroids, nil
}
