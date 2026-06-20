package utils

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/digineo/go-ping"
)

const bandwidthPath = "/bandwidth"

type PingResult struct {
	AverageRTT time.Duration
	Sent       int
	Received   int
}

func (result PingResult) PacketLossRatio() float64 {
	if result.Sent == 0 {
		return 0
	}
	return float64(result.Sent-result.Received) / float64(result.Sent)
}

// ProbeHost sends individual ICMP requests so both RTT and packet loss can be
// measured. A timed-out request counts as a lost packet; failure to create the
// ICMP socket is returned as an operational error.
func ProbeHost(destinationIP string, attempts int, timeout time.Duration) (PingResult, error) {
	result := PingResult{Sent: attempts}
	if attempts <= 0 {
		return PingResult{}, fmt.Errorf("attempts must be positive")
	}

	remoteIP := net.ParseIP(destinationIP)
	if remoteIP == nil {
		return PingResult{}, fmt.Errorf("invalid destination IP %q", destinationIP)
	}

	pinger, err := ping.New("0.0.0.0", "")
	if err != nil {
		return PingResult{}, err
	}
	defer pinger.Close()

	remoteAddr := &net.IPAddr{IP: remoteIP}
	var totalRTT time.Duration
	for i := 0; i < attempts; i++ {
		rtt, pingErr := pinger.Ping(remoteAddr, timeout)
		if pingErr != nil {
			continue
		}
		result.Received++
		totalRTT += rtt
	}

	if result.Received > 0 {
		result.AverageRTT = totalRTT / time.Duration(result.Received)
	}
	return result, nil
}

// BandwidthHandler returns a fixed-size response used by peer Mentat agents to
// measure effective TCP throughput.
func BandwidthHandler(payloadBytes int64) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path != bandwidthPath {
			http.NotFound(response, request)
			return
		}
		response.Header().Set("Content-Type", "application/octet-stream")
		response.Header().Set("Content-Length", fmt.Sprintf("%d", payloadBytes))

		chunk := make([]byte, 32*1024)
		remaining := payloadBytes
		for remaining > 0 {
			next := int64(len(chunk))
			if remaining < next {
				next = remaining
			}
			written, err := response.Write(chunk[:next])
			if err != nil {
				return
			}
			remaining -= int64(written)
		}
	})
}

// MeasureBandwidth downloads the peer's probe payload and returns effective
// application throughput in bytes per second.
func MeasureBandwidth(address string, timeout time.Duration) (float64, error) {
	client := &http.Client{Timeout: timeout}
	return measureBandwidth(client, "http://"+address+bandwidthPath)
}

func measureBandwidth(client *http.Client, endpoint string) (float64, error) {
	started := time.Now()
	response, err := client.Get(endpoint)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("bandwidth probe returned HTTP %s", response.Status)
	}

	bytesRead, err := io.Copy(io.Discard, response.Body)
	if err != nil {
		return 0, err
	}
	elapsed := time.Since(started)
	if bytesRead == 0 || elapsed <= 0 {
		return 0, fmt.Errorf("bandwidth probe returned no data")
	}
	return float64(bytesRead) / elapsed.Seconds(), nil
}
