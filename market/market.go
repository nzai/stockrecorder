package market

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/nzai/Tast/config"
	"github.com/nzai/stockrecorder/io"
)

//	市场更新
type MarketUpdater interface {
	//	名称
	Name() string
	//	时区
	Timezone() string
	//	获取上市公司列表
	//	Companies() ([]Company, error)
	//	更新上市公司列表
	UpdateCompanies() ([]Company, error)
}

//	市场
type Market struct {
	//	名称
	Name string
	//	时区
	Timezone string
	//	上市公司
	Companies []Company

	//	更新器
	updater MarketUpdater
}

func newMarket(updater MarketUpdater) (*Market, error) {

	m := &Market{
		Name:     updater.Name(),
		Timezone: updater.Timezone(),
		updater:  updater}

	//	尝试更新上市公司列表
	log.Printf("[%s]更新上市公司列表-开始", m.Name)
	companies, err := updater.UpdateCompanies()
	if err != nil {

		//	如果更新失败，则尝试从上次的存档文件中读取上市公司列表
		log.Printf("[%s]更新上市公司列表失败，尝试从存档读取", m.Name)
		companies, err = m.loadCompanies()
		if err != nil {
			log.Printf("[%s]尝试从存档读取上市公司列表-失败", m.Name)
			return nil, err
		}

		m.Companies = companies
		log.Printf("[%s]尝试从存档读取上市公司列表-成功,共%d家上市公司", m.Name, len(m.Companies))

		return m, nil
	}

	//	存档
	m.Companies = companies
	err = m.saveCompanies()
	if err != nil {
		return nil, err
	}
	log.Printf("[%s]更新上市公司列表-成功,共%d家上市公司", m.Name, len(m.Companies))

	return m, nil
}

//	上市公司保存路径
func (m Market) companiesSavePath() string {
	dir, _ := config.GetDataDir()
	return filepath.Join(dir, m.Name, companiesFileName)
}

//	保存上市公司到文件
func (m Market) saveCompanies() error {

	lines := make([]string, 0)
	for _, company := range m.Companies {
		lines = append(lines, fmt.Sprintf("%s\t%s\n", company.Code, company.Name))
	}

	return io.WriteLines(m.companiesSavePath(), lines)
}

// 从文件读取上市公司
func (m Market) loadCompanies() ([]Company, error) {

	lines, err := io.ReadLines(m.companiesSavePath())
	if err != nil {
		return nil, err
	}

	companies := make([]Company, 0)
	for _, line := range lines {
		parts := strings.Split(line, "\t")
		if len(parts) != 2 {
			return nil, errors.New("错误的上市公司列表格式")
		}

		companies = append(companies, Company{Code: parts[0], Name: parts[1]})
	}

	return companies, nil
}

func Start() error {
	_, err := newMarket(America{})

	return err
}
