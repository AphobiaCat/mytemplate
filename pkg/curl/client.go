package curl

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"mytemplate/pkg/log"
	"time"
)

type CurlClient struct {
	ctx     *context.Context
	url     string
	headers map[string]string
	params  url.Values
	bodyS   interface{}
	timeout time.Duration
	isGet   bool
}

func (cc *CurlClient) Header(key string, val string) *CurlClient {
	if cc.headers == nil {
		cc.headers = make(map[string]string)
	}

	cc.headers[key] = val

	return cc
}

func (cc *CurlClient) Body(val interface{}) *CurlClient {
	cc.bodyS = val

	return cc
}

func (cc *CurlClient) Param(key string, val string) *CurlClient {
	cc.params.Add(key, val)

	return cc
}

func (cc *CurlClient) Timeout(timeout_sec int) *CurlClient {
	cc.timeout = time.Duration(timeout_sec * 1000 * 1000 * 1000)

	return cc
}

func (cc *CurlClient) Send() (ret string, err error) {

	var req *http.Request

	if cc.isGet {
		req, err = http.NewRequest("GET", cc.url+"?"+cc.params.Encode(), nil)
		req.WithContext(*cc.ctx)
		if err != nil {
			log.DebugError("Error creating request:", err)
			return
		}
	} else {
		var jsonData []byte
		jsonData, err = json.Marshal(cc.bodyS)
		if err != nil {
			log.DebugError("Error marshalling JSON:", err)
		}

		req, err = http.NewRequest("POST", cc.url, bytes.NewBuffer(jsonData))
		req.WithContext(*cc.ctx)
		if err != nil {
			log.DebugError("Error creating request:", err)
			return
		}
	}

	for key, val := range cc.headers {
		req.Header.Set(key, val)
	}

	client := &http.Client{
		Timeout: cc.timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.DebugError("Error making request:", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.DebugError("Error reading response:", err)
		return
	}

	ret = string(body)

	return
}
