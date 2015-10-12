package db

import (
	"testing"

	_ "gopkg.in/mgo.v2"
	//	"gopkg.in/mgo.v2/bson"
	"log"
	"time"

	"github.com/nzai/stockrecorder/config"
)

func init() {

	//	读取配置文件
	err := config.SetRootDir(`g:\gohome\src\github.com\nzai\stockrecorder\`)
	if err != nil {
		log.Fatal("读取配置文件错误: ", err)
		return
	}

	defer func() {
		// 捕获panic异常
		log.Print("发生了致命错误")
		if err := recover(); err != nil {
			log.Print(err)
		}
	}()
}

func TestSaveRaw60(t *testing.T) {

	raw := Raw60{
		Market:  "Test",
		Code:    "aa",
		Date:    time.Now(),
		Json:    "{abc:true}",
		Status:  1,
		Message: ""}

	SaveRaw60(raw)

	time.Sleep(time.Second * 2)
}

//func TestQuery(t *testing.T) {
//	session, err := Get()
//	if err != nil {
//		log.Printf("[DB]\t获取数据库连接失败:%s", err.Error())
//		return
//	}
//	defer session.Close()

//	m := bson.M{"Market": "Test", "Code": "aa"}
//	t.Logf("m:%v", m)
//	q := session.DB("stock").C("Raw60").Find(m)
//	t.Logf("q:%v", q)

//	count, err := q.Count()

//	t.Logf("count:%d", count)
//}
