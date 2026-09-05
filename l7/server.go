package l7

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/iamonah/loadbalancer/config"
	"github.com/rs/zerolog/log"
)

type LoadBalancer struct {
	servicePools map[string]*ServerPool
	config       *config.Config
}

func NewLoadBalancer(cfg *config.Config) (*LoadBalancer, error) {
	svcPools := make(map[string]*ServerPool)
	for _, service := range cfg.Services {
		pool, err := NewServerPool(service)
		if err != nil {
			return nil, fmt.Errorf("Failed to create server pool for service %s: %w", service.Name, err)
		}
		svcPools[service.Name] = pool
	}

	return &LoadBalancer{
		config:       cfg,
		servicePools: svcPools,
	}, nil
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Info().Msgf("Received new request for %s", r.URL.String())
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	//service name from urls
	serviceName := parts[0]
	if serviceName == "" {
		serviceName = lb.config.DefaultService
	}

	pool, ok := lb.servicePools[serviceName]
	if !ok {
		http.NotFound(w, r)
		return
	}

	server := pool.GetNextServer()
	log.Info().Msgf("Forwarding request to %s", server.url.String())
	server.proxy.ServeHTTP(w, r)
}
