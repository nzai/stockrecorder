package market

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/nzai/stockrecorder/config"
	"github.com/nzai/stockrecorder/io"
)

const (
	companiesFileName = "companies.txt"
)

//	公司
type Company struct {
	Name string
	Code string
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

//	上市公司列表保存路径
func storePath(market Market) string {
	return filepath.Join(config.GetDataDir(), market.Name(), companiesFileName)
}

//	保存上市公司列表到文件
func (l CompanyList) Save(market Market) error {
	lines := make([]string, 0)
	companies := ([]Company)(l)
	for _, company := range companies {
		lines = append(lines, fmt.Sprintf("%s\t%s\n", company.Code, company.Name))
	}

	return io.WriteLines(storePath(market), lines)
}

//	从存档读取上市公司列表
func (l *CompanyList) Load(market Market) error {
	lines, err := io.ReadLines(storePath(market))
	if err != nil {
		return err
	}

	companies := make([]Company, 0)
	for _, line := range lines {
		parts := strings.Split(line, "\t")
		if len(parts) != 2 {
			return errors.New("错误的上市公司列表格式")
		}

		companies = append(companies, Company{Code: parts[0], Name: parts[1]})
	}

	cl := CompanyList(companies)
	l = &cl

	return nil
}

//	抓取市场上市公司信息
func (l *CompanyList) Refresh(market Market) error {

	//	尝试更新上市公司列表
	log.Printf("[%s]\t更新上市公司列表-开始", market.Name())
	companies, err := market.LastestCompanies()
	if err != nil {

		//	如果更新失败，则尝试从上次的存档文件中读取上市公司列表
		log.Printf("[%s]\t更新上市公司列表失败，尝试从存档读取", market.Name())
		err = l.Load(market)
		if err != nil {
			log.Printf("[%s]\t尝试从存档读取上市公司列表-失败", market.Name())
			return err
		}

		log.Printf("[%s]\t尝试从存档读取上市公司列表-成功,共%d家上市公司", market.Name(), len(companies))

		return nil
	}

	cl := CompanyList(companies)
	l = &cl

	//	存档
	err = l.Save(market)
	if err != nil {
		return err
	}

	log.Printf("[%s]\t更新上市公司列表-成功,共%d家上市公司", market.Name(), len(companies))

	return nil
}
