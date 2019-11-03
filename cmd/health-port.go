package cmd

import (
	"fmt"
	server2 "github.com/pschulten/health-port/server"
	"github.com/spf13/cobra"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var rootCmd = &cobra.Command{
	Use:   "health-port",
	Short: "health-port - expose semantic health check",
	Long:  `health-port - expose semantic health check by running a dummy server on healthiness`,
	Run:   help,
}

func help(cmd *cobra.Command, args []string) {
	cmd.Help()
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	hps := server2.HealthPortServer{
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
