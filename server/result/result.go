package result

//	通用结果
type HttpResult struct {
	Success bool
	Message string
	Data    interface{}
}

func Create(data interface{}) HttpResult {
	return HttpResult{Success: data != nil, Data: data}
}

func Failed(message string) HttpResult {
	return HttpResult{Success: false, Message: message}
}
