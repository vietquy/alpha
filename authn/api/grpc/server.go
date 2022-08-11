package grpc

import (
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	alpha "github.com/vietquy/alpha"
	"github.com/vietquy/alpha/authn"
	"github.com/vietquy/alpha/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ alpha.AuthNServiceServer = (*grpcServer)(nil)

type grpcServer struct {
	issue    kitgrpc.Handler
	identify kitgrpc.Handler
}

// NewServer returns new AuthnServiceServer instance.
func NewServer(svc authn.Service) alpha.AuthNServiceServer {
	return &grpcServer{
		issue: kitgrpc.NewServer(
			issueEndpoint(svc),
			decodeIssueRequest,
			encodeIssueResponse,
		),
		identify: kitgrpc.NewServer(
			identifyEndpoint(svc),
			decodeIdentifyRequest,
			encodeIdentifyResponse,
		),
	}
}

func (s *grpcServer) Issue(ctx context.Context, req *alpha.IssueReq) (*alpha.Token, error) {
	_, res, err := s.issue.ServeGRPC(ctx, req)
	if err != nil {
		return nil, encodeError(err)
	}
	return res.(*alpha.Token), nil
}

func (s *grpcServer) Identify(ctx context.Context, token *alpha.Token) (*alpha.UserID, error) {
	_, res, err := s.identify.ServeGRPC(ctx, token)
	if err != nil {
		return nil, encodeError(err)
	}
	return res.(*alpha.UserID), nil
}

func decodeIssueRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*alpha.IssueReq)
	return issueReq{issuer: req.GetIssuer(), keyType: req.GetType()}, nil
}

func encodeIssueResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(identityRes)
	return &alpha.Token{Value: res.id}, encodeError(res.err)
}

func decodeIdentifyRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*alpha.Token)
	return identityReq{token: req.GetValue()}, nil
}

func encodeIdentifyResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(identityRes)
	return &alpha.UserID{Value: res.id}, encodeError(res.err)
}

func encodeError(err error) error {
	switch {
	case errors.Contains(err, nil):
		return nil
	case errors.Contains(err, authn.ErrMalformedEntity):
		return status.Error(codes.InvalidArgument, "received invalid token request")
	case errors.Contains(err, authn.ErrUnauthorizedAccess):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Contains(err, authn.ErrKeyExpired):
		return status.Error(codes.Unauthenticated, err.Error())
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
