package db

import (
	"log"
	"time"

	"gopkg.in/mgo.v2/bson"
)

const (
	retryTimes           = 50
	retryIntervalSeconds = 10
)

var (
	//	存储队列
	saveQueue chan Raw60 = nil
)

func init() {
	saveQueue = make(chan Raw60)

	go saveToDB()
}

//	以队列的方式保存到数据库
func saveToDB() {
	session, err := Get()
	if err != nil {
		log.Printf("[DB]\t获取数据库连接失败:%s", err.Error())
		return
	}
	defer session.Close()

	collection := session.DB("stock").C("Raw60")
	for {
		raw := <-saveQueue
		//	所有新增的记录都是未处理状态
		raw.Status = 0

		rawlist := make([]interface{}, 0)
		rawlist = append(rawlist, raw)

		//	如果队列长度超过1，就批量新增
		queueLength := len(saveQueue)
		if queueLength > 0 {
			//	读取队列
			for index := 0; index < queueLength; index++ {
				raw := <-saveQueue
				//	所有新增的记录都是未处理状态
				raw.Status = 0

				rawlist = append(rawlist, raw)
			}
		}

		var err error
		for times := retryTimes - 1; times >= 0; times-- {
			//	保存到数据库
			err = collection.Insert(rawlist...)
			if err == nil {
				break
			}

			if times > 0 {
				//	延时
				time.Sleep(time.Duration(retryIntervalSeconds) * time.Second)
			}
		}

		if err != nil {
			log.Printf("[DB]\t保存[%s %s %s]出错,已经重试%d次,不再重试:%s", raw.Market, raw.Code, raw.Date.Format("2006-01-02 15:04:05"), retryTimes, err.Error())
		}
	}
}

//	保存
func SaveRaw60(raw Raw60) {
	saveQueue <- raw
}

//	检查分析结果是否存在
func Raw60Exists(market, code string, date time.Time) (bool, error) {
	session, err := Get()
	if err != nil {
		return false, err
	}
	defer session.Close()

	count, err := session.DB("stock").C("Raw60").Find(bson.M{"Market": market, "Code": code, "Date": date}).Count()
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
