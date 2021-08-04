package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/amarchese96/mentat/pkg/k8s"
	"github.com/amarchese96/mentat/pkg/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	sleep_seconds, err := strconv.Atoi(os.Getenv("SLEEP_SECONDS"))

	if err != nil || sleep_seconds <= 0 {
		log.Fatalf("SLEEP_SECONDS must be a positive integer")
	}

	hostname, err := k8s.GetNodeName()
	if err != nil {
		log.Fatalf("failed getting hostname: %s", err)
	}

	distributionConfig, err := utils.ParseNodeLatenciesDistributionConfig("/app/node_latencies_distribution.json")
	if err != nil {
		log.Fatalf("failed parsing node latencies distribution config file: %s", err)
	}

	histogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "node_latency",
		Help:    "Time take ping other nodes",
		Buckets: []float64{1, 2, 5, 6, 10},
	}, []string{"origin_node", "destination_node"})

	err = prometheus.Register(histogram)
	if err != nil {
		log.Fatalf("failed registering historgram: %s", err)
	}

	go func() {
		for {
			nodes, err := k8s.GetNodeList()
			if err != nil {
				log.Fatalf("failed getting node list: %s", err)
			}

			if len(nodes) == 0 {
				log.Fatal("getNodes returned 0 nodes")
			}

			for _, node := range nodes {
				if node.Hostname != hostname {
					rtt, err := utils.PingHost(node.Hostname, node.Ip)

					if err != nil {
						log.Printf("failed pinging node '%s' : %s", node.Hostname, err)
					} else {
						linkInfo := distributionConfig[hostname][node.Hostname]
						randomContribute := utils.GetValueFromDistribution(linkInfo.Distribution, linkInfo.Params)
						log.Printf("Time: %vs\n", rtt.Seconds()+randomContribute)
						histogram.WithLabelValues(hostname, node.Hostname).Observe(rtt.Seconds() + randomContribute)
					}
				}

			}

			time.Sleep(time.Duration(sleep_seconds) * time.Second)
		}

	}()

	http.Handle("/metrics", promhttp.Handler())
	_ = http.ListenAndServe(":2112", nil)
}
