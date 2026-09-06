package l7

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/iamonah/loadbalancer/config"
	"github.com/rs/zerolog/log"
)

type l7Server struct {
	// Finds the server pool by matcher.
	// Note(self): The matcher could be more sophisticated, such as regex-
	// or subdomain-based matching, but for now we use simple string matching.
	servicePools map[string]*BackendPool
	config       *config.Config
}

func NewLoadBalancer(cfg *config.Config) (*l7Server, error) {
	svcPools := make(map[string]*BackendPool)
	for _, service := range cfg.Services {
		pool, err := NewServerPool(service)
		if err != nil {
			return nil, fmt.Errorf("Failed to create server pool for service %s: %w", service.Name, err)
		}
		svcPools[service.Matcher] = pool
	}

	return &l7Server{
		config:       cfg,
		servicePools: svcPools,
	}, nil
}

func (lb *l7Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Info().Msgf("Received new request for %s", r.URL.String())

	pool, ok := lb.findPool(r.URL.RawPath)
	if !ok {
		http.NotFound(w, r)
		return
	}
	server := pool.GetNextBackend()
	log.Info().Msgf("Forwarding request to service %s at %s", pool.serviceName, server.url.String())
	server.forward(w, r)
}

// Note(self): Prefix matching is currently O(N) because we iterate over
// configured matchers. This is fine for now, but i will consider using a trie/radix
// tree if the number of routes grows significantly.
func (lb *l7Server) findPool(path string) (*BackendPool, bool) {
	for matcher, pool := range lb.servicePools {
		if path == matcher || strings.HasPrefix(path, matcher+"/") {
			return pool, true
		}
	}

	return nil, false
}
