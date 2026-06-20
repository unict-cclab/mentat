package utils

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestPacketLossRatio(t *testing.T) {
	result := PingResult{Sent: 5, Received: 3}
	if got, want := result.PacketLossRatio(), 0.4; got != want {
		t.Fatalf("PacketLossRatio() = %v, want %v", got, want)
	}
}

func TestBandwidthHandlerAndMeasurement(t *testing.T) {
	const payloadBytes = 1024 * 1024
	request := httptest.NewRequest(http.MethodGet, bandwidthPath, nil)
	recorder := httptest.NewRecorder()
	BandwidthHandler(payloadBytes).ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("BandwidthHandler() status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if got := recorder.Body.Len(); got != payloadBytes {
		t.Fatalf("BandwidthHandler() bytes = %d, want %d", got, payloadBytes)
	}

	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Body:       io.NopCloser(strings.NewReader("probe payload")),
			Header:     make(http.Header),
		}, nil
	})}
	bytesPerSecond, err := measureBandwidth(client, "http://peer/bandwidth")
	if err != nil {
		t.Fatalf("MeasureBandwidth() error = %v", err)
	}
	if bytesPerSecond <= 0 {
		t.Fatalf("MeasureBandwidth() = %v, want positive throughput", bytesPerSecond)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

func TestProbeHostRejectsInvalidInput(t *testing.T) {
	if _, err := ProbeHost("not-an-ip", 3, time.Second); err == nil {
		t.Fatal("ProbeHost() accepted an invalid IP")
	}
	if _, err := ProbeHost("127.0.0.1", 0, time.Second); err == nil {
		t.Fatal("ProbeHost() accepted zero attempts")
	}
}
