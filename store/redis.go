package store

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/nzai/stockrecorder/market"
	"gopkg.in/redis.v5"
)

var (
	// ErrUnknownQuoteFormat 未知的Quote格式
	ErrUnknownQuoteFormat = errors.New("未知的Quote格式")
)

// RedisConfig Redis存储配置
type RedisConfig struct {
	Options redis.Options
}

// Redis Redis存储
type Redis struct {
	client *redis.Client
}

// NewRedis 新建Redis存储
func NewRedis(config RedisConfig) *Redis {

	client := redis.NewClient(&config.Options)

	_, err := client.Ping().Result()
	if err != nil {
		log.Fatal(err)
	}

	return &Redis{client: client}
}

// Exists 判断是否存在
func (s Redis) Exists(_market market.Market, date time.Time) (bool, error) {

	// key:america:20160101:offset value:18000
	offsetKey := fmt.Sprintf("%s:%s:offset", strings.ToLower(_market.Name()), date.Format("20060102"))

	return s.client.Exists(offsetKey).Result()
}

// Save 保存
func (s Redis) Save(quote market.DailyQuote) error {

	for _, cdq := range quote.Quotes {

		err := s.saveCompanyDailyQuote(quote.Market, quote.Date, cdq)
		if err != nil {
			return err
		}
	}

	// key:america:20160101:offset value:18000
	offsetKey := fmt.Sprintf("%s:%s:offset", strings.ToLower(quote.Market.Name()), quote.Date.Format("20060102"))
	err := s.client.Set(offsetKey, strconv.Itoa(quote.UTCOffset), 0).Err()
	if err != nil {
		return err
	}

	return nil
}

// saveCompanyDailyQuote 保存公司报价
func (s Redis) saveCompanyDailyQuote(_market market.Market, date time.Time, cdq market.CompanyDailyQuote) error {

	// key:america:20160101:aapl:name value:Apple Inc.
	nameKey := fmt.Sprintf("%s:%s:%s:name", strings.ToLower(_market.Name()), date.Format("20060102"), strings.ToLower(cdq.Code))
	err := s.client.Set(nameKey, cdq.Name, 0).Err()
	if err != nil {
		return err
	}

	// key:america:20160101:company value:[a aa aapl fb ibm ...]
	companyKey := fmt.Sprintf("%s:%s:company", strings.ToLower(_market.Name()), date.Format("20060102"))
	err = s.client.SAdd(companyKey, strings.ToLower(cdq.Code)).Err()
	if err != nil {
		return err
	}

	err = s.saveQuoteSerie(_market, date, cdq.Code, "pre", cdq.Pre)
	if err != nil {
		return err
	}

	err = s.saveQuoteSerie(_market, date, cdq.Code, "regular", cdq.Regular)
	if err != nil {
		return err
	}

	err = s.saveQuoteSerie(_market, date, cdq.Code, "post", cdq.Post)
	if err != nil {
		return err
	}

	return nil
}

// saveQuoteSerie 保存报价序列
func (s Redis) saveQuoteSerie(_market market.Market, date time.Time, code, typeName string, series market.QuoteSeries) error {

	if series.Count == 0 {
		return nil
	}

	// key:america:20160101:aapl:pre field:timestamp value:open|close|max|min|volume
	key := fmt.Sprintf("%s:%s:%s:%s", strings.ToLower(_market.Name()), date.Format("20060102"), strings.ToLower(code), typeName)

	values := make(map[string]string, series.Count)
	for index := 0; index < int(series.Count); index++ {

		values[strconv.Itoa(int(series.Timestamp[index]))] = fmt.Sprintf("%d|%d|%d|%d|%d",
			series.Open[index],
			series.Close[index],
			series.Max[index],
			series.Min[index],
			series.Volume[index],
		)
	}

	return s.client.HMSet(key, values).Err()
}

