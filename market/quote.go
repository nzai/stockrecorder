package market

import (
	"encoding/binary"
	"fmt"
	"log"
	"strings"
	"time"
)

// DailyQuote 市场每日报价
type DailyQuote struct {
	Market
	UTCOffset int
	Date      time.Time
	Quotes    []CompanyDailyQuote
}

// Marshal 序列化
func (q DailyQuote) Marshal() []byte {

	count := uint32(len(q.Quotes))
	buffer := make([]byte, 12+count*4)
	binary.BigEndian.PutUint32(buffer[:4], uint32(q.UTCOffset+43200))
	binary.BigEndian.PutUint32(buffer[4:8], uint32(q.Date.Unix()))
	binary.BigEndian.PutUint32(buffer[8:12], count)

	for index, quote := range q.Quotes {
		binary.BigEndian.PutUint32(buffer[12+index*4:16+index*4], uint32(len(buffer)))
		buffer = append(buffer, quote.Marshal()...)
	}

	return buffer
}

// Unmarshal 反序列化
func (q *DailyQuote) Unmarshal(buffer []byte) {

	q.UTCOffset = int(binary.BigEndian.Uint32(buffer[:4])) - 43200
	q.Date = time.Unix(int64(binary.BigEndian.Uint32(buffer[4:8])+000), 0)
	count := binary.BigEndian.Uint32(buffer[8:12])

	for index := 0; index < int(count); index++ {

		offset := binary.BigEndian.Uint32(buffer[12+index*4 : 16+index*4])
		quote := CompanyDailyQuote{}
		quote.Unmarshal(buffer[offset:])

		q.Quotes = append(q.Quotes, quote)
	}
}

// Equal 判断是否相等
func (q DailyQuote) Equal(s DailyQuote) error {

	if q.UTCOffset != s.UTCOffset {
		return fmt.Errorf("DailyQuote UTCOffset不相等:q.UTCOffset=[%d] s.UTCOffset=[%d]", q.UTCOffset, s.UTCOffset)
	}

	if q.Date.Unix() != s.Date.Unix() {
		return fmt.Errorf("DailyQuote Date不相等:q.Date=[%s] s.Date=[%s]", q.Date.Format("2006-01-02 15:04:05"), s.Date.Format("2006-01-02 15:04:05"))
	}

	for index, quote := range q.Quotes {

		err := quote.Equal(s.Quotes[index])
		if err == nil {
			continue
		}

		return fmt.Errorf("DailyQuote Quotes不相等:index=%d  %v", index, err)
	}

	return nil
}

// CompanyDailyQuote 公司每日报价
type CompanyDailyQuote struct {
	Code    string
	Name    string
	Pre     QuoteSeries
	Regular QuoteSeries
	Post    QuoteSeries
}

// Marshal 序列化
func (q CompanyDailyQuote) Marshal() []byte {
	name := []byte(q.Name)

	buffer := make([]byte, 19+len(name))
	copy(buffer[:16], []byte(q.Code))
	binary.BigEndian.PutUint16(buffer[16:18], uint16(len(name)))
	copy(buffer[18:18+len(name)], name)
	buffer[18+len(name)] = 0

	buffer = append(buffer, q.Pre.Marshal()...)
	buffer = append(buffer, q.Regular.Marshal()...)
	buffer = append(buffer, q.Post.Marshal()...)

	return buffer
}

// Unmarshal 反序列化
func (q *CompanyDailyQuote) Unmarshal(buffer []byte) {
	q.Code = strings.Trim(string(buffer[:16]), string(0x0))
	nameLen := binary.BigEndian.Uint16(buffer[16:18])
	q.Name = strings.Trim(string(buffer[18:18+nameLen]), string(0x0))

	quotePos := int(19 + nameLen)
	q.Pre.Unmarshal(buffer[quotePos:])
	q.Regular.Unmarshal(buffer[quotePos+q.Pre.Len():])
	q.Post.Unmarshal(buffer[quotePos+q.Pre.Len()+q.Regular.Len():])
}

// Equal 判断是否相等
func (q CompanyDailyQuote) Equal(s CompanyDailyQuote) error {
	if q.Code != s.Code {
		return fmt.Errorf("CompanyDailyQuote Code不相等:q.Code=[%s] s.Code=[%s]", q.Code, s.Code)
	}

	if q.Name != s.Name {
		return fmt.Errorf("CompanyDailyQuote Name不相等:q.Name=[%s] s.Name=[%s]", q.Name, s.Name)
	}

	err := q.Pre.Equal(s.Pre)
	if err != nil {
		return fmt.Errorf("CompanyDailyQuote Pre不相等:%v", err)
	}

	err = q.Regular.Equal(s.Regular)
	if err != nil {
		return fmt.Errorf("CompanyDailyQuote Regular不相等:%v", err)
	}

	err = q.Post.Equal(s.Post)
	if err != nil {
		return fmt.Errorf("CompanyDailyQuote Post不相等:%v", err)
	}

	return nil
}

// Glance 显示摘要
func (q CompanyDailyQuote) Glance(logger *log.Logger, location *time.Location) {

	logger.Printf("上市公司:%s\t%s", q.Code, q.Name)
	q.Pre.Glance(logger, "Pre", location)
	q.Regular.Glance(logger, "Regular", location)
	q.Post.Glance(logger, "Post", location)
	logger.Println("")
}

// QuoteSeries 报价序列
type QuoteSeries struct {
	Count     uint32
	Timestamp []uint32
	Open      []uint32
	Close     []uint32
	Max       []uint32
	Min       []uint32
	Volume    []uint32
}

