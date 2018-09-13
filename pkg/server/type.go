package server

import (
	"github.com/vinkdong/gox/log"
)

type PortRequest struct {
	Ports []int `json:"ports"`
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
