package curl

import (
	"context"
	"net/url"
	"time"
)

type CurlManager struct {
	defaultHeaders         map[string]string
	defaultHeadersHaveInit bool
	reqTimeout             time.Duration
}

func NewCurlManager() (ret *CurlManager) {

	ret = &CurlManager{
		defaultHeaders:         make(map[string]string),
		defaultHeadersHaveInit: false,
		reqTimeout:             30 * time.Second, // default timeout of 30 seconds
	}

	header := make(map[string]string)
	header["Accept"] = "*/*"
	header["Content-Type"] = "application/json"

	ret.SetDefaultHeaders(header)

	return
}

func (cm *CurlManager) Post(ctx *context.Context, baseUrl string) *CurlClient {
	ret := &CurlClient{
		ctx:     ctx,
		url:     baseUrl,
		timeout: cm.reqTimeout,
		isGet:   false,
		params:  url.Values{},
		headers: cm.defaultHeaders,
	}

	return ret
}

func (cm *CurlManager) Get(ctx *context.Context, baseUrl string) *CurlClient {
	ret := &CurlClient{
		ctx:     ctx,
		url:     baseUrl,
		timeout: cm.reqTimeout,
		isGet:   true,
		params:  url.Values{},
		headers: cm.defaultHeaders,
	}
	return ret
}

func (cm *CurlManager) SetDefaultHeaders(headerMap map[string]string, keepOld ...bool) (oldConfig map[string]string) {
	if !cm.defaultHeadersHaveInit {
		cm.defaultHeaders = make(map[string]string)
	}

	//clear(defaultHeaders) go 1.21 or later

	oldConfig = cm.defaultHeaders

	if len(keepOld) == 0 {
		for key, _ := range cm.defaultHeaders {
			delete(cm.defaultHeaders, key)
		}
	}

	for key, val := range headerMap {
		cm.defaultHeaders[key] = val
	}

	return oldConfig
}

func (cm *CurlManager) SetDefaultTimeout(timeoutSec int64) {
	cm.reqTimeout = time.Duration(timeoutSec * int64(time.Second))
}
