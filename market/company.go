package market

import (
	"encoding/binary"
	"fmt"
	"strings"
)

// Company 公司
type Company struct {
	Name string // 名称
	Code string // 代码
}

// Marshal 序列化
func (c Company) Marshal() []byte {

	name := []byte(c.Name)
	nameLen := len(name)

	buffer := make([]byte, 19+nameLen)
	copy(buffer[:16], []byte(c.Code))
	binary.BigEndian.PutUint16(buffer[16:18], uint16(nameLen))
	copy(buffer[18:18+nameLen], name)
	buffer[18+nameLen] = 0

	return buffer
}

// Unmarshal 反序列化
func (c *Company) Unmarshal(buffer []byte) int {

	c.Code = strings.Trim(string(buffer[:16]), string(0x0))
	nameLen := int(binary.BigEndian.Uint16(buffer[16:18]))
	c.Name = strings.Trim(string(buffer[18:18+nameLen]), string(0x0))

	return 19 + nameLen
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
