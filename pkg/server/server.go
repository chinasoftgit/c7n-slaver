package server

import (
	"net/http"
	"github.com/vinkdong/gox/log"
	"io/ioutil"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"strings"
	"syscall"
	"k8s.io/apimachinery/pkg/api/resource"
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
var DomainMap  = make(map[string]string)

func NewServer(port int) *Server {
	s := &Server{
		Addr:      fmt.Sprintf(":%d", port),
		ServerMux: http.NewServeMux(),
	}
	return s
}

func (s *Server) HandlerInit() {
	s.ServerMux.HandleFunc("/network", networkCheckHandler)
	s.ServerMux.HandleFunc("/ports/start", startPortHandler)
	s.ServerMux.HandleFunc("/ports/stop", stopPortHandler)
	//s.ServerMux.HandleFunc("/storage", storageCheckHandler)
	s.ServerMux.HandleFunc("/cmd", cmdHandler)
	s.ServerMux.HandleFunc("/mysql", mysqlCheckHandler)
	s.ServerMux.HandleFunc("/c7n/acme-challenge", c7nAcmeHandler)
	s.ServerMux.HandleFunc("/forward", forwardHandler)
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

func mysqlCheckHandler(w http.ResponseWriter, r *http.Request){
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.Write([]byte(`{"success":false}`))
	}
	dataRequest := &Requst{}
	json.Unmarshal(data, dataRequest)
	db,err := dataRequest.Mysql.ConnetMySql()
	defer db.Close()
	if err != nil {
		w.Write([]byte(`{"success":false}`))
		return
	}
	err = dataRequest.Executed(db)
	if err != nil {
		w.Write([]byte(`{"success":false}`))
	} else {
		log.Info("execute success")
		w.Write([]byte(`{"success":true}`))
	}
}

func startPortHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.Write([]byte(`{"success":false}`))
	}
	portRequest := &PortRequest{}
	json.Unmarshal(data, portRequest)
	err = portRequest.StartServers()
	if err != nil {
		w.Write([]byte(`{"success":false}`))
	} else {
		w.Write([]byte(`{"success":true}`))
	}
}

func stopPortHandler(w http.ResponseWriter, r *http.Request) {

	for _, s := range startedServer {
		log.Infof("stop server %s", s.Server.Addr)
		err := s.Server.Shutdown(nil)
		if err != nil {
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
		w.Write([]byte(`{"success":false}`))
	}
	serverAddr := &ServerAddr{}
	json.Unmarshal(data, serverAddr)
	err = serverAddr.StartNetCheck()
	if err != nil {
		w.Write([]byte(`{"success":false}`))
	} else {
		w.Write([]byte(`{"success":true}`))
	}
}

func DiskUsage(path string) (diskFree int64) {
	fs := syscall.Statfs_t{}
	err := syscall.Statfs(path, &fs)
	if err != nil {
		return
	}
	diskFree = int64(fs.Bfree * uint64(fs.Bsize))
	return
}
func storageCheckHandler(w http.ResponseWriter, r *http.Request)  {
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.Write([]byte(`{"success":false}`))
	}
	mountPathInfo := &MountPathInfo{}
	json.Unmarshal(data, mountPathInfo)
	re, err := resource.ParseQuantity(mountPathInfo.Require)
	if err != nil {
		log.Error(err)
		return
	}
	diskFree := DiskUsage(mountPathInfo.Path)
	storageStatus := &StorageStatus{}
	if re.CmpInt64(diskFree) <= 0 {
		storageStatus.Success = true
	} else {
		storageStatus.Success = false
		memorySize := resource.NewQuantity(diskFree, resource.BinarySI)
		storageStatus.Free = fmt.Sprintf("%v",memorySize)
	}
	b, _ := json.Marshal(storageStatus)
	w.Write([]byte(b))
}

func cmdHandler(w http.ResponseWriter, r *http.Request)  {
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.Write([]byte(`{"success":false}`))
	}
	cmdExec := &CommandExec{}
	json.Unmarshal(data, cmdExec)
	err = cmdExec.ExecuteCommand()
	if err != nil {
		w.Write([]byte(`{"success":false}`))
	} else {
		log.Infof("execute command %s success", cmdExec.CommandLine)
		w.Write([]byte(`{"success":true}`))
	}
}
func c7nAcmeHandler(w http.ResponseWriter, r *http.Request)  {
	if r.Method == http.MethodPost {
		data , err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			log.Error(err)
			w.Write([]byte(`{"success":false}`))
		} else {
			r := &Request{}
			json.Unmarshal(data,r)
			DomainMap[r.Domain] = r.Value
			log.Infof("add domain map: %s => %s",r.Domain,r.Value)
			w.Write([]byte(`{"success":true}`))
		}
	} else {
		domain := r.Host
		loc := strings.Index(domain,":")
		if loc != -1 {
			domain = domain[:loc]
		}
		log.Infof("has request in, domain is %s ,method is %s",domain,r.Method)
		w.Write([]byte(DomainMap[domain]))
	}
}

func forwardHandler(w http.ResponseWriter, r *http.Request)  {
	data , err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error(err)
		w.Write([]byte(`{"success":false}`))
	}
	defer r.Body.Close()
	f := &Forward{}
	json.Unmarshal(data,f)
	req,err  := http.NewRequest(f.Method,f.Url,strings.NewReader(f.Body))
	req.Header = r.Header
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error(err)
		w.Write([]byte(`{"success":false}`))
	} else {
		body, _:= ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		w.WriteHeader(resp.StatusCode)
		w.Write(body)
	}
}
