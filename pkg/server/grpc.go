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
	"io"
	"github.com/choerodon/c7n-slaver/pkg/mysql"
	"strings"
	"os/exec"
	"io/ioutil"
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
			return r, err
		}

		client := http.Client{}
		resp, err := client.Do(req)

		if err != nil {
			r.Success = false
			r.Message = err.Error()
			return r, err
		}

		if resp.StatusCode >= 400 || resp.StatusCode < 200 {
			r.Success = false
			r.Message = fmt.Sprintf("get response code %d", resp.StatusCode)
			return r, nil
		}
	}
	if c.Type == "socket" {
		addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
		conn, err := net.DialTimeout("tcp", addr, time.Second*2)
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

func (s *Server) ExecuteCommand(stream pb.RouteCall_ExecuteCommandServer) error {

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		log.Infof("executing: %s %s", in.Name, strings.Join(in.Args, " "))
		out, err := exec.Command(in.Name, in.Args...).Output()
		if err != nil {
			in.Success = false
			in.Message = err.Error()
			log.Error(err.Error())
		} else {
			in.Success = true
			in.Message = string(out)
		}
		if err := stream.Send(in); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) ExecuteSql(stream pb.RouteCall_ExecuteSqlServer) error {
	in, err := stream.Recv()
	if err != nil {
		log.Error(err)
		return err
	}
	mysql := mysql.Mysql{
		Username: in.Mysql.Username,
		Password: in.Mysql.Password,
		Host:     in.Mysql.Host,
		Port:     in.Mysql.Port,
	}
	db, err := mysql.Connect()
	defer db.Close()

	for {
		in, err = stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		log.Infof("executing: %s", in.Sql)
		_, err = db.Exec(in.Sql)

		if err != nil {
			goto err
		}
		log.Success("executed")
		stream.Send(&pb.RouteSql{
			Success: true,
		})
	}
err:
	log.Error(err)
	if err != nil {
		stream.Send(&pb.RouteSql{
			Success: false,
			Message: err.Error(),
		})
		return err
	}
	return nil
}

func (s *Server) ExecuteRequest(stream pb.RouteCall_ExecuteRequestServer) error {
	client := http.Client{}
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		reqUrl := fmt.Sprintf("%s://%s:%d%s", in.Schema, in.Host, in.Port, in.Path)
		log.Infof("%s: %s", in.Method, reqUrl)

		req, err := http.NewRequest(in.Method, reqUrl, strings.NewReader(in.Body))
		var header = make(map[string][]string)
		for k, v := range in.Header {
			header[k] = v.Value
		}
		req.Header = header
		resp, err := client.Do(req)
		out := &pb.Result{}
		if err != nil {
			out.Success = false
			out.Message = err.Error()
			log.Error(err.Error())
		} else {
			out.StatusCode = int32(resp.StatusCode)
			data, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				out.Success = false
			} else {
				out.Success = true
				out.Message = string(data)
			}
		}
		if err := stream.Send(out); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) InitGRpcServer(port int) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Errorf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterRouteCallServer(grpcServer, s)
	if err := grpcServer.Serve(lis); err != nil {
		log.Error(err)
	}
}
