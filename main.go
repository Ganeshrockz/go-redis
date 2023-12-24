package main

import (
	"fmt"

	"github.com/ganeshrockz/go-redis/client"
	"github.com/ganeshrockz/go-redis/server"
)

func main() {
	server := server.New()
	go server.RunServer()

	<-server.SignalCh
	go func() {
		for err := range server.ErrCh {
			fmt.Println("error running server " + err.Error())
		}
	}()
	client.New().Run()
	close(server.ErrCh)
}
