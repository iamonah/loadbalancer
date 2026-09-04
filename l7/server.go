package l7

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/iamonah/loadbalancer/config"
	"github.com/iamonah/loadbalancer/l7/strategy"
	"github.com/rs/zerolog/log"
)

type LoadBalancer struct {
	Services map[string]*ServerPool
	Config   *config.Config
}

func NewLoadBalancer(cfg *config.Config) (*LoadBalancer, error) {
	pool := make(map[string]*ServerPool)

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
		var scheduler strategy.Strategy

		switch cfg.Services[service.Name].Strategy {
		case "round-robin":
			scheduler = strategy.NewRoundRobin(uint32(len(cfg.Services[service.Name].Replicas)))
		// case "weighted-round-robin":
		// 	scheduler = strategy.NewWeightedRoundRobin( )
		default:
			return nil, fmt.Errorf("unsupported strategy: %s", cfg.Services[service.Name].Strategy)
		}
		pool[service.Name] = &ServerPool{servers: servers, Strategy: scheduler}
	}

	return &LoadBalancer{
		Config:   cfg,
		Services: pool,
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

	next := pool.Strategy.NextServer()
	server := pool.servers[next]
	log.Info().Msgf("Forwarding request to %s", server.url.String())
	server.proxy.ServeHTTP(w, r)
}
