package io

import (
	"os"
	"path/filepath"
)

//	写文件
func WriteFile(filePath string, lines []string) error {

	//	检查文件所处目录是否存在
	fileDir := filepath.Dir(filePath)
	_, err := os.Stat(fileDir)
	if os.IsNotExist(err) {
		//	如果不存在就先创建目录
		err = os.Mkdir(fileDir, 0x777)
		if err != nil {
			return err
		}
	}

	//	打开文件
	file, err := os.OpenFile(filePath, os.O_CREATE, 0x777)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, line := range lines {

		//	将股价写入文件
		_, err = file.WriteString(line)
		if err != nil {
			return err
		}
	}

	return nil
}
