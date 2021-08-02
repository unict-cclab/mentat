package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"os"
)

type DistributionInfo struct {
	Distribution string             `json:"distribution,omitempty"`
	Params       map[string]float64 `json:"params,omitempty"`
}

func parseNodeLatenciesDistributionConfig(configFilePath string) (map[string]map[string]DistributionInfo, error) {
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

func getValueFromDistribution(distribution string, params map[string]float64) float64 {
	switch distribution {
	case "gaussian":
		return math.Abs(rand.NormFloat64()*params["std"] + params["mean"])
	default:
		return 0.0
	}
}
