package server

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNetworkCheckHandler(t *testing.T) {
	read := strings.NewReader(`{"ipList": ["192.168.99.100","192.168.99.101","192.168.99.102"]}`)
	req := httptest.NewRequest("POST", "http://localhost:9000", read)
	networkCheckHandler(nil, req)
}

func TestStartPortHandler(t *testing.T) {
	read := strings.NewReader(`{"ports": [8080,8081,9999]}`)
	req := httptest.NewRequest("POST", "http://localhost", read)
	startPortHandler(nil, req)
}
