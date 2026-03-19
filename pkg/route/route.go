package route

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"net/http"

	"mytemplate/internal/global"
	"mytemplate/pkg/log"
	"mytemplate/pkg/redis"
	"mytemplate/pkg/util"
)

var allowOrigins = []string{"*"}
var allowMethods = []string{"*"} //[]string{"GET", "POST", "PUT", "DELETE"}
var allowHeaders = []string{"*"}

const midDataKey = "MidData"

type mitCallback func(params map[string]string) (map[string]string, bool)
type handlerFunc func(*context.Context, []byte) (interface{}, error)

type RouteManager struct {
	httpService *gin.Engine

	routes    []Route
	routesLen uint32
}

type Route struct {
	api          string
	isGet        bool
	recvParams   []string
	alert        string
	midCallbacks []mitCallback
	midParams    [][]string
	midAlert     []string
	midIndex     int
	reqLimit     int
	reloadLimitS int64
	needUserIp   bool
	callback     handlerFunc
}

func New() *RouteManager {
	ret := &RouteManager{}

	return ret
}

func (rm *RouteManager) RoutePost(api string, handler interface{}) *Route {

	ret := rm.RouteGet(api, handler)
	ret.isGet = false

	return ret
}

func (rm *RouteManager) RouteGet(api string, handler interface{}) *Route {
	rm.routes = append(rm.routes, Route{})

	ret := &(rm.routes[rm.routesLen])
	rm.routesLen += 1

	ret.api = api
	ret.isGet = true
	ret.midIndex = -1
	ret.reqLimit = 0
	ret.reloadLimitS = 60

	v := reflect.ValueOf(handler)
	t := v.Type()

	reqType := t.In(1)

	wrapper := func(ctx *context.Context, data []byte) (interface{}, error) {

		req := reflect.New(reqType.Elem()).Interface()

		err := json.Unmarshal(data, req)
		if err != nil {
			return nil, err
		}

		results := v.Call([]reflect.Value{
			reflect.ValueOf(ctx),
			reflect.ValueOf(req),
		})

		if !results[1].IsNil() {
			return nil, results[1].Interface().(error)
		}

		return results[0].Interface(), nil
	}

	ret.callback = wrapper

	return ret
}

func (r *Route) RecvParams(params ...string) *Route {
	r.recvParams = append(r.recvParams, params...)

	return r
}

func (r *Route) Alert(alert string) *Route {

	r.alert = alert

	return r
}

func (r *Route) Middle(middle mitCallback) *Route {
	r.midCallbacks = append(r.midCallbacks, middle)
	r.midIndex += 1
	r.midParams = append(r.midParams, []string{})
	r.midAlert = append(r.midAlert, "")

	return r
}

func (r *Route) MiddleParams(params ...string) *Route {
	if r.midIndex >= 0 {
		r.midParams[r.midIndex] = append(r.midParams[r.midIndex], params...)
	}

	return r
}

func (r *Route) MiddleAlert(alert string) *Route {
	if r.midIndex >= 0 {
		r.midAlert[r.midIndex] = alert
	}

	return r
}

func (r *Route) ReqLimit(count int, reloadtime ...int64) *Route {
	r.reqLimit = count
	if len(reloadtime) != 0 {
		r.reloadLimitS = reloadtime[0]
	}

	return r
}

func (r *Route) NeedUserIp() *Route {
	r.needUserIp = true

	return r
}

func streamcontrol(api string, ip string, calllimit int, reloadtime int64) bool {

	if calllimit == 0 {
		return true
	}

	rediskey := "streamcontrol:" + api + "" + ip
	count := redis.TimerCount(rediskey, int64(calllimit), reloadtime)

	if count >= 0 {
		return true
	} else {
		return false
	}
}

func processroutemiddlewaremodule(process mitCallback, needheader []string, errinfo string) gin.HandlerFunc {
	return func(c *gin.Context) {

		useheaderarray := make(map[string]string)

		for _, val := range needheader {
			useheaderarray[val] = c.GetHeader(val)
		}

		userdata, ret := process(useheaderarray)

		if len(userdata) != 0 {
			userinfointerface, exist := c.Get(midDataKey)

			var newuserinfo map[string]string

			if exist {
				newuserinfo = userinfointerface.(map[string]string)
			} else {
				newuserinfo = make(map[string]string)
			}

			for key, val := range userdata {
				newuserinfo[key] = val
			}

			c.Set(midDataKey, newuserinfo)
		}

		if ret {
			c.Next()

		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"code": -1, "error": errinfo})
			c.Abort()
		}
	}
}

