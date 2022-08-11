package grpc

import (
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/vietquy/alpha"
	"github.com/vietquy/alpha/things"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ alpha.ThingsServiceServer = (*grpcServer)(nil)

type grpcServer struct {
	canAccessByKey kitgrpc.Handler
	canAccessByID  kitgrpc.Handler
	identify       kitgrpc.Handler
}

// NewServer returns new ThingsServiceServer instance.
func NewServer(svc things.Service) alpha.ThingsServiceServer {
	return &grpcServer{
		canAccessByKey: kitgrpc.NewServer(
			canAccessEndpoint(svc),
			decodeCanAccessByKeyRequest,
			encodeIdentityResponse,
		),
		canAccessByID: kitgrpc.NewServer(
			canAccessByIDEndpoint(svc),
			decodeCanAccessByIDRequest,
			encodeEmptyResponse,
		),
		identify: kitgrpc.NewServer(
			identifyEndpoint(svc),
			decodeIdentifyRequest,
			encodeIdentityResponse,
		),
	}
}

func (gs *grpcServer) CanAccessByKey(ctx context.Context, req *alpha.AccessByKeyReq) (*alpha.ThingID, error) {
	_, res, err := gs.canAccessByKey.ServeGRPC(ctx, req)
	if err != nil {
		return nil, encodeError(err)
	}

	return res.(*alpha.ThingID), nil
}

func (gs *grpcServer) CanAccessByID(ctx context.Context, req *alpha.AccessByIDReq) (*empty.Empty, error) {
	_, res, err := gs.canAccessByID.ServeGRPC(ctx, req)
	if err != nil {
		return nil, encodeError(err)
	}

	return res.(*empty.Empty), nil
}

func (gs *grpcServer) Identify(ctx context.Context, req *alpha.Token) (*alpha.ThingID, error) {
	_, res, err := gs.identify.ServeGRPC(ctx, req)
	if err != nil {
		return nil, encodeError(err)
	}

	return res.(*alpha.ThingID), nil
}

func decodeCanAccessByKeyRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*alpha.AccessByKeyReq)
	return AccessByKeyReq{thingKey: req.GetToken(), projectID: req.GetProjectID()}, nil
}

func decodeCanAccessByIDRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*alpha.AccessByIDReq)
	return accessByIDReq{thingID: req.GetThingID(), projectID: req.GetProjectID()}, nil
}

func decodeIdentifyRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*alpha.Token)
	return identifyReq{key: req.GetValue()}, nil
}

func encodeIdentityResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(identityRes)
	return &alpha.ThingID{Value: res.id}, encodeError(res.err)
}

func encodeEmptyResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(emptyRes)
	return &empty.Empty{}, encodeError(res.err)
}

func encodeError(err error) error {
	switch err {
	case nil:
		return nil
	case things.ErrMalformedEntity:
		return status.Error(codes.InvalidArgument, "received invalid can access request")
	case things.ErrUnauthorizedAccess:
		return status.Error(codes.PermissionDenied, "missing or invalid credentials provided")
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
