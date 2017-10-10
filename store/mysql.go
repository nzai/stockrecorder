package store

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/nzai/stockrecorder/market"

	// mysql driver
	_ "github.com/go-sql-driver/mysql"
)

// MysqlConfig Mysql存储配置
type MysqlConfig struct {
	ConnectionString string
}

// Mysql Mysql存储
type Mysql struct {
	db *sql.DB
}

// NewMysql 新建文件系统存储服务
func NewMysql(config MysqlConfig) *Mysql {

	db, err := sql.Open("mysql", config.ConnectionString)
	if err != nil {
		log.Fatalf("创建数据库连接失败: %v", err)
	}

	return &Mysql{db}
}

// Exists 判断是否存在
func (s Mysql) Exists(_market market.Market, date time.Time) (bool, error) {

	row := s.db.QueryRow("select count(0) from quote where type = ? and start = ? and duration = ?", _market.Name(), date.Unix(), int64(time.Hour)*24)
	var count int64
	err := row.Scan(&count)

	return count > 0, err
}

// Save 保存
func (s Mysql) Save(quote market.DailyQuote) error {

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	err = s.saveQuote(tx, quote)
	if err != nil {
		err1 := tx.Rollback()
		if err1 != nil {
			return err1
		}

		return err
	}

	return tx.Commit()
}

// saveQuote 保存Quote
func (s Mysql) saveQuote(tx *sql.Tx, quote market.DailyQuote) error {

	quotes := quote.ToQuote()
	count := len(quotes)
	if count == 0 {
		return nil
	}

	batchSize := 100000
	start := 0
	for start < count {
		end := start + batchSize
		if end > count {
			end = count
		}

		buffer := bytes.NewBufferString("insert into quote(type,`key`,start,duration,open,close,max,min,volume) values")
		for index, quote := range quotes[start:end] {
			if index > 0 {
				buffer.WriteString(",")
			}

			buffer.WriteString(fmt.Sprintf(`("%s","%s",%d,%d,%.2f,%.2f,%.2f,%.2f,%d)`, quote.Type, quote.Key, quote.Start, quote.Duration, quote.Open, quote.Close, quote.Max, quote.Min, quote.Volume))
		}

		_, err := tx.Exec(buffer.String())
		if err != nil {
			log.Print(buffer.String())
			return err
		}

		log.Printf("[%s] %s count:%d / %d", quote.Market.Name(), quote.Date.Format("2006-01-02"), start, count)

		start += batchSize
	}

	return nil
}

// Load 读取
func (s Mysql) Load(_market market.Market, date time.Time) (market.DailyQuote, error) {

	mdq := market.DailyQuote{Market: _market, Date: date}

	quotes, err := s.loadCompany(_market, date)
	if err != nil {
		return mdq, err
	}

	var lastCode string
	var lastStart int
	for index, quote := range quotes {

		if quote.Key == lastCode || lastStart == 0 {
			continue
		}

		var cq market.CompanyDailyQuote
		cq.FromQuote(quotes[lastStart:index])

		mdq.Quotes = append(mdq.Quotes, cq)
	}

	return mdq, nil
}

func (s Mysql) loadCompany(_market market.Market, date time.Time) ([]market.Quote, error) {

	rows, err := s.db.Query("select id, type, `key`, start, duration, open, close, max, min, volume from quote where type = ? and start >= ? and start < ? duration = ? order by `key` asc, start asc",
		_market.Name(),
		date.Unix(),
		date.AddDate(0, 0, 1).Unix(),
		int64(time.Minute),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var quotes []market.Quote
	for rows.Next() {
		var quote market.Quote
		err = quote.ScanRows(rows)
		if err != nil {
			return nil, err
		}

		quotes = append(quotes, quote)
	}

	return quotes, nil
}
