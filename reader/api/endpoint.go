package api

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/vietquy/alpha/reader"
)

func listMessagesEndpoint(svc reader.MessageRepository) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		req := request.(listMessagesReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		page, err := svc.ReadAll(req.projectID, req.pageMeta)
		if err != nil {
			return nil, err
		}

		return pageRes{
			PageMetadata: page.PageMetadata,
			Total:        page.Total,
			Messages:     page.Messages,
		}, nil
	}
}
