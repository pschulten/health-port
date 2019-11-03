package main

import (
	server "github.com/pschulten/health-port"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	hps := server.HealthPortServer{
		Addr: ":1161",
	}
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-c
		log.Printf("Exiting on signal: %v", sig)
		os.Exit(1)
	}()

	running := false
	for {
		log.Println("** main begin loop")
		time.Sleep(time.Second)

		if check() {
			if !running {
				go hps.ListenAndServe()
			}
			running = true
		} else {
			if running {
				hps.Shutdown()
			}
			running = false
		}
		log.Println("** main end loop")
	}

}

func check() bool {
	res, err := http.Get("http://localhost:2015/index.html")
	if err != nil {
		return false
	}

	if err := res.Body.Close(); err != nil {
		log.Printf("Close response body failed???, ignoring: %v", err)
	}
	if res.StatusCode > 299 {
		return false
	}
	return true
}
