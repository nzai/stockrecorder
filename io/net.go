package io

import (
	"io/ioutil"
	"net/http"
)

//	发送GET请求并返回结果字符串
func GetString(url string) (string, error) {
	return GetStringReferer(url, "")
}

//	发送GET请求并返回结果字符串(带Referer)
func GetStringReferer(url, referer string) (string, error) {
	
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	
	if referer != "" {
		request.Header.Set("Referer", referer)
	}
	
	client := &http.Client{}	
	response, err := client.Do(request)
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