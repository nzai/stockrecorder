package quote

import (
	"fmt"
	"io"

	"github.com/nzai/go-utility/io/ioutil"
	"go.uber.org/zap"
)

// Company 公司
type Company struct {
	Name string // 名称
	Code string // 代码
}

// Marshal 序列化
func (c Company) Marshal(w io.Writer) error {
	bw := ioutil.NewBinaryWriter(w)
	err := bw.String(c.Code)
	if err != nil {
		zap.L().Error("write company code failed", zap.Error(err), zap.Any("company", c))
		return err
	}

	err = bw.String(c.Name)
	if err != nil {
		zap.L().Error("write company name failed", zap.Error(err), zap.Any("company", c))
		return err
	}

	return nil
}

// Unmarshal 反序列化
func (c *Company) Unmarshal(r io.Reader) error {
	br := ioutil.NewBinaryReader(r)

	code, err := br.String()
	if err != nil {
		zap.L().Error("read company code failed", zap.Error(err))
		return err
	}

	name, err := br.String()
	if err != nil {
		zap.L().Error("read company name failed", zap.Error(err))
		return err
	}

	c.Code = code
	c.Name = name

	return nil
}

// Equal 是否相同
func (c Company) Equal(s Company) error {

	if c.Code != s.Code {
		return fmt.Errorf("Company Code不相等:c.Code=[%s] s.Code=[%s]", c.Code, s.Code)
	}

	if c.Name != s.Name {
		return fmt.Errorf("Company Name不相等:c.Name=[%s] s.Name=[%s]", c.Name, s.Name)
	}

	return nil
}

// CompanyList 公司列表
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
