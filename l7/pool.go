package l7

import (
	"fmt"
	"net/http/httputil"
	"net/url"

	"github.com/iamonah/loadbalancer/config"
	"github.com/iamonah/loadbalancer/strategy"
)

type Server struct {
	url   *url.URL
	proxy *httputil.ReverseProxy
}

type ServerPool struct {
	serviceName string
	servers     []*Server
	Strategy    strategy.Strategy
}

func NewServerPool(svcCfg *config.Service) (*ServerPool, error) {
	servers := make([]*Server, 0, len(svcCfg.Replicas))

	if len(svcCfg.Replicas) == 0 {
		return nil, fmt.Errorf("No replicas defined for service %s", svcCfg.Name)
	}
	for _, replica := range svcCfg.Replicas {
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

	strategy, err := strategy.NewStrategy(
		strategy.StrategyConfig{Type: strategy.StrategyType(svcCfg.Strategy), Weights: svcCfg.Weights},
		uint32(len(svcCfg.Replicas)),
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to create strategy: %w", err)
	}
	return &ServerPool{serviceName: svcCfg.Name, servers: servers, Strategy: strategy}, nil
}

func (sp *ServerPool) GetNextServer() *Server {
	nextIndex := sp.Strategy.NextServer()
	return sp.servers[nextIndex]
}
