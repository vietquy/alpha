package api

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/vietquy/alpha/http"
)

func sendMessageEndpoint(svc http.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(publishReq)
		err := svc.Publish(ctx, req.token, req.msg)
		return nil, err
	}
}
