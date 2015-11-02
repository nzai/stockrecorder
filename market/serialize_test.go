package market

import (
	"testing"
	"time"

	"log"

	"github.com/nzai/stockrecorder/config"
)

func init() {

	//	读取配置文件
	err := config.SetRootDir(`g:\gohome\src\github.com\nzai\stockrecorder\`)
	if err != nil {
		log.Fatal("读取配置文件发生错误: ", err)
		return
	}
}

func TestLoadPeroid60(t *testing.T) {

	date, _ := time.Parse("20060102", "20151001")
	peroids, err := loadPeroid60(America{}, "AAPL", date)
	if err != nil {
		log.Fatal("读取分时数据发生错误: ", err)
		return
	}

	t.Log(peroids)
}
