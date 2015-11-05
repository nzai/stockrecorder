package server

import (
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo"
	"github.com/nzai/stockrecorder/market"
	"github.com/nzai/stockrecorder/server/result"
)

//	注册路由
func registerRoute(e *echo.Echo) {

	e.Get("/", welcome)

	e.Get("/1m/query", queryPeroid60)
}

func welcome(c *echo.Context) error {
	return c.String(http.StatusOK, "Welcome to stockrecorder http service!")
}

//	查询分时数据
func queryPeroid60(c *echo.Context) error {

	_market := c.Query("market")
	code := c.Query("code")
	start := c.Query("start")
	end := c.Query("end")

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
	peroids, err := market.QueryPeroid60(_market, code, _start, _end)
	if err != nil {
		log.Printf("[Query]\t查询分时数据发生错误(m=%s c=%s s=%s e=%s):%s", _market, code, start, end, err.Error())
		return c.JSON(http.StatusOK, result.Failed("查询分时数据发生错误"))
	}

	resultList := make([][]float32, 0)
	for _, p := range peroids {
		if p.Volume == 0 {
			continue
		}

		resultList = append(resultList, []float32{p.Open, p.Close, p.High, p.Low, float32(p.Volume)})
	}

	return c.JSON(http.StatusOK, result.Create(resultList))
}
