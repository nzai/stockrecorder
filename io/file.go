package io

import (
	"bufio"
	"io/ioutil"
	"os"
	"path/filepath"
)

//	写文件
func WriteLines(filePath string, lines []string) error {

	//	打开文件
	file, err := openForWrite(filePath)
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

//	写入字符串
func WriteString(filePath, content string) error {

	//	打开文件
	file, err := openForWrite(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)

	return err
}

//	写入缓冲区
func WriteBytes(filePath string, buffer []byte) error {
	//	打开文件
	file, err := openForWrite(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(buffer)

	return err
}

//	保证目录存在
func EnsureDir(dir string) error {
	if IsExists(dir) {
		return nil
	}
	
	//	递推
	err := EnsureDir(filepath.Dir(dir))
	if err != nil {
		return err
	}
	
	return os.Mkdir(dir, 0666)
}

//	打开文件
func openForWrite(filePath string) (*os.File, error) {

	//	保证文件所处目录是否存在
	err := EnsureDir(filepath.Dir(filePath))
	if err != nil {
		return nil, err
	}

	//	打开文件
	return os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0666)
}

//	打开文件
func openForRead(filePath string) (*os.File, error) {
	//	检查文件
	_, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	//	打开文件
	return os.OpenFile(filePath, os.O_RDONLY, 0666)
}

//	读取文件
func ReadLines(filePath string) ([]string, error) {
	//	打开文件
	file, err := openForRead(filePath)
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

//	读取所有
func ReadAllBytes(filePath string) ([]byte, error) {
	//	打开文件
	file, err := openForRead(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return ioutil.ReadAll(file)
}

//	读取所有
func ReadAllString(filePath string) (string, error) {
	buffer, err := ReadAllBytes(filePath)
	if err != nil {
		return "", err
	}

	return string(buffer), nil
}

//	判断文件是否存在
func IsExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}
