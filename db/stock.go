package db

import (
	"fmt"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

)

//	检查分析结果是否存在
func DailyExists(market, code string, day time.Time) (bool, error) {
	session, err := Get()
	if err != nil {
		return false, err
	}
	defer session.Close()

	count, err := session.DB("stock").C("DailyResult").Find(bson.M{"Market": market, "Code": code, "Date": day}).Count()
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

//	保存分析结果
func DailySave(dar *DailyAnalyzeResult) error {
	session, err := Get()
	if err != nil {
		return err
	}
	defer session.Close()

	//	分析结果
	err = session.DB("stock").C("DailyResult").Insert(&dar.DailyResult)
	if err != nil {
		return fmt.Errorf("保存DailyResult出错:%s", err.Error())
	}

	//	Pre
	err = batchInsertPeroid60(session.DB("stock").C("Pre60"), dar.Pre)
	if err != nil {
		return fmt.Errorf("保存Pre60出错:%s", err.Error())
	}

	//	Regular
	err = batchInsertPeroid60(session.DB("stock").C("Regular60"), dar.Regular)
	if err != nil {
		return fmt.Errorf("保存Regular60出错:%s", err.Error())
	}

	//	Post
	err = batchInsertPeroid60(session.DB("stock").C("Post60"), dar.Post)
	if err != nil {
		return fmt.Errorf("保存Post60出错:%s", err.Error())
	}

	return nil
}

//	批量保存
func batchInsertPeroid60(c *mgo.Collection, array []Peroid60) error {

	if len(array) == 0 {
		return nil
	}

	var docs []interface{}
	for _, item := range array {
		docs = append(docs, item)
	}

	return c.Insert(docs...)
}
