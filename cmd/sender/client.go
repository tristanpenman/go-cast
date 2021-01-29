package main

import (
	"crypto/tls"
	"fmt"
)

func startClient(hostname *string, port *uint) {
	addr := fmt.Sprintf("%s:%d", *hostname, *port)
	logger.Info(fmt.Sprintf("addr: %s", addr))

	config := tls.Config{InsecureSkipVerify: true}
	conn, err := tls.Dial("tcp", addr, &config)
	if err != nil {
		logger.Error("client: dial error", "err", err)
		return
	}

	defer conn.Close()
}
