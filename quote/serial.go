package quote

import (
	"fmt"
	"io"

	"github.com/nzai/go-utility/io/ioutil"
	"go.uber.org/zap"
)

// SerialType 报价序列类型
type SerialType int

const (
	// SerialTypePre 盘前
	SerialTypePre SerialType = iota + 1
	// SerialTypeRegular 盘中
	SerialTypeRegular
	// SerialTypePost 盘后
	SerialTypePost
)

func (t SerialType) String() string {
	switch t {
	case SerialTypePre:
		return "Pre"
	case SerialTypeRegular:
		return "Regular"
	case SerialTypePost:
		return "Post"
	default:
		return fmt.Sprintf("unknown quote serial type: %d", t)
	}
}

// Serial 报价序列
type Serial []Quote

// // NewQuoteSerial 新建报价序列
// func NewSerial(qs market.QuoteSeries) *Serial {

// 	quotes := make([]Quote, int(qs.Count))
// 	for index := 0; index < int(qs.Count); index++ {
// 		quotes[index] = Quote{
// 			Timestamp: int64(qs.Timestamp[index]),
// 			Open:      float32(qs.Open[index]) / 100,
// 			Close:     float32(qs.Close[index]) / 100,
// 			Max:       float32(qs.Max[index]) / 100,
// 			Min:       float32(qs.Min[index]) / 100,
// 			Volume:    int64(qs.Volume[index]),
// 		}
// 	}
// 	s := QuoteSerial(quotes)

// 	return &s
// }

// Marshal 序列化
func (s Serial) Marshal(w io.Writer) error {

	err := ioutil.NewBinaryWriter(w).Int(len(s))
	if err != nil {
		zap.L().Error("marshal quote serial length failed", zap.Error(err), zap.Int("length", len(s)))
		return err
	}

	for _, quote := range s {
		err = quote.Marshal(w)
		if err != nil {
			zap.L().Error("marshal quote failed", zap.Error(err), zap.Any("quote", quote))
			return err
		}
	}

	return nil
}

// Unmarshal 反序列化
func (s *Serial) Unmarshal(r io.Reader) error {

	count, err := ioutil.NewBinaryReader(r).Int()
	if err != nil {
		zap.L().Error("unmarshal quote serial length failed", zap.Error(err))
		return err
	}

	*s = make([]Quote, count)
	for index := 0; index < count; index++ {

		err = (*s)[index].Unmarshal(r)
		if err != nil {
			zap.L().Error("unmarshal quote failed", zap.Error(err))
			return err
		}
	}

	return nil
}

// Equal 是否相同
func (s Serial) Equal(q Serial) error {

	if len(s) != len(q) {
		return fmt.Errorf("quote serial length %d is different from %d", len(s), len(q))
	}

	for index, quote := range s {
		err := quote.Equal(q[index])
		if err != nil {
			zap.L().Error("quote is not equal", zap.Any("from", quote), zap.Any("to", q[index]))
			return err
		}
	}

	return nil
}

func (s Serial) Len() int {
	return len(s)
}
func (s Serial) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s Serial) Less(i, j int) bool {
	return s[i].Timestamp < s[j].Timestamp
}

// Rollup 汇总
func (s Serial) Rollup() *Quote {

	if len(s) == 0 {
		return &Quote{}
	}

	quote := &Quote{
		Timestamp: s[0].Timestamp,
		Open:      s[0].Open,
		Close:     s[len(s)-1].Close,
		High:      s[0].High,
		Low:       s[0].Low,
		Volume:    s[0].Volume,
	}

	for index := 1; index < len(s); index++ {
		if s[index].High > quote.High {
			quote.High = s[index].High
		}

		if s[index].Low < quote.Low {
			quote.Low = s[index].Low
		}

		quote.Volume += s[index].Volume
	}

	return quote
}
