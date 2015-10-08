package market

import (
	"errors"
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
	return filepath.Join(config.Get().DataDir, market.Name(), companiesFileName)
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
