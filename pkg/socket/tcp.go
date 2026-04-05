package socket

import (
	"encoding/binary"
	"fmt"
	"io"
	"mytemplate/pkg/log"
	"net"
)

/*****************************************************************************
	For host manager to manager clients
*****************************************************************************/

type tcpRouteManager struct {
	ln net.Listener
}

type tcpRouteClient struct {
	conn       net.Conn
	sizeBuffer []byte
	buffer     []byte
	addr       string
}

func (o *tcpRouteManager) Init(bindPort string) {
	ln, err := net.Listen("tcp", ":"+bindPort)
	if err != nil {
		log.DebugError("Error listening:", err)
		panic(err)
	}
	o.ln = ln

	log.Log("TCP server is listening on port ", bindPort, "...")
}

func (o *tcpRouteManager) NewClient() (client proxyClient, addr string, err error) {
	conn, err := o.ln.Accept()

	if err != nil {
		return
	}

	client = &tcpRouteClient{
		conn:       conn,
		sizeBuffer: make([]byte, 4),
		buffer:     make([]byte, 1024),
		addr:       conn.RemoteAddr().String(),
	}

	return
}

func (o *tcpRouteClient) Close() {
	log.DebugLog("udp user[", o.addr, "] close")
	o.conn.Close()
}

func (o *tcpRouteClient) RecvMsg() (msg string, err error) {

	_, err = io.ReadFull(o.conn, o.sizeBuffer)
	if err != nil {
		return
	}

	msgLen := int(binary.BigEndian.Uint32(o.sizeBuffer))

	if cap(o.buffer) < msgLen {
		o.buffer = make([]byte, msgLen)
	}

	data := o.buffer[:msgLen]

	_, err = io.ReadFull(o.conn, data)
	if err != nil {
		return
	}

	msg = string(data)

	return

}

func writeFull(conn net.Conn, data []byte) error {
	total := 0
	for total < len(data) {
		n, err := conn.Write(data[total:])
		if err != nil {
			return err
		}
		total += n
	}
	return nil
}

func (o *tcpRouteClient) SendMsg(msg string) (err error) {
	data := []byte(msg)

	// send size frist
	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(data)))

	if err = writeFull(o.conn, lenBuf); err != nil {
		return
	}

	err = writeFull(o.conn, data)

	return
}

/*****************************************************************************
	For User Client to connect to host
*****************************************************************************/

type tcpUserClient struct {
	conn       net.Conn
	sizeBuffer []byte
	buffer     []byte
}

func (o *tcpUserClient) Init(url string) (err error) {
	o.conn, err = net.Dial("tcp", url)
	if err != nil {
		err = fmt.Errorf("tcp connect failed: %s", err.Error())
	}

	o.sizeBuffer = make([]byte, 4)
	o.buffer = make([]byte, 1024)

	return
}

func (o *tcpUserClient) Close() {
	o.conn.Close()
}

func (o *tcpUserClient) RecvMsg() (msg string, err error) {
	_, err = io.ReadFull(o.conn, o.sizeBuffer)
	if err != nil {
		return
	}

	msgLen := int(binary.BigEndian.Uint32(o.sizeBuffer))

	if cap(o.buffer) < msgLen {
		o.buffer = make([]byte, msgLen)
	}

	data := o.buffer[:msgLen]

	_, err = io.ReadFull(o.conn, data)
	if err != nil {
		return
	}

	msg = string(data)

	return
}

func (o *tcpUserClient) SendMsg(msg string) (err error) {

	data := []byte(msg)

	// send size frist
	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(data)))

	if err = writeFull(o.conn, lenBuf); err != nil {
		return
	}

	err = writeFull(o.conn, data)
	return
}
