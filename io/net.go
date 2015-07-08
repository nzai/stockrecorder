package io

import (
	"io/ioutil"
	"net/http"
)

//	发送GET请求并返回结果字符串
func GetString(url string) (string, error) {

	response, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	buffer, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return string(buffer), nil
}
