package server

import (
	"testing"
	"net/http/httptest"
	"strings"
)

func TestNetworkCheckHandler(t *testing.T)  {
	read := strings.NewReader("{}")
	req := httptest.NewRequest("POST","http://localhost",read)
	networkCheckHandler(nil,req)
}

func TestStartPortHandler(t *testing.T)  {
	read := strings.NewReader(`{"ports": [8080,8081,9999]}`)
	req := httptest.NewRequest("POST","http://localhost",read)
	startPortHandler(nil,req)
}