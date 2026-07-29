package indexer

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (transport roundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return transport(request)
}

func testHTTPClient(handler func(*http.Request) (int, string)) *http.Client {
	return &http.Client{Transport: roundTripperFunc(func(request *http.Request) (*http.Response, error) {
		status, body := handler(request)
		return &http.Response{
			StatusCode: status,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
			Request:    request,
		}, nil
	})}
}

func TestSyncClientGetBalanceSupportsLargeAmounts(t *testing.T) {
	const largeAmount = "100000000000.00000001"
	httpClient := testHTTPClient(func(r *http.Request) (int, string) {
		if r.URL.Path != "/balance" {
			return http.StatusNotFound, ""
		}
		return http.StatusOK, fmt.Sprintf(`{"incoming":"0","available":%q,"outgoing":"0","current":%q}`, largeAmount, largeAmount)
	})

	client := NewSyncClient("http://indexer", httpClient)
	balance, err := client.GetBalance(context.Background(), "DExample")
	if err != nil {
		t.Fatalf("get balance: %v", err)
	}
	if balance.Available != largeAmount || balance.Current != largeAmount {
		t.Fatalf("unexpected balance: %+v", balance)
	}
}

func TestSyncClientRejectsInvalidOrOversizedBalance(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "numeric amount", body: `{"incoming":0,"available":"0","outgoing":"0","current":"0"}`},
		{name: "invalid amount", body: `{"incoming":"one","available":"0","outgoing":"0","current":"0"}`},
		{name: "oversized response", body: strings.Repeat(" ", maxIndexerResponseBytes+1)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			httpClient := testHTTPClient(func(_ *http.Request) (int, string) {
				return http.StatusOK, test.body
			})
			client := NewSyncClient("http://indexer", httpClient)
			if _, err := client.GetBalance(context.Background(), "DExample"); err == nil {
				t.Fatal("expected invalid response to fail")
			}
		})
	}
}

func TestSyncClientReturnsStatusError(t *testing.T) {
	httpClient := testHTTPClient(func(_ *http.Request) (int, string) {
		return http.StatusServiceUnavailable, "unavailable"
	})
	client := NewSyncClient("http://indexer", httpClient)
	_, err := client.GetBalance(context.Background(), "DExample")
	if err == nil || !strings.Contains(err.Error(), "status 503") {
		t.Fatalf("expected status error, got %v", err)
	}
}
