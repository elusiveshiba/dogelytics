package main

import (
	"bytes"
	"io"
	"net/http"
	"strings"
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

func TestVersionCommand(t *testing.T) {
	var output bytes.Buffer
	if err := run([]string{"version"}, strings.NewReader(""), &output, &bytes.Buffer{}); err != nil {
		t.Fatalf("version command: %v", err)
	}
	if !strings.Contains(output.String(), "dogelytics dev") {
		t.Fatalf("unexpected version output: %q", output.String())
	}
}

func TestHealthcheckCommand(t *testing.T) {
	previousClient := healthcheckClient
	healthcheckClient = &http.Client{Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Body:       io.NopCloser(strings.NewReader("ready")),
			Header:     make(http.Header),
		}, nil
	})}
	t.Cleanup(func() { healthcheckClient = previousClient })
	var output bytes.Buffer
	if err := run([]string{"healthcheck", "--url", "http://dogelytics.test/readyz"}, strings.NewReader(""), &output, &bytes.Buffer{}); err != nil {
		t.Fatalf("healthcheck command: %v", err)
	}
	if output.String() != "ready\n" {
		t.Fatalf("unexpected healthcheck output: %q", output.String())
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func TestReadPasswordFromStandardInput(t *testing.T) {
	password, err := readPassword(strings.NewReader("correct horse battery staple\n"), &bytes.Buffer{}, true)
	if err != nil {
		t.Fatalf("read password: %v", err)
	}
	if password != "correct horse battery staple" {
		t.Fatalf("unexpected password value: %q", password)
	}
}
