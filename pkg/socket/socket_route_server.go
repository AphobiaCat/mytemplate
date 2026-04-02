package socket

import (
	"fmt"
	"mytemplate/pkg/log"
)

/*
	Msg Struct string
	Api|Header|Body
*/

type routeCallbackType1 func(string) (string, error)
type routeCallbackType2 func(string)
type routeCallbackWrap func(string) (ret string, err error, needReturn bool)

type hostManager struct {
	proxy    proxyManager
	bindPort string
	routes   map[string]routeCallbackWrap
}

type socketRouteManager struct {
	hostManagers  []*hostManager
	managersIndex int
}

func (srm *socketRouteManager) NewProxyManager(socketType int, bindPort string) *hostManager {
	ret := &hostManager{
		bindPort: bindPort,
		routes:   make(map[string]routeCallbackWrap),
	}

	var proxy proxyManager

	switch socketType {
	case SocketTypeTcp:
		proxy = &tcpRouteManager{}

	case SocketTypeUdp:
		proxy = &udpRouteManager{}

	default:
		panic(fmt.Errorf("no support socket type[%d]", socketType))
	}

	ret.proxy = proxy
	srm.hostManagers = append(srm.hostManagers, ret)
	srm.managersIndex += 1

	return ret
}

func (srm *socketRouteManager) InitAll() {
	for _, proxyManager := range srm.hostManagers {
		go func() {
			proxy := proxyManager.proxy

			clientsMap := make(map[string]proxyClient)
			proxy.Init(proxyManager.bindPort)

			for {
				client, clientAddr, err := proxy.NewClient()

				if err != nil {
					log.DebugError("Error accepting connection:", err)
					continue
				}

				if _, exist := clientsMap[clientAddr]; !exist {
					clientsMap[clientAddr] = client
				}
			}
		}()
	}
}

func (o *hostManager) Route(api string, handler interface{}) {
	var wrapFunc routeCallbackWrap

	switch handlerFunc := handler.(type) {
	case func(string) (string, error):
		wrapFunc = func(input string) (string, error, bool) {
			ret, err := handlerFunc(input)
			return ret, err, true
		}

	case func(string):
		wrapFunc = func(input string) (string, error, bool) {
			handlerFunc(input)
			return "", nil, false
		}
	default:
		log.DebugError("handler[", handlerFunc, "] func type no support")
	}

	o.routes[api] = wrapFunc

	log.Log("Socket--> ", api)
}

func commonProcess(client proxyClient, handlers map[string]routeCallbackWrap) {
	defer client.Close()
	for {
		msg, err := client.RecvMsg()
		if err != nil {
			log.DebugError("client err ", err)
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

		if len(header) == 0 || len(content) == 0 || msgLenErr {
			log.DebugError("header or content no exist")
			continue
		}

		// wait to use
		_ = header

		ret, err, needReturn := handlers[api](content)
		if err != nil {
			log.DebugError("client err ", err)
			continue
		}

		if needReturn {
			err = client.SendMsg(ret)
			if err != nil {
				log.DebugError("client err ", err)
				return
			}
		}
	}
}

func NewManager() *socketRouteManager {
	ret := &socketRouteManager{}
	return ret
}
