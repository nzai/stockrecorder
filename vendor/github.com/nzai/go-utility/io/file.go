package io

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// WriteLines 写文件
func WriteLines(filePath string, lines []string) error {

	//	打开文件
	file, err := OpenForWrite(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, line := range lines {

		//	将字符串写入文件
		_, err = file.WriteString(line + "\n")
		if err != nil {
			return err
		}
	}

	return nil
}

// WriteString 写入字符串
func WriteString(filePath, content string) error {

	//	打开文件
	file, err := OpenForWrite(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)

	return err
}

// WriteBytes 写入数据
func WriteBytes(filePath string, data []byte) error {
	//	打开文件
	file, err := OpenForWrite(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)

	return err
}

// WriteGzipBytes 压缩数据并写入文件
func WriteGzipBytes(filePath string, data []byte) error {

	// gzip 最高压缩
	buffer := new(bytes.Buffer)
	w, err := gzip.NewWriterLevel(buffer, gzip.BestCompression)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	if err != nil {
		return err
	}
	w.Flush()
	w.Close()

	zipped, err := ioutil.ReadAll(buffer)
	if err != nil {
		return err
	}

	// 存盘
	return WriteBytes(filePath, zipped)
}

// EnsureDir 保证目录存在
func EnsureDir(dir string) error {
	if IsExists(dir) {
		return nil
	}

	//	递推
	err := EnsureDir(filepath.Dir(dir))
	if err != nil {
		return err
	}

	return os.Mkdir(dir, 0755)
}

// OpenForWrite 打开文件以便写入
func OpenForWrite(filePath string) (*os.File, error) {

	//	保证文件所处目录是否存在
	err := EnsureDir(filepath.Dir(filePath))
	if err != nil {
		return nil, err
	}

	//	打开文件
	return os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0660)
}

// OpenForRead 打开文件一遍读取
func OpenForRead(filePath string) (*os.File, error) {
	//	检查文件
	_, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	//	打开文件
	return os.OpenFile(filePath, os.O_RDONLY, 0660)
}

// ReadLines 从文件中读取字符串数组
func ReadLines(filePath string) ([]string, error) {
	//	打开文件
	file, err := OpenForRead(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	//	读取
	scanner := bufio.NewScanner(file)
	lines := make([]string, 0)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}

// ReadAllBytes 从文件中读取数据
func ReadAllBytes(filePath string) ([]byte, error) {
	//	打开文件
	file, err := OpenForRead(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return ioutil.ReadAll(file)
}

// ReadAllGzipBytes 从文件中读取Gzip压缩所属
func ReadAllGzipBytes(filePath string) ([]byte, error) {

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader, err := gzip.NewReader(file)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return ioutil.ReadAll(reader)
}

// ReadAllString 从文件中读取字符串
func ReadAllString(filePath string) (string, error) {
	buffer, err := ReadAllBytes(filePath)
	if err != nil {
		return "", err
	}

	return string(buffer), nil
}

// IsExists 判断文件是否存在
func IsExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

var (
	// reservedFilenames 文件名保留字
	reservedFilenames = []string{"con", "prn", "aux", "nul", "com1", "com2", "com3", "com4", "com5", "com6", "com7", "com8", "com9", "lpt1", "lpt2", "lpt3", "lpt4", "lpt5", "lpt6", "lpt7", "lpt8", "lpt9"}
)

// IsReservedFileName 是否是文件名保留字
func IsReservedFileName(filePath string) bool {

	base := filepath.Base(filePath)
	parts := strings.Split(base, ".")
	var title string
	if len(parts) > 0 {
		title = strings.ToLower(parts[0])
	}

	for _, name := range reservedFilenames {
		if title != name {
			continue
		}

		return true
	}

	return false
}

// EnsureFileName 保证文件名不受保留字影响
func EnsureFileName(filePath string) string {

	isReserved := IsReservedFileName(filePath)
	if !isReserved {
		return filePath
	}

	dir := filepath.Dir(filePath)
	base := filepath.Base(filePath)
	ext := filepath.Ext(filePath)

	parts := strings.Split(base, ".")
	var title string
	if len(parts) > 0 {
		title = parts[0]
	}

	return filepath.Join(dir, "reserved."+title+ext)
}
