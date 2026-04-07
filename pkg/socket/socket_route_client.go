package socket

import (
	"fmt"
	"mytemplate/pkg/log"
	"mytemplate/pkg/util"
)

type clientManager struct {
	proxy       proxyUserClient
	url         string
	callbacks   map[string]routeCallbackWrap
	close       bool
	finishInit  bool
	waitSendMsg []ClientReturn
}

type socketRouteClientManager struct {
}

func (srcm *socketRouteClientManager) NewProxyClient(socketType int, url string) *clientManager {
	ret := &clientManager{
		url:       url,
		callbacks: make(map[string]routeCallbackWrap),
	}

	var proxy proxyUserClient

	switch socketType {
	case SocketTypeTcp:
		proxy = &tcpUserClient{}

	case SocketTypeUdp:
		proxy = &udpUserClient{}

	default:
		panic(fmt.Errorf("no support socket type[%d]", socketType))
	}

	ret.proxy = proxy

	return ret
}

func (o *clientManager) ProcessCallback(api string, handler interface{}) {
	var wrapFunc routeCallbackWrap

	switch handlerFunc := handler.(type) {
	case func(string) (*ClientReturn, error):
		wrapFunc = func(input string) (ret *ClientReturn, err error, needReturn bool) {
			ret, err = handlerFunc(input)
			return ret, err, true
		}

	case func(string):
		wrapFunc = func(input string) (ret *ClientReturn, err error, needReturn bool) {
			handlerFunc(input)
			return nil, nil, false
		}
	default:
		log.DebugError("handler[", handlerFunc, "] func type no support")
	}

	o.callbacks[api] = wrapFunc

	log.Log("Socket--> ", api)
}

func (o *clientManager) Init() {
	err := o.proxy.Init(o.url)
	if err != nil {
		if !o.close {
			log.DebugLog("wait 1000 ms retry connect")
			util.Sleep(1000)
			go o.Init()
		}
	} else {
		o.finishInit = true
		go o.recvMsg()

		// resend last msg
		tempArray := o.waitSendMsg
		o.waitSendMsg = []ClientReturn{}
		for _, msg := range tempArray {
			o.SendMsg(msg.Api, msg.Header, msg.Content)
		}
	}
}

func (o *clientManager) closeProxy() {
	o.proxy.Close()
}

func (o *clientManager) Close() {
	o.close = true
	if !o.finishInit {
		return
	}
	o.closeProxy()
}

func (o *clientManager) SendMsg(api, header, content string) (err error) {
	if !o.finishInit {
		o.waitSendMsg = append(o.waitSendMsg, ClientReturn{Api: api, Header: header, Content: content})
		return
	}
	err = o.proxy.SendMsg(api + "|" + header + "|" + content)
	if err != nil {
		o.waitSendMsg = append(o.waitSendMsg, ClientReturn{Api: api, Header: header, Content: content})
	}
	return
}

func (o *clientManager) recvMsg() {
	defer func() {
		o.finishInit = false
		if o.close {
			return
		}
		o.closeProxy()
		log.DebugLog("wait 1000 ms retry connect")
		util.Sleep(1000)
		o.Init()
	}()
	for {
		msg, err := o.proxy.RecvMsg()
		if err != nil {
			log.DebugError("recv msg error :", err)
			return
		}

		var api, header, content string
		var msgLenErr bool = false

		for index, char := range msg {
			if char == '|' {
				api = msg[:index]

				if index+1 > len(msg) {
					msgLenErr = true
					break
				}

				msg = msg[(index + 1):]
				break
			}
		}

		if len(api) == 0 || msgLenErr {
			log.DebugError("api no exist")
			continue
		}

		for index, char := range msg {
			if char == '|' {
				header = msg[:index]

				if index+1 > len(msg) {
					msgLenErr = true
					break
				}

				content = msg[(index + 1):]
				break
			}
		}

		if len(content) == 0 || msgLenErr {
			log.DebugError("content no exist")
			continue
		}

		// wait to use
		_ = header

		ret, err, needReturn := o.callbacks[api](content)
		if err != nil {
			log.DebugError("client err ", err)
			continue
		}

		if needReturn && ret != nil {
			err = o.SendMsg(ret.Api, ret.Header, ret.Content)
			if err != nil {
				log.DebugError("client err ", err)
				return
			}
		}
	}
}

func NewClient() *socketRouteClientManager {
	ret := &socketRouteClientManager{}
	return ret
}
