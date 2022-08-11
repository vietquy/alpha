package grpc

import (
	"github.com/go-kit/kit/endpoint"
	"github.com/vietquy/alpha/things"
	context "golang.org/x/net/context"
)

func canAccessEndpoint(svc things.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(AccessByKeyReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		id, err := svc.CanAccessByKey(ctx, req.projectID, req.thingKey)
		if err != nil {
			return identityRes{err: err}, err
		}
		return identityRes{id: id, err: nil}, nil
	}
}

func canAccessByIDEndpoint(svc things.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(accessByIDReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		err := svc.CanAccessByID(ctx, req.projectID, req.thingID)
		return emptyRes{err: err}, err
	}
}

func identifyEndpoint(svc things.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(identifyReq)
		id, err := svc.Identify(ctx, req.key)
		if err != nil {
			return identityRes{err: err}, err
		}
		return identityRes{id: id, err: nil}, nil
	}
}