func (rm *RouteManager) InitRoute(bindaddr string) {
	gin.SetMode(gin.ReleaseMode)
	rm.httpService = gin.New()

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = allowOrigins
	corsConfig.AllowMethods = allowMethods
	corsConfig.AllowHeaders = allowHeaders

	rm.httpService.Use(cors.New(corsConfig))

	rm.httpService.SetTrustedProxies(global.AppConfig.TrustedProxies) //only trust local proxy

	for _, route := range rm.routes {
		if route.callback == nil {
			log.DebugError("route[", route.api, "] no handler func")
			continue
		}

		if route.isGet {

			getRouteprocess := func(ginContext *gin.Context) {
				ctx := ginContext.Request.Context()
				defer func() {
					if err := recover(); err != nil {
						log.DebugError(route.api, " err:", err)
					}
				}()

				clientIP := ginContext.ClientIP()

				if !streamcontrol(route.api, clientIP, route.reqLimit, route.reloadLimitS) {
					ginContext.JSON(http.StatusOK, gin.H{
						"code":  -429,
						"error": "too many requests",
					})
					return
				}

				params := make(map[string]string)

				for _, keyval := range route.recvParams {
					if val, exists := ginContext.GetQuery(keyval); exists {
						params[keyval] = val
					} else {
						log.DebugError(route.api, " key[", keyval, "] no exist")
					}
				}

				if route.needUserIp {
					params["ip"] = clientIP
				}

				midParamsi, _ := ginContext.Get(midDataKey)

				if midParams, ok := midParamsi.(map[string]string); ok {
					for key, val := range midParams {
						params[key] = val
					}
				}

				jsonParams, err := json.Marshal(params)
				if err != nil {
					log.DebugError(route.api, " json marshal err:", err)
				}

				ret, err := route.callback(&ctx, jsonParams)

				if err == nil {
					ginContext.JSON(http.StatusOK, gin.H{
						"code": 0,
						"data": ret,
					})
				} else {
					ginContext.JSON(http.StatusOK, gin.H{
						"code":  -1,
						"error": err.Error(),
					})

					log.DebugError(route.api, "  err:", route.alert)
				}
			}

			if len(route.midCallbacks) > 0 {
				midsfunc := []gin.HandlerFunc{}

				for index, midprocess := range route.midCallbacks {
					midsfunc = append(midsfunc, processroutemiddlewaremodule(midprocess, route.midParams[index], route.midAlert[index]))
				}
				midsfunc = append(midsfunc, getRouteprocess)

				rm.httpService.GET(route.api, midsfunc...)
			} else {
				rm.httpService.GET(route.api, getRouteprocess)
			}

			log.DebugLog("Get  --> ", route.api)
		} else {

			postRouteprocess := func(ginContext *gin.Context) {
				ctx := ginContext.Request.Context()

				defer func() {
					if err := recover(); err != nil {
						log.DebugError(route.api, " err:", err)
					}
				}()

				clientIP := ginContext.ClientIP()

				if !streamcontrol(route.api, clientIP, route.reqLimit, route.reloadLimitS) {
					ginContext.JSON(http.StatusOK, gin.H{
						"code":  -429,
						"error": "too many requests",
					})
					return
				}

				body, err := ginContext.GetRawData()

				if err != nil {
					log.DebugError(route.api, " input data no exist:", body)
				}

				bodystr := string(body)

				if route.needUserIp || len(route.midCallbacks) > 0 {
					tmpmap := make(map[string]interface{})
					util.ParserJson(bodystr, &tmpmap)

					if route.needUserIp {
						tmpmap["ip"] = clientIP
					}

					midParamsi, _ := ginContext.Get(midDataKey)

					if midParams, ok := midParamsi.(map[string]string); ok {
						for key, val := range midParams {
							if key == "jwt" {
								tmpmap[key] = json.RawMessage(val)
							} else {
								tmpmap[key] = val
							}
						}
					}

					body, err = json.Marshal(tmpmap)
					if err != nil {
						log.DebugError(route.api, " json marshal err:", err)
					}
				}

				ret, err := route.callback(&ctx, body)

				if err == nil {
					ginContext.JSON(http.StatusOK, gin.H{
						"code": 0,
						"data": ret,
					})
				} else {
					ginContext.JSON(http.StatusOK, gin.H{
						"code":  -1,
						"error": err.Error(),
					})

					log.DebugError(route.api, " err:", route.alert)
				}
			}

			if len(route.midCallbacks) > 0 {
				midsfunc := []gin.HandlerFunc{}

				for index, midprocess := range route.midCallbacks {
					midsfunc = append(midsfunc, processroutemiddlewaremodule(midprocess, route.midParams[index], route.midAlert[index]))
				}
				midsfunc = append(midsfunc, postRouteprocess)

				rm.httpService.POST(route.api, midsfunc...)
			} else {
				rm.httpService.POST(route.api, postRouteprocess)
			}

			log.DebugLog("Post --> ", route.api)
		}
	}

	log.DebugLog("bind addr :", bindaddr)
	if err := rm.httpService.Run(bindaddr); err != nil {
		panic(err)
	}
}
