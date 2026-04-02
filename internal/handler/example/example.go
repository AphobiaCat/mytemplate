package example

import (
	"context"

	"mytemplate/internal/types/example"
	"mytemplate/pkg/log"
	"mytemplate/pkg/route"
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

func TestSocketExample(msg string) (ret string, err error) {
	return
}

func TestSocketExampleNoreturn(msg string) {

}
