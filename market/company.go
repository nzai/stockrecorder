package market

import (
	"fmt"

	"github.com/nzai/stockrecorder/db"
	"gopkg.in/mgo.v2/bson"
)

const (
	companiesFileName = "companies.txt"
)

//	公司
type Company struct {
	Market string
	Name   string
	Code   string
}

//	公司列表
type CompanyList []Company

func (l CompanyList) Len() int {
	return len(l)
}
func (l CompanyList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}
func (l CompanyList) Less(i, j int) bool {
	return l[i].Code < l[j].Code
}

//	保存上市公司列表到文件
func (l CompanyList) Save(market Market) error {

	//	连接数据库
	session, err := db.Get()
	if err != nil {
		return fmt.Errorf("[DB]\t获取数据库连接失败:%s", err.Error())
	}
	defer session.Close()

	companies := ([]Company)(l)
	list := make([]interface{}, 0)
	for _, company := range companies {
		list = append(list, company)
	}

	collection := session.DB("stock").C("Company")

	//	删除原有记录
	_, err = collection.RemoveAll(bson.M{"market": market.Name()})
	if err != nil {
		return fmt.Errorf("[DB]\t删除原有上市公司发生错误: %s", err.Error())
	}

	return collection.Insert(list...)
}

//	从存档读取上市公司列表
func (l *CompanyList) Load(market Market) error {
	//	连接数据库
	session, err := db.Get()
	if err != nil {
		return fmt.Errorf("[DB]\t获取数据库连接失败:%s", err.Error())
	}
	defer session.Close()

	var companies []Company
	err = session.DB("stock").C("Company").Find(bson.M{"market": market.Name()}).Sort("code").All(&companies)
	if err != nil {
		return fmt.Errorf("[DB]\t查询上市公司发生错误: %s", err.Error())
	}

	cl := CompanyList(companies)
	l = &cl

	return nil
}
