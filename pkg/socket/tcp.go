package socket

import (
	"mytemplate/pkg/log"
	"net"
)

type tcpRouteManager struct {
	newClientsChan chan *tcpRouteClient
	errChan        chan error
}

type tcpRouteClient struct {
	conn   net.Conn
	buffer []byte
	addr   string
}

func (o *tcpRouteManager) Init(bindPort string) {
	ln, err := net.Listen("tcp", ":"+bindPort)
	if err != nil {
		log.DebugError("Error listening:", err)
		panic(err)
	}
	defer ln.Close()

	log.Log("TCP server is listening on port ", bindPort, "...")

	o.newClientsChan = make(chan *tcpRouteClient, 10)
	o.errChan = make(chan error, 10)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.DebugError("Error accepting connection: ", err)
			continue
		}

		tmp_client := &tcpRouteClient{
			conn:   conn,
			buffer: make([]byte, 1024),
			addr:   conn.RemoteAddr().String(),
		}

		o.newClientsChan <- tmp_client
	}
}

func (o *tcpRouteManager) NewClient() (client proxyClient, addr string, err error) {
	select {
	case newClient := <-o.newClientsChan:
		addr = newClient.addr
		client = newClient

	case routeErr := <-o.errChan:
		err = routeErr
		log.DebugError("udpRouteManager get error ", err)
	}

	return
}

func (o *tcpRouteClient) Close() {
	log.DebugLog("udp user[", o.addr, "] close")
	o.conn.Close()
}

func (o *tcpRouteClient) RecvMsg() (msg string, err error) {
	_, err = o.conn.Read(o.buffer)
	msg = string(o.buffer)
	return
}

func (o *tcpRouteClient) SendMsg(msg string) (err error) {
	_, err = o.conn.Write([]byte(msg))

	return
}
