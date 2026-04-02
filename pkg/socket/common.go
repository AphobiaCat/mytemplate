package socket

const (
	SocketTypeUnknown int = iota
	SocketTypeTcp
	SocketTypeUdp
)

type proxyManager interface {
	Init(bindPort string)
	NewClient() (client proxyClient, addr string, err error)
}

type proxyClient interface {
	Close()
	RecvMsg() (msg string, err error)
	SendMsg(msg string) (err error)
}
