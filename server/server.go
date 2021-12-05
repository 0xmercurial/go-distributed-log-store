package server

import (
	"context"
	"logstore/internal/log/proto"

	"google.golang.org/grpc"
)

type CommitLog interface {
	Append(*proto.Record) (uint64, error)
	Read(uint64) (*proto.Record, error)
}

type Config struct {
	CommitLog CommitLog
}

func NewGRPCServer(config *Config) (*grpc.Server, error) {
	grpcSrv := grpc.NewServer()
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
