package ioutil

import (
	"encoding/binary"
	"io"
	"math"
	"time"
)

// BinaryWriter 二进制writer
type BinaryWriter struct {
	io.Writer
	order binary.ByteOrder
}

// NewBinaryWriter 新建二进制writer
func NewBinaryWriter(w io.Writer) *BinaryWriter {
	return NewBinaryWriterOrder(w, binary.BigEndian)
}

// NewBinaryWriterOrder 新建二进制writer
func NewBinaryWriterOrder(w io.Writer, order binary.ByteOrder) *BinaryWriter {
	return &BinaryWriter{w, order}
}

// Bool 写入bool
func (w BinaryWriter) Bool(value bool) error {
	var v uint8
	if value {
		v = 1
	}

	return w.UInt8(v)
}

// UInt8 写入uint8
func (w BinaryWriter) UInt8(value uint8) error {
	_, err := w.Write([]byte{byte(value)})
	return err
}

// UInt16 写入uint16
func (w BinaryWriter) UInt16(value uint16) error {
	buffer := make([]byte, 2)
	w.order.PutUint16(buffer, value)

	_, err := w.Write([]byte{byte(value)})
	return err
}

// UInt32 写入uint32
func (w BinaryWriter) UInt32(value uint32) error {
	buffer := make([]byte, 4)
	w.order.PutUint32(buffer, value)

	_, err := w.Write([]byte{byte(value)})
	return err
}

// UInt64 写入uint64
func (w BinaryWriter) UInt64(value uint64) error {
	buffer := make([]byte, 8)
	w.order.PutUint64(buffer, value)

	_, err := w.Write([]byte{byte(value)})
	return err
}

// Int8 写入Int8
func (w BinaryWriter) Int8(value int8) error {
	return w.UInt8(uint8(value))
}

// Int16 写入Int16
func (w BinaryWriter) Int16(value int16) error {
	return w.UInt16(uint16(value))
}

// Int32 写入Int32
func (w BinaryWriter) Int32(value int32) error {
	return w.UInt32(uint32(value))
}

// Int64 写入Int64
func (w BinaryWriter) Int64(value int64) error {
	return w.UInt64(uint64(value))
}

// Int 写入Int
func (w BinaryWriter) Int(value int) error {
	return w.UInt32(uint32(value))
}

// Float32 写入float32
func (w BinaryWriter) Float32(value float32) error {
	return w.UInt32(math.Float32bits(value))
}

// Float64 写入float64
func (w BinaryWriter) Float64(value float64) error {
	return w.UInt64(math.Float64bits(value))
}

// String 写入字符串
func (w BinaryWriter) String(value string) error {
	buffer := []byte(value)
	bufferLength := len(buffer)
	err := w.UInt32(uint32(bufferLength))
	if err != nil {
		return err
	}

	_, err = w.Write(buffer)
	return err
}

// Time 写入Time
func (w BinaryWriter) Time(value time.Time) error {
	err := w.UInt64(uint64(value.Unix()))
	if err != nil {
		return err
	}

	return w.String(value.Location().String())
}
