package config

import (
	"os"
	"path/filepath"
)

const (
	dataDirName = "data"
)

var rootDir string
var dataDir string

//	设置配置文件
func SetRootDir(root string) error {

	rootDir = root
	dataDir = filepath.Join(rootDir, dataDirName)
	_, err := os.Stat(dataDir)
	if os.IsNotExist(err) {
		err = os.Mkdir(dataDir, 0x644)
		if err != nil {
			return err
		}
	}

	return nil
}

//	获取数据保存目录
func GetDataDir() string {

	return dataDir
}
