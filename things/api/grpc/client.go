package grpc

import (
	"time"

	"github.com/go-kit/kit/endpoint"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/vietquy/alpha"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var _ alpha.ThingsServiceClient = (*grpcClient)(nil)

type grpcClient struct {
	timeout        time.Duration
	canAccessByKey endpoint.Endpoint
	canAccessByID  endpoint.Endpoint
	identify       endpoint.Endpoint
}

// NewClient returns new gRPC client instance.
func NewClient(conn *grpc.ClientConn, timeout time.Duration) alpha.ThingsServiceClient {
	svcName := "alpha.ThingsService"

	return &grpcClient{
		timeout: timeout,
		canAccessByKey: kitgrpc.NewClient(
			conn,
			svcName,
			"CanAccessByKey",
			encodeCanAccessByKeyRequest,
			decodeIdentityResponse,
			alpha.ThingID{},
		).Endpoint(),
		canAccessByID: kitgrpc.NewClient(
			conn,
			svcName,
			"CanAccessByID",
			encodeCanAccessByIDRequest,
			decodeEmptyResponse,
			empty.Empty{},
		).Endpoint(),
		identify: kitgrpc.NewClient(
			conn,
			svcName,
			"Identify",
			encodeIdentifyRequest,
			decodeIdentityResponse,
			alpha.ThingID{},
		).Endpoint(),
	}
}

func (client grpcClient) CanAccessByKey(ctx context.Context, req *alpha.AccessByKeyReq, _ ...grpc.CallOption) (*alpha.ThingID, error) {
	ctx, cancel := context.WithTimeout(ctx, client.timeout)
	defer cancel()

	ar := AccessByKeyReq{
		thingKey: req.GetToken(),
		projectID:   req.GetProjectID(),
	}
	res, err := client.canAccessByKey(ctx, ar)
	if err != nil {
		return nil, err
	}

	ir := res.(identityRes)
	return &alpha.ThingID{Value: ir.id}, ir.err
}

func (client grpcClient) CanAccessByID(ctx context.Context, req *alpha.AccessByIDReq, _ ...grpc.CallOption) (*empty.Empty, error) {
	ar := accessByIDReq{thingID: req.GetThingID(), projectID: req.GetProjectID()}
	res, err := client.canAccessByID(ctx, ar)
	if err != nil {
		return nil, err
	}

	er := res.(emptyRes)
	return &empty.Empty{}, er.err
}

func (client grpcClient) Identify(ctx context.Context, req *alpha.Token, _ ...grpc.CallOption) (*alpha.ThingID, error) {
	ctx, cancel := context.WithTimeout(ctx, client.timeout)
	defer cancel()

	res, err := client.identify(ctx, identifyReq{key: req.GetValue()})
	if err != nil {
		return nil, err
	}

	ir := res.(identityRes)
	return &alpha.ThingID{Value: ir.id}, ir.err
}

func encodeCanAccessByKeyRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(AccessByKeyReq)
	return &alpha.AccessByKeyReq{Token: req.thingKey, ProjectID: req.projectID}, nil
}

func encodeCanAccessByIDRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(accessByIDReq)
	return &alpha.AccessByIDReq{ThingID: req.thingID, ProjectID: req.projectID}, nil
}

func encodeIdentifyRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(identifyReq)
	return &alpha.Token{Value: req.key}, nil
}

func decodeIdentityResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(*alpha.ThingID)
	return identityRes{id: res.GetValue(), err: nil}, nil
}

func decodeEmptyResponse(_ context.Context, _ interface{}) (interface{}, error) {
	return emptyRes{}, nil
}
