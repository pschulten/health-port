package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

type HealthPortServer struct {
	Addr     string
	listener net.Listener
	channel  chan struct{}
}

func (hps *HealthPortServer) ListenAndServe() error {
	var server http.Server
	addr := hps.Addr
	if addr == "" {
		addr = ":1161"
	}
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("HTTP server Listen: %v", err)
	}
	hps.listener = listener
	stop := make(chan os.Signal)
	signal.Notify(stop, syscall.SIGINT)

	quit := make(chan struct{})
	hps.channel = quit

	for {
		go func() {
			defer func() {
				log.Println("shutting down server...")
				if err := server.Shutdown(context.TODO()); err != nil {
					log.Printf("HTTP server shutdown, ignoring: %v", err)
				}
			}()

			fmt.Printf("Serving on %s \n", hps.Addr)
			if err := server.Serve(listener); err != nil {
				log.Printf("HTTP server serve, ignoring: %v", err)
			}
		}()

		select {
		case <-quit:
			fmt.Println("Got quit")
			return nil
		}
	}
	return nil
}

func (hps *HealthPortServer) Shutdown() {
	close(hps.channel)
	log.Println("shutting down...")
	err := hps.listener.Close()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("closed...")
}
