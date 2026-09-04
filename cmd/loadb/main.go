package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	port = flag.Int("port", 8080, "listening port")
)

type Service struct {
	Name     string
	Replicas []string
}

type Config struct {
	Services       []Service
	Strategy       string
	DefaultService string
}

type Server struct {
	url   *url.URL
	proxy *httputil.ReverseProxy
}

type ServerPool struct {
	servers []*Server
	// current server (circular approach)
	current atomic.Uint32
}

func (sp *ServerPool) NextServer() uint32 {
	length := uint32(len(sp.servers))
	for {
		current := sp.current.Load()
		next := current + 1

		if next >= length {
			next = 0
		}

		if sp.current.CompareAndSwap(current, next) {
			return next
		}
	}
}

type LoadBalancer struct {
	Services map[string]*ServerPool
	Config   *Config
}

func NewLoadBalancer(cfg *Config) (*LoadBalancer, error) {
	services := make(map[string]*ServerPool)

	for _, service := range cfg.Services {
		servers := make([]*Server, 0)

		for _, replica := range service.Replicas {
			parsedURL, err := url.Parse(replica)
			if err != nil {
				return nil, fmt.Errorf("Failed to parse URL: %w", err)
			}
			server := &Server{
				url:   parsedURL,
				proxy: httputil.NewSingleHostReverseProxy(parsedURL),
			}
			servers = append(servers, server)
		}
		services[service.Name] = &ServerPool{servers: servers}
	}
	return &LoadBalancer{
		Config:   cfg,
		Services: services,
	}, nil
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// and forward to the appropriate service
	log.Info().Msgf("Received new request for %s", r.URL.String())
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	serviceName := parts[0]
	if serviceName == "" {
		serviceName = lb.Config.DefaultService
	}

	pool, ok := lb.Services[serviceName]
	if !ok {
		http.NotFound(w, r)
		return
	}

	next := pool.NextServer()
	server := pool.servers[next]
	log.Info().Msgf("Forwarding request to %s", server.url.String())
	server.proxy.ServeHTTP(w, r)
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	flag.Parse()

	var cfg *Config = &Config{
		Services: []Service{
			{
				Name:     "TestService",
				Replicas: []string{"http://localhost:8081", "http://localhost:8082", "http://localhost:8083"},
			},
		},
		Strategy:       "round-robin",
		DefaultService: "TestService",
	}
	lb, err := NewLoadBalancer(cfg)
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
