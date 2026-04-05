package socket

import (
	"fmt"
	"mytemplate/pkg/log"
	"sync"
)

/*
	Msg Struct string
	Api|Header|Body
*/

type hostManager struct {
	proxy    proxyHost
	bindPort string
	routes   map[string]routeCallbackWrap
}

type socketRouteManager struct {
	hostManagers  []*hostManager
	managersIndex int
}

func (srm *socketRouteManager) NewProxyHost(socketType int, bindPort string) *hostManager {
	ret := &hostManager{
		bindPort: bindPort,
		routes:   make(map[string]routeCallbackWrap),
	}

	var proxy proxyHost

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
	for _, proxyHost := range srm.hostManagers {
		go func() {
			proxy := proxyHost.proxy

			clientsMap := make(map[string]proxyClient)
			var mapMutex sync.Mutex

			proxy.Init(proxyHost.bindPort)

			for {
				client, clientAddr, err := proxy.NewClient()

				if err != nil {
					log.DebugError("Error accepting connection:", err)
					continue
				}

				mapMutex.Lock()
				_, exist := clientsMap[clientAddr]

				if !exist {
					clientsMap[clientAddr] = client
					go commonProcess(client, proxyHost.routes, clientsMap, clientAddr, &mapMutex)
				}
				mapMutex.Unlock()
			}
		}()
	}
}

func (o *hostManager) Route(api string, handler interface{}) {
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

	o.routes[api] = wrapFunc

	log.Log("Socket--> ", api)
}

func commonProcess(client proxyClient, handlers map[string]routeCallbackWrap, clientMaps map[string]proxyClient, clientAddr string, mapLock *sync.Mutex) {
	defer func() {
		client.Close()
		mapLock.Lock()
		defer mapLock.Unlock()

		delete(clientMaps, clientAddr)
	}()
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

		if len(content) == 0 || msgLenErr {
			log.DebugError("content no exist")
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
			err = client.SendMsg(ret.Api + "|" + ret.Header + "|" + ret.Content)
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
