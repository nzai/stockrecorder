package quote

import (
	"io"
	"time"

	"github.com/nzai/go-utility/io/ioutil"
	"go.uber.org/zap"
)

// Exchange 交易所
type Exchange struct {
	Code        string
	Name        string
	Location    *time.Location
	YahooSuffix string
}

// Marshal 序列化
func (e Exchange) Marshal(w io.Writer) error {

	bw := ioutil.NewBinaryWriter(w)
	err := bw.String(e.Code)
	if err != nil {
		zap.L().Error("write exchange code failed", zap.Error(err), zap.Any("code", e.Code))
		return err
	}

	err = bw.String(e.Name)
	if err != nil {
		zap.L().Error("write exchange name failed", zap.Error(err), zap.Any("name", e.Name))
		return err
	}

	err = bw.String(e.Location.String())
	if err != nil {
		zap.L().Error("write exchange location failed", zap.Error(err), zap.Any("location", e.Location.String()))
		return err
	}

	err = bw.String(e.YahooSuffix)
	if err != nil {
		zap.L().Error("write exchange yahoo suffix failed", zap.Error(err), zap.Any("yahoo suffix", e.YahooSuffix))
		return err
	}

	return nil
}

// Unmarshal 反序列化
func (e *Exchange) Unmarshal(r io.Reader) error {

	br := ioutil.NewBinaryReader(r)
	code, err := br.String()
	if err != nil {
		zap.L().Error("read exchange code failed", zap.Error(err))
		return err
	}

	name, err := br.String()
	if err != nil {
		zap.L().Error("read exchange name failed", zap.Error(err))
		return err
	}

	locationName, err := br.String()
	if err != nil {
		zap.L().Error("read exchange location name failed", zap.Error(err))
		return err
	}

	location, err := time.LoadLocation(locationName)
	if err != nil {
		zap.L().Error("read exchange location failed", zap.Error(err), zap.String("location name", locationName))
		return err
	}

	yahooSuffix, err := br.String()
	if err != nil {
		zap.L().Error("read exchange yahoo suffix failed", zap.Error(err))
		return err
	}

	e.Code = code
	e.Name = name
	e.Location = location
	e.YahooSuffix = yahooSuffix

	return nil
}
