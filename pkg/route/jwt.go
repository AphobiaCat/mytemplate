package route

import (
	"context"
	"fmt"
	"mytemplate/internal/global"
	"mytemplate/pkg/log"
	"mytemplate/pkg/util"

	"github.com/golang-jwt/jwt/v5"

	"time"
)

var jwtKey = []byte{}

type JwtReq struct {
	Token string `header:"token"`
}

type JwtRes struct {
	UserInfo string `json:"userinfo"`
}

func RouteGenerateJwtByStr(data string, expirationSecsOption ...int) (string, error) {
	expirationtime := 60 * 60

	if len(expirationSecsOption) != 0 && expirationSecsOption[0] > 0 {
		expirationtime = expirationSecsOption[0]
	}

	claims := jwt.MapClaims{
		"data": data,
		"exp":  time.Now().Add(time.Duration(expirationtime) * time.Second).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	ret, err := token.SignedString(jwtKey)

	if err != nil {
		log.DebugError("generate jwt failed. user data:", data)
		return ret, err
	}

	return ret, err
}

func RouteGenerateJwt(data interface{}, expirationSecsOption ...int) (string, error) {
	expirationtime := 60 * 60

	if len(expirationSecsOption) != 0 && expirationSecsOption[0] > 0 {
		expirationtime = expirationSecsOption[0]
	}

	claims := jwt.MapClaims{
		"data": util.BuildJson(data),
		"exp":  time.Now().Add(time.Duration(expirationtime) * time.Second).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	ret, err := token.SignedString(jwtKey)

	if err != nil {
		log.DebugError("generate jwt failed. user data:", data)
		return ret, err
	}

	return ret, err
}

func RouteParserJwt(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		log.DebugError("tokenString[", tokenString, "] parser jwt failed. err:", err)

		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {

		ret, exist := claims["data"]

		if !exist {
			log.DebugError("jwt haven't data")
			return "", err
		}

		return ret.(string), err
	}

	log.DebugError("parser jwt failed. err:", err)

	return "", err
}

func RouteAuthJwtMid(ctx *context.Context, req *JwtReq) (ret *JwtRes, err error) {

	data, err := RouteParserJwt(req.Token)

	if err == nil {
		ret = &JwtRes{
			UserInfo: data,
		}
		return
	}

	ret = &JwtRes{}
	err = fmt.Errorf("parse jwt token error:%s", err.Error())

	return
}

func init() {
	go func() {
		util.Sleep(1000)
		jwtKey = []byte(global.AppConfig.ServerJwtKey)
	}()
}
