package example

import (
	"context"
	"fmt"

	"mytemplate/internal/global"
	"mytemplate/internal/types/example"
	"mytemplate/pkg/log"
	"mytemplate/pkg/route"
	"mytemplate/pkg/socket"
	"mytemplate/pkg/util"
)

func TestGetExample(ctx *context.Context, req *example.GetExampleRequest) (ret interface{}, err error) {

	ret = &example.GetExampleResponse{
		Message: "Hello " + util.BuildJson(req),
	}

	return
}

func TestMid(ctx *context.Context, req *example.MidRequest) (*example.MidResponse, error) {

	return &example.MidResponse{
		User: "mid:" + req.User,
		Msg:  "mid:" + req.Msg,
	}, nil
}

func TestPostExample(ctx *context.Context, req *example.PostExampleRequest) (ret interface{}, err error) {
	token, _ := route.RouteGenerateJwtByStr("dunty test", 1200)
	log.DebugLog("generate jwt token[", token, "]")

	ret = &example.PostExampleResponse{
		Message: "Hello " + util.BuildJson(req),
	}

	return
}

var socketHaveReturnApi = "example/socket"
var socketNotReturnApi = "example/socket/notreturn"

func TestSocketExample(msg string) (ret *socket.ClientReturn, err error) {
	ret = &socket.ClientReturn{
		Api:     socketNotReturnApi,
		Content: "haah " + msg,
	}

	log.DebugLog("TestSocketExample recv msg ", msg)

	return
}

func TestSocketExampleNoreturn(msg string) {
	log.DebugLog("TestSocketExampleNoreturn recv msg ", msg)
}

func setupSocketServer() {
	manager := socket.NewManager()

	tcpManager := manager.NewProxyHost(socket.SocketTypeTcp, "8877")
	tcpManager.Route(socketHaveReturnApi, TestSocketExample)
	tcpManager.Route(socketNotReturnApi, TestSocketExampleNoreturn)

	udpManager := manager.NewProxyHost(socket.SocketTypeUdp, "7788")
	udpManager.Route(socketHaveReturnApi, TestSocketExample)
	udpManager.Route(socketNotReturnApi, TestSocketExampleNoreturn)

	manager.InitAll()
}

func setupHttpServer() {
	// setup server routes

	routeManager := route.New()

	routeManager.RouteGet("example/get", TestGetExample).NeedUserIp()
	routeManager.RouteGet("example/get/mid", TestGetExample).NeedUserIp().Middle(TestMid)
	routeManager.RoutePost("example/post", TestPostExample)
	routeManager.RoutePost("example/post/mid", TestPostExample).NeedUserIp().Middle(TestMid)
	routeManager.RoutePost("example/post/jwt", TestPostExample).NeedUserIp().Middle(TestMid).Middle(route.RouteAuthJwtMid)

	bindAddr := fmt.Sprintf("%s:%d", global.AppConfig.Host, global.AppConfig.Port)
	log.DebugLog("Bind address: ", bindAddr)

	routeManager.InitRoute(bindAddr)
}

func setupSocketClient() {
	clientManager := socket.NewClient()
	// udpClient := clientManager.NewProxyClient(socket.SocketTypeUdp, "127.0.0.1:7788")
	// udpClient.ProcessCallback(socketHaveReturnApi, TestSocketExample)
	// udpClient.ProcessCallback(socketNotReturnApi, TestSocketExampleNoreturn)
	// udpClient.Init()
	// udpClient.SendMsg(socketHaveReturnApi, "", "hello")
	// udpClient.SendMsg(socketNotReturnApi, "", "hello")

	tcpClient := clientManager.NewProxyClient(socket.SocketTypeTcp, "127.0.0.1:8877")
	tcpClient.ProcessCallback(socketHaveReturnApi, TestSocketExample)
	tcpClient.ProcessCallback(socketNotReturnApi, TestSocketExampleNoreturn)
	tcpClient.Init()
	tcpClient.SendMsg(socketHaveReturnApi, "", "hello")
	tcpClient.SendMsg(socketNotReturnApi, "", "hello")

	util.Sleep(1000 * 10)

	// udpClient.Close()
	tcpClient.Close()
}

func Test() {
	go setupSocketServer()
	// go setupHttpServer()
	// go setupSocketClient()

	for {
		util.Sleep(10000)
	}

}
