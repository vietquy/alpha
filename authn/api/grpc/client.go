package grpc

import (
	"time"

	"github.com/go-kit/kit/endpoint"
	kitgrpc "github.com/go-kit/kit/transport/grpc"

	"github.com/vietquy/alpha"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var _ alpha.AuthNServiceClient = (*grpcClient)(nil)

type grpcClient struct {
	issue    endpoint.Endpoint
	identify endpoint.Endpoint
	timeout  time.Duration
}

// NewClient returns new gRPC client instance.
func NewClient(conn *grpc.ClientConn, timeout time.Duration) alpha.AuthNServiceClient {
	return &grpcClient{
		issue: kitgrpc.NewClient(
			conn,
			"alpha.AuthNService",
			"Issue",
			encodeIssueRequest,
			decodeIssueResponse,
			alpha.UserID{},
		).Endpoint(),
		identify: kitgrpc.NewClient(
			conn,
			"alpha.AuthNService",
			"Identify",
			encodeIdentifyRequest,
			decodeIdentifyResponse,
			alpha.UserID{},
		).Endpoint(),
		timeout: timeout,
	}
}

func (client grpcClient) Issue(ctx context.Context, req *alpha.IssueReq, _ ...grpc.CallOption) (*alpha.Token, error) {
	ctx, close := context.WithTimeout(ctx, client.timeout)
	defer close()

	res, err := client.issue(ctx, issueReq{issuer: req.GetIssuer(), keyType: req.Type})
	if err != nil {
		return nil, err
	}

	ir := res.(identityRes)
	return &alpha.Token{Value: ir.id}, ir.err
}

func encodeIssueRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(issueReq)
	return &alpha.IssueReq{Issuer: req.issuer, Type: req.keyType}, nil
}

func decodeIssueResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(*alpha.UserID)
	return identityRes{res.GetValue(), nil}, nil
}

func (client grpcClient) Identify(ctx context.Context, token *alpha.Token, _ ...grpc.CallOption) (*alpha.UserID, error) {
	ctx, close := context.WithTimeout(ctx, client.timeout)
	defer close()

	res, err := client.identify(ctx, identityReq{token: token.GetValue()})
	if err != nil {
		return nil, err
	}

	ir := res.(identityRes)
	return &alpha.UserID{Value: ir.id}, ir.err
}

func encodeIdentifyRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(identityReq)
	return &alpha.Token{Value: req.token}, nil
}

func decodeIdentifyResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(*alpha.UserID)
	return identityRes{res.GetValue(), nil}, nil
}
