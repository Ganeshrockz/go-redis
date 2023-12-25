package main

import (
	"fmt"
	"sync"

	"github.com/ganeshrockz/go-redis/client"
	"github.com/ganeshrockz/go-redis/server"
)

func main() {
	server := server.New()
	go server.RunServer()

	<-server.SignalCh
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		//defer wg.Done()
		for err := range server.ErrCh {
			fmt.Println("error running server " + err.Error())
		}
	}()

	// Run concurrent clients
	go func() {
		defer wg.Done()
		client.New().Run()
	}()
	go func() {
		defer wg.Done()
		client.New().Run()
	}()
	wg.Wait()
	close(server.ErrCh)
}
