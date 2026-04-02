package handler

import (
	"fmt"
	"mytemplate/internal/global"
	"mytemplate/internal/handler/example"
	"mytemplate/pkg/log"
	"mytemplate/pkg/route"
	"mytemplate/pkg/socket"
)

func SetUpSocket() {
	manager := socket.NewManager()

	tcpManager := manager.NewProxyManager(socket.SocketTypeTcp, "8877")
	udpManager := manager.NewProxyManager(socket.SocketTypeUdp, "7788")

	tcpManager.Route("example/tcp", example.TestSocketExample)
	tcpManager.Route("example/tcp/noreturn", example.TestSocketExampleNoreturn)

	udpManager.Route("example/udp", example.TestSocketExample)
	udpManager.Route("example/udp/noreturn", example.TestSocketExampleNoreturn)

	manager.InitAll()
}

func Setup() {
	// setup server routes

	SetUpSocket()

	routeManager := route.New()

	routeManager.RouteGet("example/get", example.TestGetExample).NeedUserIp()
	routeManager.RouteGet("example/get/mid", example.TestGetExample).NeedUserIp().Middle(example.TestMid)
	routeManager.RoutePost("example/post", example.TestPostExample)
	routeManager.RoutePost("example/post/mid", example.TestPostExample).NeedUserIp().Middle(example.TestMid)
	routeManager.RoutePost("example/post/jwt", example.TestPostExample).NeedUserIp().Middle(example.TestMid).Middle(route.RouteAuthJwtMid)

	bindAddr := fmt.Sprintf("%s:%d", global.AppConfig.Host, global.AppConfig.Port)
	log.DebugLog("Bind address: ", bindAddr)

	routeManager.InitRoute(bindAddr)
}
