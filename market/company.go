package market

import (
	"fmt"
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

	lines := make([]string, 0)
	companies := ([]Company)(l)
	for _, company := range companies {
		lines = append(lines, fmt.Sprintf("%s\t%s", company.Code, company.Name))
	}

	return io.WriteLines(filepath.Join(config.Get().DataDir, market.Name(), companiesFileName), lines)
}

//	从存档读取上市公司列表
func (l *CompanyList) Load(market Market) error {

	lines, err := io.ReadLines(filepath.Join(config.Get().DataDir, market.Name(), companiesFileName))
	if err != nil {
		return err
	}

	companies := make([]Company, 0)
	for _, line := range lines {
		parts := strings.Split(line, "\t")
		if len(parts) != 2 {
			return fmt.Errorf("[%s]\t上市公司文件格式有错误: %s", market.Name(), line)
		}

		companies = append(companies, Company{
			Market: market.Name(),
			Code:   parts[0],
			Name:   parts[1]})
	}

	cl := CompanyList(companies)
	l = &cl

	return nil
}
