package quote

import (
	"io"
	"time"

	"github.com/nzai/go-utility/io/ioutil"
	"go.uber.org/zap"
)

// CompanyDailyQuote 公司每日报价
type CompanyDailyQuote struct {
	*Company
	Pre     Serial
	Regular Serial
	Post    Serial
}

// Marshal 序列化
func (q CompanyDailyQuote) Marshal(w io.Writer) error {
	err := q.Company.Marshal(w)
	if err != nil {
		zap.L().Error("write company failed", zap.Error(err), zap.Any("company", q.Company))
		return err
	}

	err = q.Pre.Marshal(w)
	if err != nil {
		zap.L().Error("write company pre serial failed", zap.Error(err), zap.Any("company", q.Company), zap.Int("pre serial count", len(q.Pre)))
		return err
	}

	err = q.Regular.Marshal(w)
	if err != nil {
		zap.L().Error("write company regular serial failed", zap.Error(err), zap.Any("company", q.Company), zap.Int("regular serial count", len(q.Regular)))
		return err
	}

	err = q.Post.Marshal(w)
	if err != nil {
		zap.L().Error("write company post serial failed", zap.Error(err), zap.Any("company", q.Company), zap.Int("post serial count", len(q.Post)))
		return err
	}

	return nil
}

// Unmarshal 反序列化
func (q *CompanyDailyQuote) Unmarshal(r io.Reader) error {

	err := q.Company.Unmarshal(r)
	if err != nil {
		zap.L().Error("read company failed", zap.Error(err))
		return err
	}

	err = q.Pre.Unmarshal(r)
	if err != nil {
		zap.L().Error("read company pre serial failed", zap.Error(err))
		return err
	}

	err = q.Regular.Unmarshal(r)
	if err != nil {
		zap.L().Error("read company regular serial failed", zap.Error(err))
		return err
	}

	err = q.Post.Unmarshal(r)
	if err != nil {
		zap.L().Error("read company post serial failed", zap.Error(err))
		return err
	}

	return nil
}

// ExchangeDailyQuote 交易所每日报价
type ExchangeDailyQuote struct {
	*Exchange
	Date      time.Time
	Companies map[string]*CompanyDailyQuote
}

// Marshal 序列化
func (q ExchangeDailyQuote) Marshal(w io.Writer) error {
	bw := ioutil.NewBinaryWriter(w)

	err := q.Exchange.Marshal(bw)
	if err != nil {
		zap.L().Error("marshal exchange failed", zap.Error(err), zap.Any("exchange", q.Exchange))
		return err
	}

	err = bw.Time(q.Date)
	if err != nil {
		zap.L().Error("marshal date failed", zap.Error(err), zap.Time("date", q.Date))
		return err
	}

	err = bw.Int(len(q.Companies))
	if err != nil {
		zap.L().Error("marshal company count failed", zap.Error(err), zap.Int("count", len(q.Companies)))
		return err
	}

	for _, cdq := range q.Companies {
		err = cdq.Marshal(bw)
		if err != nil {
			zap.L().Error("marshal company failed", zap.Error(err), zap.Any("company", cdq.Company))
			return err
		}
	}

	return nil
}

// Unmarshal 反序列化
func (q *ExchangeDailyQuote) Unmarshal(r io.Reader) error {
	br := ioutil.NewBinaryReader(r)

	exchange := new(Exchange)
	err := exchange.Unmarshal(br)
	if err != nil {
		zap.L().Error("ummarshal exchange failed", zap.Error(err))
		return err
	}

	date, err := br.Time()
	if err != nil {
		zap.L().Error("ummarshal date failed", zap.Error(err))
		return err
	}

	count, err := br.Int()
	if err != nil {
		zap.L().Error("ummarshal company count failed", zap.Error(err))
		return err
	}

	companies := make(map[string]*CompanyDailyQuote, count)
	for index := 0; index < count; index++ {
		cdq := new(CompanyDailyQuote)
		err := cdq.Unmarshal(br)
		if err != nil {
			zap.L().Error("ummarshal company daily quote failed", zap.Error(err))
			return err
		}

		companies[cdq.Code] = cdq
	}

	q.Exchange = exchange
	q.Date = date
	q.Companies = companies

	return nil
}
