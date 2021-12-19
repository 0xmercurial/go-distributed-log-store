package server

import (
	"context"
	"logstore/internal/log/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type Config struct {
	CommitLog  CommitLog
	Authorizer Authorizer
}

type CommitLog interface {
	Append(*proto.Record) (uint64, error)
	Read(uint64) (*proto.Record, error)
}

type Authorizer interface {
	Authorize(subject, object, action string) error
}

type subjectContextKey struct{}

const (
	objWildCard  = "*"
	appendAction = "append"
	readAction   = "read"
)

func NewGRPCServer(config *Config, opts ...grpc.ServerOption) (
	*grpc.Server,
	error,
) {
	grpcSrv := grpc.NewServer(opts...)
	srv, err := newGrpcServer(config)
	if err != nil {
		return nil, err
	}
	proto.RegisterLogServer(grpcSrv, srv)
	return grpcSrv, nil
}

type grpcServer struct {
	*Config
}

func newGrpcServer(config *Config) (srv *grpcServer, err error) {
	srv = &grpcServer{
		Config: config,
	}
	return srv, nil
}

func (s *grpcServer) Append(
	ctx context.Context,
	req *proto.AppendRequest,
) (*proto.AppendResponse, error) {
	if err := s.Authorizer.Authorize(
		subject(ctx),
		objWildCard,
		appendAction,
	); err != nil {
		return nil, err
	}
	off, err := s.CommitLog.Append(req.Record)
	if err != nil {
		return nil, err
	}
	return &proto.AppendResponse{Offset: off}, nil
}

func (s *grpcServer) AppendStream(
	stream proto.Log_AppendStreamServer, //interface, not pointer to interface
) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			return err
		}
		res, err := s.Append(stream.Context(), req)
		if err != nil {
			return err
		}
		if err = stream.Send(res); err != nil {
			return err
		}
	}
}

func (s *grpcServer) Read(ctx context.Context, req *proto.ReadRequest) (
	*proto.ReadResponse, error) {
	if err := s.Authorizer.Authorize(
		subject(ctx),
		objWildCard,
		readAction,
	); err != nil {
		return nil, err
	}
	record, err := s.CommitLog.Read(req.Offset)
	if err != nil {
		return nil, err
	}
	return &proto.ReadResponse{Record: record}, nil
}

func (s *grpcServer) ReadStream(
	req *proto.ReadRequest,
	stream proto.Log_ReadStreamServer,
) error {
	for {
		select {
		case <-stream.Context().Done():
			return nil
		default:
			res, err := s.Read(stream.Context(), req)
			switch err.(type) {
			case nil:
			case proto.ErrOffOutOfRange:
				continue
			default:
				return err
			}
			if err = stream.Send(res); err != nil {
				return err
			}
			req.Offset++
		}
	}
}

func authenticate(ctx context.Context) (context.Context, error) {
	peer, ok := peer.FromContext(ctx)
	if !ok {
		return ctx, status.New(
			codes.Unknown,
			"couldn't find peer info",
		).Err()
	}

	if peer.AuthInfo == nil {
		return context.WithValue(ctx, subjectContextKey{}, ""), nil
	}

	tlsInfo := peer.AuthInfo.(credentials.TLSInfo)
	subject := tlsInfo.State.VerifiedChains[0][0].Subject.CommonName
	ctx = context.WithValue(ctx, subjectContextKey{}, subject)

	return ctx, nil
}
