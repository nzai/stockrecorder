package ioutil

import (
	"encoding/binary"
	"io"
	"math"
	"time"
)

// BinaryReader 二进制writer
type BinaryReader struct {
	io.Reader
	order binary.ByteOrder
}

// NewBinaryReader 新建二进制reader
func NewBinaryReader(r io.Reader) *BinaryReader {
	return NewBinaryReaderOrder(r, binary.BigEndian)
}

// NewBinaryReaderOrder 新建二进制reader
func NewBinaryReaderOrder(r io.Reader, order binary.ByteOrder) *BinaryReader {
	return &BinaryReader{r, order}
}

// Bool 读取bool
func (r BinaryReader) Bool() (bool, error) {
	value, err := r.UInt8()
	if err != nil {
		return false, err
	}

	if value > 0 {
		return true, nil
	}

	return false, nil
}

// UInt8 读取uint8
func (r BinaryReader) UInt8() (uint8, error) {
	buffer := make([]byte, 1)
	_, err := r.Read(buffer)
	if err != nil {
		return 0, err
	}

	return uint8(buffer[0]), nil
}

// UInt16 读取uint16
func (r BinaryReader) UInt16() (uint16, error) {
	buffer := make([]byte, 2)
	_, err := r.Read(buffer)
	if err != nil {
		return 0, err
	}

	return r.order.Uint16(buffer), nil
}

// UInt32 读取uint32
func (r BinaryReader) UInt32() (uint32, error) {
	buffer := make([]byte, 4)
	_, err := r.Read(buffer)
	if err != nil {
		return 0, err
	}

	return r.order.Uint32(buffer), nil
}

// UInt64 读取uint64
func (r BinaryReader) UInt64() (uint64, error) {
	buffer := make([]byte, 8)
	_, err := r.Read(buffer)
	if err != nil {
		return 0, err
	}

	return r.order.Uint64(buffer), nil
}

// Int8 读取int8
func (r BinaryReader) Int8() (int8, error) {
	value, err := r.UInt8()
	return int8(value), err
}

// Int16 读取int16
func (r BinaryReader) Int16() (int16, error) {
	value, err := r.UInt16()
	return int16(value), err
}

// Int32 读取int32
func (r BinaryReader) Int32() (int32, error) {
	value, err := r.UInt32()
	return int32(value), err
}

// Int64 读取int64
func (r BinaryReader) Int64() (int64, error) {
	value, err := r.UInt64()
	return int64(value), err
}

// Int 读取int
func (r BinaryReader) Int() (int, error) {
	value, err := r.UInt32()
	return int(value), err
}

// Float32 读取float32
func (r BinaryReader) Float32() (float32, error) {
	value, err := r.UInt32()
	if err != nil {
		return 0, nil
	}

	return math.Float32frombits(value), nil
}

// Float64 读取float64
func (r BinaryReader) Float64() (float64, error) {
	value, err := r.UInt64()
	if err != nil {
		return 0, nil
	}

	return math.Float64frombits(value), nil
}

// String 读取字符串
func (r BinaryReader) String() (string, error) {
	size, err := r.UInt32()
	if err != nil {
		return "", err
	}

	buffer := make([]byte, int(size))
	_, err = r.Read(buffer)
	if err != nil {
		return "", err
	}

	return string(buffer), nil
}

// Time 读取时间
func (r BinaryReader) Time() (time.Time, error) {
	timestamp, err := r.UInt64()
	if err != nil {
		return time.Time{}, err
	}

	locationName, err := r.String()
	if err != nil {
		return time.Time{}, err
	}

	location, err := time.LoadLocation(locationName)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(int64(timestamp), 0).In(location), nil
}
