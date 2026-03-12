package example

import (
	"context"

	"mytemplate/internal/types/example"
)

func TestGetExample(ctx *context.Context, req *example.GetExampleRequest) (ret interface{}, err error) {

	ret = &example.GetExampleResponse{
		Message: "Hello " + req.Name,
	}

	return
}

func TestPostExample(ctx *context.Context, req *example.PostExampleRequest) (ret interface{}, err error) {

	ret = &example.PostExampleResponse{
		Message: "Hello " + req.Name,
	}

	return
}
