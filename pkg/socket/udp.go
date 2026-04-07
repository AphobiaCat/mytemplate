package socket

import (
	"fmt"
	"mytemplate/pkg/log"
	"mytemplate/pkg/util"
	"net"
)

/*****************************************************************************
	For host manager to manager clients
*****************************************************************************/

type udpRouteManager struct {
	conn           *net.UDPConn
	newClientsChan chan *udpRouteClient
	errChan        chan error
}

type udpRouteClient struct {
	conn             *net.UDPConn
	addr             *net.UDPAddr
	recvMsgChan      chan string
	lastClientOpTime int64
}

func (o *udpRouteManager) Init(bindPort string) {

	addr, err := net.ResolveUDPAddr("udp", ":"+bindPort)
	if err != nil {
		log.DebugError("Error resolving address:", err)
		panic(err)
	}

	o.conn, err = net.ListenUDP("udp", addr)

	if err != nil {
		log.DebugError("Error listening:", err)
		panic(err)
	}

	log.Log("UDP server is listening on port ", bindPort, "...")

	buffer := make([]byte, 1024)
	clientsMap := make(map[string]*udpRouteClient)

	o.newClientsChan = make(chan *udpRouteClient, 10)
	o.errChan = make(chan error, 10)

	go func() {
		defer func() {
			o.conn.Close()
			err := fmt.Errorf("unexpect return")
			o.errChan <- err
		}()

		for {
			n, addr, err := o.conn.ReadFromUDP(buffer)

			if err != nil {
				log.DebugError("Error accepting connection:", err)
				continue
			}

			clientAddr := addr.String()

			client, exist := clientsMap[clientAddr]

			if !exist {
				newClient := &udpRouteClient{
					conn:        o.conn,
					addr:        addr,
					recvMsgChan: make(chan string, 10),
				}
				clientsMap[clientAddr] = newClient
				client = newClient
				o.newClientsChan <- newClient
			}

			client.recvMsgChan <- string(buffer[:n])
			client.lastClientOpTime = util.NowTimeS()
		}
	}()
}

func (o *udpRouteManager) NewClient() (client proxyClient, addr string, err error) {

	select {
	case newClient := <-o.newClientsChan:
		addr = newClient.addr.String()
		client = newClient

	case routeErr := <-o.errChan:
		err = routeErr
		log.DebugError("udpRouteManager get error ", err)
	}

	return
}

func (o *udpRouteClient) Close() {
	log.DebugLog("udp user[", o.addr.String(), "] close")
}

func (o *udpRouteClient) recvMsg() (msg string, err error) {
	msg = <-o.recvMsgChan
	return
}

func (o *udpRouteClient) SendMsg(msg string) (err error) {
	_, err = o.conn.WriteToUDP([]byte(msg), o.addr)

	return
}

/*****************************************************************************
	For User Client to connect to host
*****************************************************************************/

type udpUserClient struct {
	conn   *net.UDPConn
	buffer []byte
}

func (o *udpUserClient) Init(url string) (err error) {
	remoteAddr, err := net.ResolveUDPAddr("udp", url)
	if err != nil {
		err = fmt.Errorf("address parser failed: %s", err.Error())
		log.DebugError(err)
		return
	}

	// create udp socket
	o.conn, err = net.DialUDP("udp", nil, remoteAddr)
	if err != nil {
		err = fmt.Errorf("udp connect failed: %s", err.Error())
		log.DebugError(err)
		return
	}

	o.buffer = make([]byte, 1024)
	return
}

func (o *udpUserClient) Close() {
	o.conn.Close()
}

func (o *udpUserClient) RecvMsg() (msg string, err error) {
	n, err := o.conn.Read(o.buffer)
	if err != nil {
		return
	}
	msg = string(o.buffer[:n])
	return
}

func (o *udpUserClient) SendMsg(msg string) (err error) {
	_, err = o.conn.Write([]byte(msg))
	return
}
