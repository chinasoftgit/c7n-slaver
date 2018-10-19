package server

import (
	"google.golang.org/grpc"
	pb "github.com/choerodon/c7n-slaver/pkg/protobuf"
	"context"
	"github.com/vinkdong/gox/log"
	"net"
	"fmt"
	"net/http"
)

func (s *Server) CheckHealth(ctx context.Context, c *pb.Check) (*pb.Result, error) {
	log.Infof("checking %s", c.Host)
	r := &pb.Result{
		Success: true,
	}
	if c.Type == "httpGet" {
		url := fmt.Sprintf("%s://%s:%d%d",c.Schema,c.Host,c.Port,c.Path)
		log.Infof("getting %s",url)
		resp ,err := http.Get(url)
		if err !=nil {
			r.Success = false
			r.Message = err.Error()
			return r,err
		}
		if resp.StatusCode >= 400 || resp.StatusCode < 200 {
			r.Success = false
			r.Message = fmt.Sprintf("get response code %s",resp.StatusCode)
			return r,nil
		}
	}
	if c.Type == "socket" {
		addr := fmt.Sprintf("%s:%d",c.Host,c.Port)
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			log.Infof("Connection error: %s", err)
			r.Success = false
			r.Message = err.Error()
		} else {
			defer conn.Close()
		}
	}
	return r, nil
}

func (s *Server) InitGRpcServer(port int) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d",port))
	if err != nil {
		log.Errorf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterRouteCallServer(grpcServer, s)
	if err := grpcServer.Serve(lis); err != nil {
		log.Error(err)
	}
}
