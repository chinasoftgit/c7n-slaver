package server

import (
	"google.golang.org/grpc"
	pb "github.com/choerodon/c7n-slaver/pkg/protobuf"
	"context"
)

func (s *Server) CheckHealth(ctx context.Context, c *pb.Check) (*pb.Result, error) {
	return nil, nil
}

func (s *Server) InitGRpcServer()  {
	grpcServer := grpc.NewServer()
	pb.RegisterRouteCallServer(grpcServer, s)
	s.ServerMux.Handle("/grpc",grpcServer)
}
