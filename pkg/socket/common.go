package socket

const (
	SocketTypeUnknown int = iota
	SocketTypeTcp
	SocketTypeUdp
)

type proxyHost interface {
	Init(bindPort string)
	NewClient() (client proxyClient, addr string, err error)
}

type proxyClient interface {
	Close()
	RecvMsg() (msg string, err error)
	SendMsg(msg string) (err error)
}

type proxyUserClient interface {
	Init(url string) (err error)
	Close()
	RecvMsg() (msg string, err error)
	SendMsg(msg string) (err error)
}

type ClientReturn struct {
	Api     string `json:"api"`
	Header  string `json:"header"`
	Content string `json:"content"`
}

type routeCallbackType1 func(string) (*ClientReturn, error)
type routeCallbackType2 func(string)
type routeCallbackWrap func(string) (ret *ClientReturn, err error, needReturn bool)
