package main

import (
	"flag"
	"net/http"
	"strconv"

	"github.com/iamonah/loadbalancer/config"
	"github.com/iamonah/loadbalancer/l7"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	port = flag.Int("port", 8080, "listening port")
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	flag.Parse()

	cfg := config.LoadConfig()
	lb, err := l7.NewLoadBalancer(cfg)
	if err != nil {
		log.Error().Msg("Failed to create load balancer: " + err.Error())
		return
	}
	server := http.Server{
		Addr:    ":" + strconv.Itoa(*port),
		Handler: lb,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Error().Msg("Failed to start server: " + err.Error())
	}
}
