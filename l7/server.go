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
	// this could later come from an external config file or service discovery mechanism
	config *config.Config
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

	pool, ok := lb.findPool(r.URL.Path)
	if !ok {
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		log.Warn().Msgf("No matching service pool found for request path: %s", r.URL.Path)
		return
	}

	server := pool.getNextBackend()
	log.Info().Msgf(
		"L7 Proxy routing request: path: %s host: %s scheme: %s",
		r.URL.Path,
		r.URL.Host,
		r.URL.Scheme,
	)
	server.forward(w, r)
}

// Note(self): Prefix matching is currently O(N) because we iterate over
// configured matchers. This is fine for now, but i will consider using a trie/radix
// tree if the number of routes grows significantly.
func (lb *l7Server) findPool(reqPath string) (*BackendPool, bool) {
	log.Info().Msgf("Finding pool for request path: %s", reqPath)
	for matcher, pool := range lb.servicePools {
		if reqPath == matcher || strings.HasPrefix(reqPath, matcher+"/") {
			log.Info().Msgf("Matched request path %s to service %s", reqPath, pool.serviceName)
			return pool, true
		}
	}

	return nil, false
}
