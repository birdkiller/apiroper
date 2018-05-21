package apiroper

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

// 发送post请求
func post(url string, data []byte, header map[string]string) (code int, body []byte, err error) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return 500, nil, err
	}
	for k, v := range header {
		req.Header.Set(k, v)
	}

	rsp, err := client.Do(req)
	if rsp != nil {
		defer rsp.Body.Close()
	}
	if err != nil {
		return 500, nil, err
	}
	code = rsp.StatusCode
	body, err = ioutil.ReadAll(rsp.Body)
	return

}