// Marshal 序列化
func (s QuoteSeries) Marshal() []byte {
	buffer := make([]byte, s.Len())

	binary.BigEndian.PutUint32(buffer[:4], s.Count)

	var values []uint32
	values = append(values, s.Timestamp...)
	values = append(values, s.Open...)
	values = append(values, s.Close...)
	values = append(values, s.Max...)
	values = append(values, s.Min...)
	values = append(values, s.Volume...)

	for index, value := range values {
		binary.BigEndian.PutUint32(buffer[4+index*4:8+index*4], value)
	}

	if int(s.Count)*6+4 > len(buffer) {
		panic(fmt.Errorf("s.Count:%d   len(buffer):%d   Len:%d", s.Count, len(buffer), s.Len()))
	}

	return buffer
}

// Unmarshal 反序列化
func (s *QuoteSeries) Unmarshal(data []byte) {

	s.Count = binary.BigEndian.Uint32(data[:4])

	if int(s.Count)*6+4 > len(data) {
		panic(fmt.Errorf("s.Count:%d   len(data):%d   Len:%d", s.Count, len(data), s.Len()))
	}

	valueCount := int(s.Count * 6)
	values := make([]uint32, valueCount)
	for index := 0; index < valueCount; index++ {
		values[index] = binary.BigEndian.Uint32(data[4+index*4 : 8+index*4])
	}

	s.Timestamp = values[:s.Count]
	s.Open = values[s.Count : s.Count*2]
	s.Close = values[s.Count*2 : s.Count*3]
	s.Max = values[s.Count*3 : s.Count*4]
	s.Min = values[s.Count*4 : s.Count*5]
	s.Volume = values[s.Count*5 : s.Count*6]
}

// Len 长度
func (s QuoteSeries) Len() int {
	return int(s.Count)*4*6 + 4
}

// Equal 是否相同
func (s QuoteSeries) Equal(q QuoteSeries) error {
	if s.Count != q.Count {
		return fmt.Errorf("QuoteSeries Count不相等:s.Count=%d q.Count=%d", s.Count, q.Count)
	}

	if len(q.Open) != int(q.Count) {
		return fmt.Errorf("QuoteSeries Count不相等:len(q.Open)=%d int(q.Count)=%d", len(q.Open), int(q.Count))
	}

	if len(q.Close) != int(q.Count) {
		return fmt.Errorf("QuoteSeries Count不相等:len(q.Close)=%d int(q.Count)=%d", len(q.Close), int(q.Count))
	}

	if len(q.Max) != int(q.Count) {
		return fmt.Errorf("QuoteSeries Count不相等:len(q.Max)=%d int(q.Count)=%d", len(q.Max), int(q.Count))
	}

	if len(q.Min) != int(q.Count) {
		return fmt.Errorf("QuoteSeries Count不相等:len(q.Min)=%d int(q.Count)=%d", len(q.Min), int(q.Count))
	}

	if len(q.Volume) != int(q.Count) {
		return fmt.Errorf("QuoteSeries Count不相等:len(q.Volume)=%d int(q.Count)=%d", len(q.Volume), int(q.Count))
	}

	err := s.arrayEqual(s.Timestamp, q.Timestamp)
	if err != nil {
		return fmt.Errorf("QuoteSeries Timestamp不相等:%v", err)
	}

	err = s.arrayEqual(s.Open, q.Open)
	if err != nil {
		return fmt.Errorf("QuoteSeries Open不相等:%v", err)
	}

	err = s.arrayEqual(s.Close, q.Close)
	if err != nil {
		return fmt.Errorf("QuoteSeries Close不相等:%v", err)
	}

	err = s.arrayEqual(s.Max, q.Max)
	if err != nil {
		return fmt.Errorf("QuoteSeries Max不相等:%v", err)
	}

	err = s.arrayEqual(s.Min, q.Min)
	if err != nil {
		return fmt.Errorf("QuoteSeries Min不相等:%v", err)
	}

	err = s.arrayEqual(s.Volume, q.Volume)
	if err != nil {
		return fmt.Errorf("QuoteSeries Volume不相等:%v", err)
	}

	return nil
}

// arrayEqual 数组是否相同
func (s QuoteSeries) arrayEqual(a []uint32, b []uint32) error {
	if len(a) != len(b) {
		return fmt.Errorf("数组长度不相等:%d %d", len(a), len(b))
	}

	for index, value := range a {
		if value != b[index] {
			return fmt.Errorf("数组值不相等:[%d] %d %d", index, value, b[index])
		}
	}

	return nil
}

// Glance 显示摘要
func (s QuoteSeries) Glance(logger *log.Logger, title string, location *time.Location) {
	count := 5
	if s.Count < 5 {
		count = int(s.Count)
	}

	logger.Printf("%s Count: %d", title, s.Count)
	for index := 0; index < count; index++ {
		logger.Printf("%s FIRST [%d]: time:%s\topen:%.2f\tclose:%.2f\tmax:%.2f\tmin:%.2f\tvolume:%d",
			title,
			index,
			time.Unix(int64(s.Timestamp[index]), 0).In(location).Format("2006-01-02 15:04:05"),
			float32(s.Open[index])/100,
			float32(s.Close[index])/100,
			float32(s.Max[index])/100,
			float32(s.Min[index])/100,
			s.Volume[index],
		)
	}

	for index := int(s.Count) - count; index < int(s.Count); index++ {
		logger.Printf("%s LAST [%d]: time:%s\topen:%.2f\tclose:%.2f\tmax:%.2f\tmin:%.2f\tvolume:%d",
			title,
			index,
			time.Unix(int64(s.Timestamp[index]), 0).In(location).Format("2006-01-02 15:04:05"),
			float32(s.Open[index])/100,
			float32(s.Close[index])/100,
			float32(s.Max[index])/100,
			float32(s.Min[index])/100,
			s.Volume[index],
		)
	}
}
