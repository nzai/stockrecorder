package server

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo"
	"github.com/nzai/stockrecorder/market"
	"github.com/nzai/stockrecorder/server/result"
)

//	注册路由
func registerRoute(e *echo.Echo) {

	e.Get("/", welcome)
	e.Favicon("favicon.ico")

	e.Get("/:market/:code/:start/:end/1m", queryPeroid60)
}

func welcome(c *echo.Context) error {
	return c.String(http.StatusOK, "Welcome to stockrecorder http service!")
}

//	查询分时数据
func queryPeroid60(c *echo.Context) error {

	_market := c.Param("market")
	code := c.Param("code")
	start := c.Param("start")
	end := c.Param("end")

	//	log.Printf("m=%s c=%s s=%s e=%s", _market, code, start, end)
	if _market == "" || code == "" || start == "" || end == "" {
		return c.JSON(http.StatusOK, result.Failed("查询参数为空"))
	}

	_start, err := time.Parse("20060102", start)
	if err != nil {
		return c.JSON(http.StatusOK, result.Failed("查询参数不正确"))
	}

	_end, err := time.Parse("20060102", end)
	if err != nil {
		return c.JSON(http.StatusOK, result.Failed("查询参数不正确"))
	}

	//	查询
	peroids, err := market.QueryPeroid60(strings.Title(_market), code, _start, _end)
	if err != nil {
		log.Printf("[Query]\t查询分时数据发生错误(m=%s c=%s s=%s e=%s):%s", _market, code, start, end, err.Error())
		return c.JSON(http.StatusOK, result.Failed("查询分时数据发生错误"))
	}

	resultList := make([]string, 0)
	for _, p := range peroids {
		if p.Volume == 0 {
			continue
		}

		resultList = append(resultList, fmt.Sprintf("%s|%.3f|%.3f|%.3f|%.3f|%d", p.Time, p.Open, p.Close, p.High, p.Low, p.Volume))
	}

	return c.JSON(http.StatusOK, result.Create(resultList))
}
