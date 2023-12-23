package main

import (
	"github.com/ganeshrockz/go-redis/client"
	"github.com/ganeshrockz/go-redis/server"
)

func main() {
	server := server.Server{
		SignalCh: make(chan struct{}),
	}
	go server.RunServer()

	<-server.SignalCh
	client.RunClient()
}
