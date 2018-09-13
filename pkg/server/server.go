package server

import (
	"net/http"
	"github.com/vinkdong/gox/log"
	"io/ioutil"
	"encoding/json"
	"fmt"
)

const (
	stopServer = 0
	stopPorts  = 1
)

type Server struct {
	Addr      string
	ServerMux *http.ServeMux
	Server    *http.Server
}

var startedServer []*Server

func NewServer(port int) *Server {
	s := &Server{
		Addr:      fmt.Sprintf(":%d", port),
		ServerMux: http.NewServeMux(),
	}
	return s
}

func (s *Server) HandlerInit() {
	s.ServerMux.HandleFunc("/", networkCheckHandler)
	s.ServerMux.HandleFunc("/ports/start", startPortHandler)
	s.ServerMux.HandleFunc("/ports/stop", stopPortHandler)
}

func (s *Server) AddHealthHandler() {
	s.ServerMux.HandleFunc("/health", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte(`{"status":"OK"}`))
	})
}

func (s *Server) Start() error {
	server := &http.Server{
		Addr:    s.Addr,
		Handler: s.ServerMux,
	}
	s.Server = server
	log.Infof("server starting on %s", s.Addr)
	return server.ListenAndServe()
}

func startPortHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.Write([]byte(`{"success":false"}`))
	}
	portRequest := &PortRequest{}
	json.Unmarshal(data, portRequest)
	err = portRequest.StartServers()
	if err != nil {
		w.Write([]byte(`{"success":false"}`))
	}
	w.Write([]byte(`{"success":true}`))

}

func stopPortHandler(w http.ResponseWriter, r *http.Request) {

	for _, s := range startedServer {
		log.Infof("stop server %s", s.Server.Addr)
		err := s.Server.Shutdown(nil)
		if err !=nil {
			log.Error(err)
		}
	}
	startedServer = make([]*Server, 0)
	w.Write([]byte(`{"success":true}`))
}

func networkCheckHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.Write([]byte(`{"success":false"}`))
	}
	w.Write([]byte("ok"))
	log.Infof("%s", data)
}
