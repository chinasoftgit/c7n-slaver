package server

import (
	"google.golang.org/grpc"
	pb "github.com/choerodon/c7n-slaver/pkg/protobuf"
	"context"
	"github.com/vinkdong/gox/log"
	"net"
	"fmt"
	"net/http"
	"time"
)

func (s *Server) CheckHealth(ctx context.Context, c *pb.Check) (*pb.Result, error) {

	r := &pb.Result{
		Success: true,
	}
	if c.Type == "httpGet" {

		url := fmt.Sprintf("%s://%s:%d%s", c.Schema, c.Host, c.Port, c.Path)
		if c.Port == 80 || c.Port == 443 {
			url = fmt.Sprintf("%s://%s%s", c.Schema, c.Host, c.Path)
		}

		log.Infof("http checking %s", url)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			r.Success = false
			r.Message = err.Error()
			return r,err
		}

		client := http.Client{}
		resp ,err := client.Do(req)

		if err !=nil {
			r.Success = false
			r.Message = err.Error()
			return r,err
		}

		if resp.StatusCode >= 400 || resp.StatusCode < 200 {
			r.Success = false
			r.Message = fmt.Sprintf("get response code %d",resp.StatusCode)
			return r,nil
		}
	}
	if c.Type == "socket" {
		addr := fmt.Sprintf("%s:%d",c.Host,c.Port)
		conn, err := net.DialTimeout("tcp", addr,time.Second * 2)
		log.Infof("socket checking %s", addr)
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
