package market

// Company 公司
type Company struct {
	Name string // 名称
	Code string // 代码
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
