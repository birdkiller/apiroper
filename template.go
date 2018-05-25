package apiroper

// 资源模板
type Resource struct {
	Url     string                 // 获取资源所用的url
	Method  string                 // 获取资源所用的http函数
	Input   map[string]interface{} // 获取资源所用的数据
	Header  map[string]string      // 获取资源所用的包头
	Timeout int64                  // 该资源超时时间 毫秒数
}

// 模板
type Template struct {
	Output    map[string]interface{} // 输出内容
	Resources map[string]Resource    // 资源列表
}

const (
	DEFAULT_TIMEOUT_DURATION int64 = 5000 //默认超时时间毫秒数
)
