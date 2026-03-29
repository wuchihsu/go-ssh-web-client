package main

import (
	"fmt"
	"os"
)

func getPort(defaultPort uint) string {
	httpPort := os.Getenv("PORT")
	if httpPort == "" {
		httpPort = fmt.Sprintf("%d", defaultPort)
	}
	return httpPort
}
