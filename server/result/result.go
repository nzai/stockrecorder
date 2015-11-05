package result

//	通用结果
type httpResult struct {
	Success bool
	Message string
	Data    interface{}
}

func Create(data interface{}) httpResult {
	return httpResult{Success: data != nil, Data: data}
}

func Failed(message string) httpResult {
	return httpResult{Success: false, Message: message}
}
