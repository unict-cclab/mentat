package main

import (
	"context"
	"log"
	"net"
	"os"
	"time"

	ping "github.com/digineo/go-ping"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Node struct {
	Hostname string
	Ip       string
}

func getNodeList() ([]Node, error) {

	var simpleList []Node

	// Get configuration from within the cluster itself
	config, err := rest.InClusterConfig()
	if err != nil {
		return simpleList, err
	}

	c, err := kubernetes.NewForConfig(config)
	if err != nil {
		return simpleList, err
	}

	nodes, err := c.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{
		Limit: 500,
	})

	if err != nil {
		return simpleList, err
	}

	for _, item := range nodes.Items {
		node := Node{
			Hostname: item.Name,
		}

		for _, nodeAddress := range item.Status.Addresses {
			if nodeAddress.Type == "InternalIP" {
				node.Ip = nodeAddress.Address
				break
			}
		}

		simpleList = append(simpleList, node)
	}

	return simpleList, nil
}

func getNodeName() (string, error) {
	name := os.Getenv("NODE_NAME")

	var err error

	// This method is unreliable since it will return just the pod name.
	if name == "" {
		name, err = os.Hostname()
	}

	return name, err

}

func pingHost(destinationHost, destinationIp string) (time.Duration, error) {
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

	log.Printf("ping %s (%s) rtt=%v\n", destinationHost, remoteAddr, rtt)

	return rtt, nil

}