// Load 读取
func (s Redis) Load(_market market.Market, date time.Time) (market.DailyQuote, error) {
	mdq := market.DailyQuote{Market: _market, Date: date}

	// key:america:20160101:offset value:18000
	offsetKey := fmt.Sprintf("%s:%s:offset", strings.ToLower(_market.Name()), date.Format("20060102"))
	offsetString, err := s.client.Get(offsetKey).Result()
	if err != nil {
		return mdq, err
	}

	mdq.UTCOffset, err = strconv.Atoi(offsetString)
	if err != nil {
		return mdq, err
	}

	companyKey := fmt.Sprintf("%s:%s:company", strings.ToLower(_market.Name()), date.Format("20060102"))
	companyCodes, err := s.client.SMembers(companyKey).Result()
	if err != nil {
		return mdq, err
	}

	// 按code排序
	sort.Strings(companyCodes)
	for _, code := range companyCodes {

		cdq, err := s.loadCompanyDailyQuote(_market, date, code)
		if err != nil {
			return mdq, nil
		}

		mdq.Quotes = append(mdq.Quotes, cdq)
	}

	return mdq, nil
}

// loadCompanyDailyQuote 读取公司报价
func (s Redis) loadCompanyDailyQuote(_market market.Market, date time.Time, code string) (market.CompanyDailyQuote, error) {

	cdq := market.CompanyDailyQuote{Company: market.Company{Code: strings.ToUpper(code)}}

	// key:america:20160101:aapl:name value:Apple Inc.
	nameKey := fmt.Sprintf("%s:%s:%s:name", strings.ToLower(_market.Name()), date.Format("20060102"), strings.ToLower(code))
	name, err := s.client.Get(nameKey).Result()
	if err != nil {
		return cdq, err
	}
	cdq.Name = name

	return cdq, nil
}

// loadQuoteSerie 读取报价序列
func (s Redis) loadQuoteSerie(_market market.Market, date time.Time, code, typeName string) (market.QuoteSeries, error) {

	qs := market.QuoteSeries{}

	// key:america:20160101:aapl:pre field:timestamp value:open|close|max|min|volume
	key := fmt.Sprintf("%s:%s:%s:%s", strings.ToLower(_market.Name()), date.Format("20060102"), strings.ToLower(code), typeName)
	exists, err := s.client.Exists(key).Result()
	if err != nil {
		return qs, err
	}

	if !exists {
		return qs, nil
	}

	values, err := s.client.HGetAll(key).Result()
	if err != nil {
		return qs, nil
	}

	count := len(values)
	qs.Count = uint32(count)
	qs.Timestamp = make([]uint32, count)
	qs.Open = make([]uint32, count)
	qs.Close = make([]uint32, count)
	qs.Max = make([]uint32, count)
	qs.Min = make([]uint32, count)
	qs.Volume = make([]uint32, count)

	timestamps := make([]string, len(values))
	for index, timestamp := range timestamps {
		timestamps[index] = timestamp
	}
	sort.Strings(timestamps)

	for index, timestamp := range timestamps {

		parts := strings.Split(values[timestamp], "|")
		if len(parts) != 5 {
			return qs, ErrUnknownQuoteFormat
		}

		ts, err := strconv.Atoi(timestamp)
		if err != nil {
			return qs, err
		}
		qs.Timestamp[index] = uint32(ts)

		open, err := strconv.Atoi(parts[0])
		if err != nil {
			return qs, err
		}
		qs.Open[index] = uint32(open)

		_close, err := strconv.Atoi(parts[1])
		if err != nil {
			return qs, err
		}
		qs.Close[index] = uint32(_close)

		max, err := strconv.Atoi(parts[2])
		if err != nil {
			return qs, err
		}
		qs.Max[index] = uint32(max)

		min, err := strconv.Atoi(parts[3])
		if err != nil {
			return qs, err
		}
		qs.Min[index] = uint32(min)

		volume, err := strconv.Atoi(parts[4])
		if err != nil {
			return qs, err
		}
		qs.Volume[index] = uint32(volume)
	}

	return qs, nil
}
