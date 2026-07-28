package main

import (
	"net/http"
	"testing"
	"time"
)

func TestNewHTTPServerAppliesTimeouts(t *testing.T) {
	server := newHTTPServer("127.0.0.1:4420", http.NotFoundHandler())
	if server.ReadHeaderTimeout != 5*time.Second || server.ReadTimeout != 15*time.Second {
		t.Fatalf("unexpected read timeouts: header=%v request=%v", server.ReadHeaderTimeout, server.ReadTimeout)
	}
	if server.WriteTimeout != 30*time.Second || server.IdleTimeout != 60*time.Second {
		t.Fatalf("unexpected write/idle timeouts: write=%v idle=%v", server.WriteTimeout, server.IdleTimeout)
	}
}
