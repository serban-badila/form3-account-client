//go:build integration || load
// +build integration load

package tests

import (
	"os"
)

const (
	hostAddressName = "HOST_ADDRESS"
	defaultHost     = "http://localhost:8080"
)

func fetchAPIHostName() string {
	if hostName, ok := os.LookupEnv(hostAddressName); ok {
		return hostName
	}
	return defaultHost
}
