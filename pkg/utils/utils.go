package utils

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net"
	"os"
	"time"

	ping "github.com/digineo/go-ping"
)

type DistributionInfo struct {
	Distribution string             `json:"distribution,omitempty"`
	Params       map[string]float64 `json:"params,omitempty"`
}

func ParseNodeLatenciesDistributionConfig(configFilePath string) (map[string]map[string]DistributionInfo, error) {
	jsonFile, err := os.Open(configFilePath)

	if err != nil {
		return nil, err
	}

	log.Printf("Successfully opened %s", configFilePath)

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var result map[string]map[string]DistributionInfo
	json.Unmarshal([]byte(byteValue), &result)

	return result, nil

}

func GetValueFromDistribution(distribution string, params map[string]float64) float64 {
	switch distribution {
	case "gaussian":
		return math.Abs(rand.NormFloat64()*params["std"] + params["mean"])
	default:
		return 0.0
	}
}

func PingHost(destinationHost, destinationIp string) (time.Duration, error) {
	bind := "0.0.0.0"

	var remoteAddr *net.IPAddr
	var pinger *ping.Pinger
	var rtt time.Duration

	remoteAddr = &net.IPAddr{
		IP: net.ParseIP(destinationIp),
	}

	pinger, err := ping.New(bind, "")
	if err != nil {
		return rtt, err
	}
	defer pinger.Close()

	timeout, _ := time.ParseDuration("30s")
	rtt, err = pinger.PingAttempts(remoteAddr, timeout, int(3))
	if err != nil {
		return rtt, err
	}

	return rtt, nil
}
