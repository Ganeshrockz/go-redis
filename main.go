package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/ganeshrockz/go-redis/server"
)

func main() {
	var sigs chan os.Signal = make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
	var wg sync.WaitGroup
	wg.Add(1)

	server := server.New(&wg, sigs)
	go server.RunServer()
	go func() {
		for err := range server.ErrCh {
			log.Printf("error running server " + err.Error())
		}
	}()

	wg.Wait()
	close(server.ErrCh)
}
