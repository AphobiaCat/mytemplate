package handler

import (
	"fmt"
	"mytemplate/internal/global"
	"mytemplate/internal/handler/example"
	"mytemplate/pkg/log"
	"mytemplate/pkg/route"
)

func Setup() {
	// setup server routes

	routeManager := route.New()

	routeManager.RouteGet("example/get", example.TestGetExample).NeedUserIp().RecvParams("name")
	routeManager.RoutePost("example/post", example.TestPostExample)

	bindAddr := fmt.Sprintf("%s:%d", global.AppConfig.Host, global.AppConfig.Port)
	log.DebugLog("Bind address: ", bindAddr)

	routeManager.InitRoute(bindAddr)
}
