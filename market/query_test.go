package market

import (
	"testing"
	"time"

	"log"

	"github.com/nzai/stockrecorder/config"
)

func init() {

	//	读取配置文件
	err := config.Init()
	if err != nil {
		log.Fatal("读取配置文件发生错误: ", err)
		return
	}
	
	markets["America"] = America{}
}

func TestQueryPeroid60(t *testing.T) {

	start, _ := time.Parse("20060102", "20151001")
	end, _ := time.Parse("20060102", "20151001")
	peroids, err := QueryPeroid60("America", "AAPL", start, end)
	if err != nil {
		log.Fatal("查询分时数据发生错误: ", err)
		return
	}

	t.Log(len(peroids))
}
