package server

import (
	"github.com/vinkdong/gox/log"
	"net/http"
	"time"
	"fmt"
)

type PortRequest struct {
	Ports []int `json:"ports"`
}

type ServerAddr struct {
	Hosts []string `json:"hosts"`
	Ports []int `json:"ports"`
}

type MountPathInfo struct {
	Path string `json:"path"`
	Require string `json:"require"`
}

type StorageStatus struct {
	Success bool `json:"success"`
	Free string  `json:"free"`
}


func (s *ServerAddr) StartNetCheck() (err error){
	c := &http.Client{
		Timeout: time.Second,
	}
	for _, ip := range s.Hosts {
		for _, port := range s.Ports {
			_, err = c.Get(fmt.Sprint("http://",ip,":",port,"/health"))
			if err != nil {
				log.Error(err)
				return
			}
			log.Info(fmt.Sprint("http://",ip,":",port,"/health"," ok"))
		}
	}
	return
}

func (p *PortRequest) StartServers() error {
	for _, port := range p.Ports {
		s := NewServer(port)
		s.AddHealthHandler()
		startedServer = append(startedServer, s)
		go func() {
			err := s.Start()
			if err != nil {
				log.Error(err)
			}
		}()
	}
	return nil
}
