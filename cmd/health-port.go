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

var (
	addr     string
	interval time.Duration
	endpoint string
)

func init() {
	rootCmd.PersistentFlags().StringVar(&addr, "addr", ":1161", "Port to expose")
	rootCmd.PersistentFlags().DurationVar(&interval, "interval", 2*time.Second, "The amount of time between health checks")
}

var rootCmd = &cobra.Command{
	Use:   "health-port health_check_url",
	Short: "health-port - expose semantic health check",
	Long:  `health-port - expose semantic health check by running a dummy server on healthiness`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		endpoint = args[0]
		main()
	},
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
		Addr: addr,
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

		time.Sleep(interval)

		if check() {
			if !running {
				log.Printf("healthy endpoint: %s\n", endpoint)
				go hps.ListenAndServe()
			}
			running = true
		} else {
			if running {
				hps.Shutdown()
			}
			running = false
		}
	}

}

func check() bool {
	//res, err := http.Get("http://localhost:2015/index.html")
	res, err := http.Get(endpoint)
	if err != nil {
		log.Printf("%s down: %v\n", endpoint, err)
		return false
	}

	if err := res.Body.Close(); err != nil {
		log.Printf("Close response body failed???, ignoring: %v", err)
	}
	if res.StatusCode > 299 {
		log.Printf("%s down: status code: %d \n", endpoint, res.StatusCode)
		return false
	}
	return true
}
